# Chirpy Server

A lightweight HTTP server written in Go to serve static files and track usage metrics.

## Overview

Chirpy Server is a basic HTTP server that provides:
- Static file serving from the `/app` endpoint
- Health check endpoint
- Request metrics tracking
- Counter reset functionality

## Features

### Static File Serving

The server serves static files from the current directory through the `/app` endpoint. Each request to this endpoint is counted for metrics purposes.

### API Endpoints

| Endpoint | Method | Description |
| --- | --- | --- |
| `/healthz` | GET | Health check endpoint that returns "OK" when the server is running |
| `/metrics` | GET | Returns the number of hits to the static file server |
| `/reset` | GET | Resets the hit counter to zero |
| `/app/*` | GET | Serves static files from the root directory |

### Metrics Tracking

The server tracks the number of requests made to the static file server using atomic counters for thread safety. The metrics can be viewed at the `/metrics` endpoint and reset at the `/reset` endpoint.

## Technical Details

### Architecture

- Built with Go's standard `net/http` package
- Uses atomic operations for thread-safe counter incrementation
- Implements middleware pattern for request counting
- Uses `http.StripPrefix` to handle static file serving properly

### Code Structure

- `apiConfig`: Stores server state (hit counter)
- `middlewareMetricsInc`: Middleware that increments the hit counter
- `metricsHandler`: Displays current hit count
- `resetHandler`: Resets the hit counter
- `readinessHandler`: Simple health check endpoint

## Running the Server

To start the server:

```bash
go run .
```

The server will start on port 8080 by default. You can access it at [http://localhost:8080](http://localhost:8080).

## Development

The server uses Go modules for dependency management. The project structure is simple with all functionality in the main package. 