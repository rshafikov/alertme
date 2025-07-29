// Package middlewares provides HTTP middleware functions for the server.
package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/settings"
	"go.uber.org/zap"
	"hash"
	"io"
	"net/http"
)

// hashWriter is a wrapper around http.ResponseWriter that buffers the response
// and calculates a hash of the response body before writing it.
type hashWriter struct {
	w       http.ResponseWriter // original response writer
	buffer  *bytes.Buffer       // buffer to hold the response body
	hasher  hash.Hash           // hasher for calculating the HMAC SHA-256 hash
	status  int                 // HTTP status code to write
	written bool                // whether any data has been written
}

// newHashWriter creates a new hashWriter that wraps the given http.ResponseWriter.
func newHashWriter(w http.ResponseWriter) *hashWriter {
	return &hashWriter{
		w:      w,
		buffer: &bytes.Buffer{},
		hasher: hmac.New(sha256.New, []byte(settings.CONF.Key)),
		status: 0,
	}
}

// Write implements the http.ResponseWriter interface.
// It writes the data to the buffer and updates the hash.
func (h *hashWriter) Write(p []byte) (int, error) {
	h.written = true
	h.hasher.Write(p)
	return h.buffer.Write(p)
}

// Header implements the http.ResponseWriter interface.
// It returns the header map from the wrapped response writer.
func (h *hashWriter) Header() http.Header {
	return h.w.Header()
}

// WriteHeader implements the http.ResponseWriter interface.
// It stores the status code to be written later during flush.
func (h *hashWriter) WriteHeader(status int) {
	h.status = status
}

// flush writes the buffered data to the original response writer.
// It calculates the hash of the response body and adds it as a header.
// It also sets the status code if one was specified.
func (h *hashWriter) flush() {
	if !h.written {
		return
	}

	if h.buffer.Len() > 0 {
		reqHash := hex.EncodeToString(h.hasher.Sum(nil))
		h.w.Header().Set("HashSHA256", reqHash)
	}

	if h.status == 0 {
		h.status = http.StatusOK
	}
	h.w.WriteHeader(h.status)

	h.w.Write(h.buffer.Bytes())
}

// Hasher is a middleware that verifies and generates HMAC SHA-256 hashes for request and response bodies.
// It checks the "HashSHA256" header of incoming requests against a computed hash of the request body.
// It also adds a "HashSHA256" header to responses containing the hash of the response body.
// If the configured key is empty, the middleware is effectively disabled.
// If the hash verification fails, it returns a 400 Bad Request error.
func Hasher(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if settings.CONF.Key == "" {
			next.ServeHTTP(w, r)
			return
		}

		if receivedHashStr := r.Header.Get("HashSHA256"); receivedHashStr != "" {
			receivedHash, err := hex.DecodeString(receivedHashStr)
			if err != nil {
				logger.Log.Debug("invalid or missing hash", zap.Error(err))
				http.Error(w, "invalid Hash", http.StatusBadRequest)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))

			expectedReqHash := hashData(body)
			if !hmac.Equal(receivedHash, expectedReqHash) {
				logger.Log.Debug("hash mismatch")
				http.Error(w, "hash mismatch", http.StatusBadRequest)
				return
			}
		}

		hw := newHashWriter(w)
		next.ServeHTTP(hw, r)
		hw.flush()
	}
	return http.HandlerFunc(fn)
}

// hashData calculates the HMAC SHA-256 hash of the given data using the configured key.
// It returns the raw hash bytes (not hex-encoded).
func hashData(data []byte) []byte {
	h := hmac.New(sha256.New, []byte(settings.CONF.Key))
	h.Write(data)
	return h.Sum(nil)
}
