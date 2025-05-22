FROM golang:1.24.3-alpine AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /usr/local/bin/app/main .

FROM alpine:latest
COPY --from=builder /usr/local/bin/app/main /usr/bin/main
ENV PATH="/usr/bin:${PATH}"
CMD ["main"]