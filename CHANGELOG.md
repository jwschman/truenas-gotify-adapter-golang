# Changelog

## [1.0.0] - 2026-03-10

### Added
- `VERSION` file as single source of truth for versioning
- CI now tags Docker images with both the version number and `latest`
- `GET /health` endpoint returning 200 OK

### Fixed
- Metrics counters and sends were being called even when `PROMETHEUS_METRICS` was disabled
- `GotifySendsFailedTotal` was not incremented when Gotify returned a 4xx/5xx HTTP status (only incremented on connection errors)
- Title/message separator block was using `fmt.Printf` inconsistently with the rest of the app which uses `log`

### Removed
- Request duration and Gotify send duration histograms (`gotify_forwarder_gotify_sends_duration_seconds`, `gotify_forwarder_send_duration_seconds`) — not useful at low alert volumes

## [Pre-1.0.0]

- Initial implementation: receive TrueNAS Slack-format webhook, parse and forward to Gotify
- Prometheus metrics support (`PROMETHEUS_METRICS=1`)
- Debug mode (`DEBUG_MODE=1`)
- Trim stale "Current alerts" and "The following alert has been cleared" history from messages
