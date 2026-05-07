package httpcache

import (
	"bytes"
	"github.com/diftappdev/unicache-enterprise/pkg/cache"
	"net/http"
	"time"
)

type keyFunc func(*http.Request) string

func Middleware(c *cache.Client, ttl time.Duration, key keyFunc) func(http.Handler) http.Handler {
	if key == nil {
		key = func(r *http.Request) string { return r.Method + ":" + r.URL.RequestURI() }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}
			k := key(r)
			if b, err := c.Get(r.Context(), k); err == nil {
				w.Header().Set("X-Cache", "HIT")
				_, _ = w.Write(b)
				return
			}
			rw := &capture{ResponseWriter: w, code: http.StatusOK}
			next.ServeHTTP(rw, r)
			if rw.code >= 200 && rw.code < 300 {
				_ = c.Set(r.Context(), k, rw.buf.Bytes(), ttl)
			}
		})
	}
}

type capture struct {
	http.ResponseWriter
	buf  bytes.Buffer
	code int
}

func (c *capture) WriteHeader(code int)        { c.code = code; c.ResponseWriter.WriteHeader(code) }
func (c *capture) Write(b []byte) (int, error) { c.buf.Write(b); return c.ResponseWriter.Write(b) }
