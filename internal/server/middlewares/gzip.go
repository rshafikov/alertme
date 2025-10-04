package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// gzipWriterPool is a pool of gzip.Writer objects for reuse.
// This reduces the overhead of creating new writers for each request.
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

// compressWriter is a wrapper around http.ResponseWriter that compresses the response with gzip.
type compressWriter struct {
	w  http.ResponseWriter // original response writer
	zw *gzip.Writer        // gzip writer for compression
}

// newCompressWriter creates a new compressWriter that wraps the given http.ResponseWriter.
// It gets a gzip.Writer from the pool and resets it to write to the response writer.
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	zw := gzipWriterPool.Get().(*gzip.Writer)
	zw.Reset(w)
	return &compressWriter{
		w:  w,
		zw: zw,
	}
}

// Header implements the http.ResponseWriter interface.
// It returns the header map from the wrapped response writer.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write implements the http.ResponseWriter interface.
// It writes the compressed data to the gzip writer.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader implements the http.ResponseWriter interface.
// It sets the Content-Encoding header to gzip for successful responses
// and writes the status code to the wrapped response writer.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close closes the gzip writer and returns it to the pool.
// This should be called when the response is complete.
func (c *compressWriter) Close() error {
	defer gzipWriterPool.Put(c.zw)
	return c.zw.Close()
}

// compressReader is a wrapper around io.ReadCloser that decompresses gzip-encoded data.
type compressReader struct {
	r  io.ReadCloser // original reader
	zr *gzip.Reader  // gzip reader for decompression
}

// newCompressReader creates a new compressReader that wraps the given io.ReadCloser.
// It creates a new gzip.Reader to decompress the data from the original reader.
// Returns an error if the gzip reader cannot be created (e.g., if the data is not valid gzip).
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read implements the io.Reader interface.
// It reads decompressed data from the gzip reader.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close implements the io.Closer interface.
// It closes both the original reader and the gzip reader.
// If closing the original reader fails, the error is returned immediately.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// GZipper is a middleware that handles gzip compression and decompression.
// It compresses responses if the client accepts gzip encoding (via the Accept-Encoding header).
// It also decompresses request bodies if they are gzip-encoded (via the Content-Encoding header).
// This middleware helps reduce bandwidth usage for compatible clients.
func GZipper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		responseWithGzip := strings.Contains(acceptEncoding, "gzip")
		if responseWithGzip {
			w.Header().Set("Content-Encoding", "gzip")
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		readGzipRequest := strings.Contains(contentEncoding, "gzip")
		if readGzipRequest {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, "compressing error", http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(ow, r)
	})
}
