package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"pawked.com/truenas-gotify-adapter/metrics"
)

// get necessary environment variables
var (
	gotifyURL     = os.Getenv("GOTIFY_URL")
	gotifyToken   = os.Getenv("GOTIFY_TOKEN")
	listenHost    = os.Getenv("LISTEN_HOST")
	listenPort    = os.Getenv("LISTEN_PORT")
	enableMetrics = os.Getenv("PROMETHEUS_METRICS") == "1" // set to true if the env is set to 1
)

// set up some structs to make things simpler

// incoming TrueNAS request
type Request struct {
	Text string `json:"text"`
}

// payload format to send to Gotify
type GotifyPayload struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	// set up the router "r"
	r := gin.Default()
	r.SetTrustedProxies(nil) // don't trust all proxies like default.  i think this is fine as it is, but I could add handling for trusted proxies if necessary later
	// listen to post requests on / and /message
	r.POST("/", onMessageHandler)
	r.POST("/message", onMessageHandler)

	// turn on prometheus metrics if enabled in env
	if enableMetrics {
		metrics.Register()
		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
		log.Printf("Prometheus metrics will be served on /metrics\n")
	} else {
		log.Printf("Prometheus metrics are disabled")
	}

	// build listen address
	listenAddress := fmt.Sprintf("%s:%s", listenHost, listenPort)

	// start listening
	log.Printf("Listening on %s...\n", listenAddress)
	if err := r.Run(listenAddress); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// Gin handler for both routes
func onMessageHandler(c *gin.Context) {

	// get the total time to handle a message (even on fail)
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.RequestDuration.Observe(duration)
	}()

	//increment the number of messages received
	metrics.RequestsTotal.Inc()

	// read the content of the alert
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("Error: Couldn't read request body:", err)
		c.Status(http.StatusBadRequest) // return error to TrueNAS
		metrics.RequestsFailedTotal.Inc()
		return
	}

	// unmarshal alert body into "request" and check that request.Text exists
	var request Request
	if err := json.Unmarshal(body, &request); err != nil || request.Text == "" { // check if error or more importantly, missing text field
		log.Println("Error: Request has invalid JSON or missing 'text' field:", err)
		c.Status(http.StatusBadRequest) // also return 400 on error or missing text field
		metrics.RequestsFailedTotal.Inc()
		return
	}

	// print entire bdy to log (for testing at the monent, I'm looking for the severity)
	log.Printf("Received payload:\n\n%s", string(body))

	// extract notification title and message from alert
	lines := strings.Split(request.Text, "\n")                  // split Text at newline
	title := strings.TrimSpace(lines[0])                        // title is string 0 without whitespace
	message := strings.TrimSpace(strings.Join(lines[1:], "\n")) // message is from string 1 without whitespace

	// print title and message to console -- This could actually just go to a log instead but i'll do that later
	fmt.Printf("========== %s ==========\n", title)
	fmt.Printf("%s\n", message)
	fmt.Println(strings.Repeat("=", len(title)) + "======================") // this is just pulled from the original script and I like how it looks

	// prepare Gotify payload
	payload := GotifyPayload{
		Title:   title,
		Message: message,
	}

	// Forward the alert to Gotify
	resp, err := sendGotifyMessage(payload)
	if err != nil {
		log.Println("Error forwarding to Gotify:", err)
		c.Status(http.StatusInternalServerError) // return 500 to TrueNAS on error
		metrics.GotifySendsFailedTotal.Inc()     // increment failed counter
		return
	}

	// Check for http reponse status code 'success'
	switch resp.StatusCode {
	case http.StatusOK: // success!
		log.Println(">> Forwarded successfully")
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden: // bad token?
		log.Printf(">> Unauthorized! GOTIFY_TOKEN is incorrect. Error Code: %d\n", resp.StatusCode)
	default: // something else?
		log.Printf(">> Unknown error while forwarding to gotify. Error Code: %d\n", resp.StatusCode)
	}
	// sets the gotify status code for truenas
	c.Status(resp.StatusCode)
}

// Forwards GotifyPayload to Gotify
func sendGotifyMessage(payload GotifyPayload) (*http.Response, error) {

	metrics.GotifySendsTotal.Inc() // increment sends counter

	start := time.Now() // start time for request

	// prepare io.Reader body for http.NewRequest
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// create new POST request "req" with payloadBytes
	body := bytes.NewBuffer(payloadBytes)
	req, err := http.NewRequest("POST", gotifyURL, body)
	if err != nil {
		return nil, err
	}

	// set request headers
	req.Header.Set("Content-Type", "application/json") // necessary - got help here https://stackoverflow.com/questions/45426137/golang-struct-as-payload-for-post-request
	req.Header.Set("X-Gotify-Key", gotifyToken)        // token

	// send request and return response
	resp, err := http.DefaultClient.Do(req)

	// get time for response
	duration := time.Since(start).Seconds()      // time since the start of the send
	metrics.GotifySendDuration.Observe(duration) // Observe records it into the histogram
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // https://pkg.go.dev/net/http close those bodies

	return resp, nil
}
