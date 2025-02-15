package middlewares

import (
	"github.com/rshafikov/alertme/internal/server/config"
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

func LoggingMiddleware(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
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

		config.Log.Infof(
			`- "%s %s" %d %s %s %d`,
			r.Method, r.RequestURI, rData.status, http.StatusText(rData.status), duration.String(), rData.size,
		)
	}
	return http.HandlerFunc(logFn)
}
