package gzip

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/vadicheck/gofermart/pkg/compress"
)

func New() func(next http.Handler) http.Handler {
	slog.Info("gzip middleware enabled")

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ow := w

			contentType := r.Header.Get("Content-Type")
			acceptEncoding := r.Header.Get("Accept-Encoding")

			supportsGzip := strings.Contains(acceptEncoding, "gzip")

			if supportsGzip && isCompressibleContentType(contentType) {
				cw := compress.NewCompressWriter(w)
				ow = cw
				defer cw.Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := compress.NewCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				r.Body = cr
				defer cr.Close()
			}

			next.ServeHTTP(ow, r)
		}
		return http.HandlerFunc(fn)
	}
}

func isCompressibleContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "application/json")
}
