package api

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rislah/fakes/internal/geoip"
	"github.com/rislah/fakes/internal/logger"
	"github.com/rislah/fakes/internal/metrics"
	"github.com/segmentio/stats/v4"
	"github.com/sirupsen/logrus"
)

type response struct {
	http.ResponseWriter
	status int
}

func (r *response) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func requestsLoggerMiddleware(l *logger.Logger, gip geoip.GeoIP) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			h.ServeHTTP(rw, r)

			ipPort := strings.Split(r.RemoteAddr, ":")
			ip := net.ParseIP(ipPort[0])
			country, err := gip.LookupCountry(ip)
			if err != nil {
				l.Warn("couldn't look up country", err, logrus.Fields{"ip": ip.String()})
			}

			duration := time.Since(startTime)
			durationMs := float32(duration.Nanoseconds()/1000) / 1000.0
			l.LogRequest(r, logrus.Fields{"duration_ms": durationMs, "country": country})
		})
	}
}

func metricsMiddleware(m metrics.Metrics) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &response{
				ResponseWriter: w,
			}
			startTime := time.Now()

			h.ServeHTTP(rw, r)

			statusCode := rw.status
			path, method := metricLabels(r)

			m.Inc(metrics.HttpRequestName, stats.T("path", path), stats.T("method", method), stats.T("statusCode", strconv.Itoa(statusCode)))
			m.Measure(metrics.HttpRequestName, time.Since(startTime), stats.T("path", path), stats.T("method", method), stats.T("statusCode", strconv.Itoa(statusCode)))
		})
	}
}

func metricLabels(r *http.Request) (string, string) {
	cr := mux.CurrentRoute(r)
	path, _ := cr.GetPathTemplate()
	methods, _ := cr.GetMethods()

	path = strings.TrimLeft(path, "/")
	path = strings.Replace(path, "/", "-", 1)
	path = strings.Replace(path, ":", "_", 1)

	return path, methods[0]
}

func contextMiddleWare() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ipPort := strings.Split(r.RemoteAddr, ":")
			ip := net.ParseIP(ipPort[0])

			ctx := context.WithValue(r.Context(), "remote_ip", ip)

			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
