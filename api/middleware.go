package api

import (
	"github.com/rislah/fakes/internal/metrics"
	"net/http"
)

type Response struct {
	http.ResponseWriter
	status int
}

func (r *Response) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func metricsMiddleware(m metrics.Metrics) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &Response{
				ResponseWriter: w,
			}
			h.ServeHTTP(rw, r)

			defer func() {
				if rw.status <= 599 && rw.status >= 500 {
					m.IncrementCounter(r.Method, metrics.Tags{
						Key:   "http_status",
						Value: "5xx",
					})
					return
				}

				if rw.status <= 499 && rw.status >= 400 {
					m.IncrementCounter(r.Method, metrics.Tags{
						Key:   "http_status",
						Value: "4xx",
					})
					return
				}

				if rw.status <= 308 && rw.status >= 300 {
					m.IncrementCounter(r.Method, metrics.Tags{
						Key:   "http_status",
						Value: "3xx",
					})
					return
				}

				if rw.status <= 226 && rw.status >= 200 {
					m.IncrementCounter(r.Method, metrics.Tags{
						Key:   "http_status",
						Value: "2xx",
					})
					return
				}
			}()
		})
	}
}
