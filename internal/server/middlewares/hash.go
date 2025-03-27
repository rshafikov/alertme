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

type hashWriter struct {
	w       http.ResponseWriter
	buffer  *bytes.Buffer
	hasher  hash.Hash
	status  int
	written bool
}

func newHashWriter(w http.ResponseWriter) *hashWriter {
	return &hashWriter{
		w:      w,
		buffer: &bytes.Buffer{},
		hasher: hmac.New(sha256.New, []byte(settings.CONF.Key)),
		status: 0,
	}
}

func (h *hashWriter) Write(p []byte) (int, error) {
	h.written = true
	h.hasher.Write(p)
	return h.buffer.Write(p)
}

func (h *hashWriter) Header() http.Header {
	return h.w.Header()
}

func (h *hashWriter) WriteHeader(status int) {
	h.status = status
}

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

func hashData(data []byte) []byte {
	h := hmac.New(sha256.New, []byte(settings.CONF.Key))
	h.Write(data)
	return h.Sum(nil)
}
