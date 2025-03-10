package middlewares

import (
	"github.com/rshafikov/alertme/internal/server/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type (
	respData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		respData *respData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.respData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.respData.status = statusCode
}

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
