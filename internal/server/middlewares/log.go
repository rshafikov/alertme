package middlewares

import (
	"github.com/rshafikov/alertme/internal/server/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type (
	// respData stores information about the HTTP response for logging.
	respData struct {
		status int // HTTP status code
		size   int // response body size in bytes
	}

	// loggingResponseWriter wraps an http.ResponseWriter to capture response data for logging.
	loggingResponseWriter struct {
		http.ResponseWriter
		respData *respData // pointer to the response data to be captured
	}
)

// Write implements the http.ResponseWriter interface.
// It writes the data to the underlying ResponseWriter and tracks the size.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.respData.size += size
	return size, err
}

// WriteHeader implements the http.ResponseWriter interface.
// It writes the status code to the underlying ResponseWriter and captures it.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.respData.status = statusCode
}

// Logger is a middleware that logs information about HTTP requests and responses.
// It captures the request method, URI, response status, duration, and response size.
// The log entry is written at the Info level using the application's logger.
func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rData := &respData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			respData:       rData,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logger.Log.Info("-",
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.Int("status", rData.status),
			zap.String("status_text", http.StatusText(rData.status)),
			zap.String("duration", duration.String()),
			zap.Int("size", rData.size),
		)
	})
}
