package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	jwtPkg "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/geoip"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/ratelimiter"

	"github.com/rislah/fakes/internal/logger"
	"github.com/sirupsen/logrus"
)

var (
	httpRequestMeasure = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
	}, []string{"code", "handler", "method"})

	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
	}, []string{"code", "method"})
)

func init() {
	prometheus.Register(httpRequestMeasure)
	prometheus.Register(httpRequestsTotal)
}

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
			country, err := gip.LookupCountryISO(ip)
			if err != nil {
				l.WarnWithFields("couldn't look up country", err, logrus.Fields{"ip": ip.String()})
			}

			duration := time.Since(startTime)
			durationMs := float32(duration.Nanoseconds()/1000) / 1000.0
			l.LogRequest(r, logrus.Fields{"duration_ms": durationMs, "country": country})
		})
	}
}

func metricsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &Response{
			ResponseWriter: w,
		}
		startTime := time.Now()

		h.ServeHTTP(rw, r)

		statusCode := rw.status
		path, method := metricLabels(r)

		var statusCodeLabel int
		if statusCode >= 500 && statusCode <= 599 {
			statusCodeLabel = 500
		} else if statusCodeLabel >= 300 && statusCodeLabel <= 399 {
			statusCodeLabel = 300
		} else if statusCodeLabel >= 400 && statusCodeLabel <= 499 {
			statusCodeLabel = 400
		} else if statusCodeLabel >= 200 && statusCodeLabel <= 299 {
			statusCodeLabel = 200
		}

		httpRequestMeasure.WithLabelValues(strconv.Itoa(statusCodeLabel), path, method).Observe(float64(time.Since(startTime).Seconds()))
		httpRequestsTotal.WithLabelValues(strconv.Itoa(statusCodeLabel), method).Inc()
	})
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

type ContextKey string

const RemoteIPContextKey ContextKey = "remote_ip"

func contextMiddleWare(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipPort := strings.Split(r.RemoteAddr, ":")
		ip := net.ParseIP(ipPort[0])
		ctx := context.WithValue(r.Context(), RemoteIPContextKey, ip)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Mux) ratelimiterMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		response := &Response{ResponseWriter: rw}
		ctx := r.Context()

		ip := ctx.Value(RemoteIPContextKey).(net.IP)
		field := ratelimiter.Field{
			Scope:      "ip",
			Identifier: ip.String(),
		}

		throttled, err := s.globalRatelimiter.ShouldThrottle(ctx, response, field)
		if err != nil {
			s.logger.LogRequestError(errors.Wrap(err, "globalRateLimiter"), r)
		}

		if throttled {
			response.WriteHeader(http.StatusTooManyRequests)
			response.WriteJSON(errors.NewErrorResponse("You are being ratelimited", http.StatusTooManyRequests))
			return
		}

		h.ServeHTTP(response, r)
	})
}

var (
	ErrAuthMissingBearerToken = &errors.WrappedError{
		Msg:  "Missing authorization token",
		Code: http.StatusUnauthorized,
	}

	ErrAuthInsufficientPrivileges = &errors.WrappedError{
		Msg:  "Insufficient privileges",
		Code: http.StatusUnauthorized,
	}
)

const jwtClaimsKey ContextKey = "jwt_claims"

func (r *Route) authMiddleware(h http.Handler, jwtWrapper jwt.Wrapper) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		resp := &Response{ResponseWriter: rw}
		ctx := req.Context()
		jwtClaims := ctx.Value(jwtClaimsKey)

		if jwtClaims == nil {
			bearerToken, err := extractAuthorizationBearerToken(req)
			if err != nil {
				e, _ := err.(*errors.WrappedError)
				resp.WriteHeader(int(e.Code))
				resp.WriteJSON(errors.NewErrorResponse(e.Msg, int(e.Code)))
				return
			}

			decoded, err := jwtWrapper.Decode(bearerToken, &jwt.UserClaims{})
			if err != nil {
				if _, ok := errors.Unwrap(err).(*jwtPkg.ValidationError); ok {
					resp.WriteHeader(int(ErrAuthInsufficientPrivileges.Code))
					resp.WriteJSON(errors.NewErrorResponse(ErrAuthInsufficientPrivileges.Msg, int(ErrAuthInsufficientPrivileges.Code)))
					return
				}

				if e, ok := errors.IsWrappedError(ctx, err); ok {
					resp.WriteHeader(int(e.Code))
					resp.WriteJSON(errors.NewErrorResponse(e.Msg, int(e.Code)))
					return
				}

				resp.WriteHeader(http.StatusInternalServerError)
				resp.WriteJSON(errors.NewErrorResponse("Internal server error has occured", http.StatusInternalServerError))
				logger.SharedGlobalLogger.LogRequestError(err, req)
				return
			}

			jwtClaims = decoded.Claims
			ctx = context.WithValue(ctx, jwtClaimsKey, decoded.Claims)
		}

		userClaims, ok := jwtClaims.(*jwt.UserClaims)
		if !ok {
			// todo
			return
		}

		if len(r.permissions) != 0 {
			for _, permission := range r.permissions {
				if !app.DoesRoleHavePermission(app.Role(userClaims.Role), permission) {
					resp.WriteHeader(int(ErrAuthInsufficientPrivileges.Code))
					resp.WriteJSON(errors.NewErrorResponse(ErrAuthInsufficientPrivileges.Msg, int(ErrAuthInsufficientPrivileges.Code)))
					return
				}
			}
		}

		if r.role != "" {
			if app.Role(userClaims.Role) != r.role {
				resp.WriteHeader(int(ErrAuthInsufficientPrivileges.Code))
				resp.WriteJSON(errors.NewErrorResponse(ErrAuthInsufficientPrivileges.Msg, int(ErrAuthInsufficientPrivileges.Code)))
				return
			}
		}

		h.ServeHTTP(resp, req.WithContext(ctx))
	})
}

func extractAuthorizationBearerToken(r *http.Request) (string, error) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		return "", ErrAuthMissingBearerToken
	}

	if strings.HasPrefix(authorization, "Bearer") {
		token := strings.TrimPrefix(authorization, "Bearer")
		return strings.TrimSpace(token), nil
	}

	return "", ErrAuthMissingBearerToken
}
