package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/geoip"
	"github.com/rislah/fakes/internal/jwt"

	"github.com/rislah/fakes/internal/logger"
	"github.com/rislah/fakes/internal/metrics"
	"github.com/segmentio/stats/v4"
	"github.com/sirupsen/logrus"
)

type Response struct {
	http.ResponseWriter
	wasWritten bool
	status     int
}

func (r *Response) WriteHeader(status int) {
	r.status = status
	r.wasWritten = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *Response) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
		r.WriteHeader(http.StatusOK)
	}
	l, err := r.ResponseWriter.Write(b)
	if err != nil {
		r.wasWritten = true
	}
	return l, err
}

func (r *Response) WriteJSON(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	r.ResponseWriter.Header().Add("Content-Type", "application/json;charset=utf-8")
	r.ResponseWriter.Header().Add("Content-Length", strconv.Itoa(len(b)))
	_, err = r.ResponseWriter.Write(b)
	return err
}

func (r *Response) Status() int {
	return r.status
}

func (r *Response) WasWritten() bool {
	return r.wasWritten
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
				l.WarnWithFields("couldn't look up country", err, logrus.Fields{"ip": ip.String()})
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
			rw := &Response{
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

type contextKey string

const RemoteIPContextKey contextKey = "remote_ip"

func contextMiddleWare(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipPort := strings.Split(r.RemoteAddr, ":")
		ip := net.ParseIP(ipPort[0])
		ctx := context.WithValue(r.Context(), RemoteIPContextKey, ip)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

var ErrAuthMissingBearerToken = &errors.WrappedError{
	Msg:  "Missing authorization token",
	Code: http.StatusUnauthorized,
}

var ErrAuthInsufficientPrivileges = &errors.WrappedError{
	Msg:  "Insufficient privileges",
	Code: http.StatusUnauthorized,
}

const JwtClaimsKey contextKey = "jwt_claims"

func AuthenticationMiddleware(handler http.Handler, jw jwt.Wrapper, roles ...string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		resp := &Response{ResponseWriter: rw}
		ctx := r.Context()
		jwtClaims := ctx.Value(JwtClaimsKey)

		if jwtClaims == nil {
			bearerToken, err := extractAuthorizationBearerToken(r)
			if err != nil {
				e, _ := err.(*errors.WrappedError)
				resp.WriteHeader(int(e.Code))
				resp.WriteJSON(errors.NewErrorResponse(e.Msg, int(e.Code)))
				return
			}

			decoded, err := jw.Decode(bearerToken, &jwt.UserClaims{})
			if err != nil {
				if e, ok := errors.IsWrappedError(ctx, err); ok {
					resp.WriteHeader(int(e.Code))
					resp.WriteJSON(errors.NewErrorResponse(e.Msg, int(e.Code)))
					return
				}

				resp.WriteHeader(http.StatusInternalServerError)
				resp.WriteJSON(errors.NewErrorResponse("Internal server error has occured", http.StatusInternalServerError))
				logger.SharedGlobalLogger.LogRequestError(err, r)
				return
			}

			jwtClaims = decoded.Claims
			ctx = context.WithValue(ctx, JwtClaimsKey, decoded.Claims)
		}

		userClaims, ok := jwtClaims.(*jwt.UserClaims)
		if !ok {
			// todo
			return
		}

		if !userIsInRole(userClaims.Role, roles...) {
			resp.WriteHeader(int(ErrAuthInsufficientPrivileges.Code))
			resp.WriteJSON(errors.NewErrorResponse(ErrAuthInsufficientPrivileges.Msg, int(ErrAuthInsufficientPrivileges.Code)))
			return
		}

		handler.ServeHTTP(resp, r.WithContext(ctx))
	})
}

func extractAuthorizationBearerToken(r *http.Request) (string, error) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		return "", ErrAuthMissingBearerToken
	}

	prefixBearer := strings.Split(authorization, "Bearer")
	if len(prefixBearer) != 2 {
		return "", ErrAuthMissingBearerToken
	}

	bearer := strings.TrimSpace(prefixBearer[1])
	if bearer == "" {
		return "", ErrAuthMissingBearerToken
	}

	return bearer, nil
}

func userIsInRole(userRole string, roles ...string) bool {
	for _, r := range roles {
		if r == userRole {
			return true
		}
	}

	return false
}
