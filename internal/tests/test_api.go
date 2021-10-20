package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtPkg "github.com/golang-jwt/jwt/v4"
	"github.com/rislah/fakes/api"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/app/user"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/geoip"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/rislah/fakes/internal/metrics"
	"github.com/rislah/fakes/internal/password"
	"github.com/rislah/fakes/internal/redis"
	"github.com/stretchr/testify/assert"
)

type apiTestCase struct {
	am      *api.Mux
	db      app.UserDB
	metrics metrics.Metrics
	redis   redis.Client
	user    user.User

	loginReq api.LoginRequest
}

type MakeUser func() (user.User, error)
type MakeMetrics func() metrics.Metrics
type MakeRedis func() (redis.Client, func() error, error)

func TestAPIGetUsers(t *testing.T, makeUserDB MakeUserDB, makeMetrics MakeMetrics, makeRedis MakeRedis) {
	tests := []struct {
		name string
		test func(ctx context.Context, apiTestCase apiTestCase)
	}{
		{
			name: "should return error if no users",
			test: func(ctx context.Context, apiTestCase apiTestCase) {
				req, err := http.NewRequest("GET", "/users", nil)
				assert.NoError(t, err)

				rr := httptest.NewRecorder()
				resp := &api.Response{ResponseWriter: rr}
				ctxIP := addIPToContext(ctx, net.ParseIP("127.0.0.1"))

				err = apiTestCase.am.GetUsers(ctxIP, resp, req)
				assert.NoError(t, err)
				assert.Equal(t, rr.Result().StatusCode, 404)
			},
		},
		{
			name: "should return users",
			test: func(ctx context.Context, apiTestCase apiTestCase) {
				err := apiTestCase.db.CreateUser(ctx, app.User{Firstname: "fname", Lastname: "lastname"})
				assert.NoError(t, err)

				req, err := http.NewRequest("GET", "/users", nil)
				ctxIP := addIPToContext(ctx, net.ParseIP("127.0.0.1"))
				assert.NoError(t, err)
				rr := httptest.NewRecorder()
				resp := &api.Response{ResponseWriter: rr}

				err = apiTestCase.am.GetUsers(ctxIP, resp, req)
				assert.NoError(t, err)
				assert.Equal(t, rr.Result().StatusCode, 200)

				var users []app.User
				err = json.NewDecoder(rr.Body).Decode(&users)
				assert.NoError(t, err)
				assert.NotEmpty(t, users)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			db, teardown, err := makeUserDB()
			assert.NoError(t, err)

			defer teardown()

			jwtWrapper := jwt.NewHS256Wrapper("secret")

			usr := user.NewUser(db, jwtWrapper)
			assert.NoError(t, err)

			redis, teardown, err := makeRedis()
			assert.NoError(t, err)

			defer teardown()

			metrics := makeMetrics()
			apiMux := api.NewMux(usr, jwtWrapper, geoip.GeoIP{}, redis, metrics, nil)
			test.test(ctx, apiTestCase{
				am:      apiMux,
				db:      db,
				user:    usr,
				metrics: metrics,
				redis:   redis,
			})
		})
	}
}

func TestAPILogin(t *testing.T, makeUserDB MakeUserDB, makeMetrics MakeMetrics, makeRedis MakeRedis) {
	tests := []struct {
		name     string
		loginReq api.LoginRequest
		test     func(ctx context.Context, apiTestCase apiTestCase)
	}{
		{
			name: "should return error if body doesnt meet validation",
			loginReq: api.LoginRequest{
				Username: "asd",
				Password: "asd",
			},
			test: func(ctx context.Context, apiTestCase apiTestCase) {
				b, err := json.Marshal(apiTestCase.loginReq)
				assert.NoError(t, err)

				buf := bytes.NewBuffer(b)
				req, err := http.NewRequest("POST", "/login", buf)
				assert.NoError(t, err)

				rr := httptest.NewRecorder()
				resp := &api.Response{ResponseWriter: rr}
				ctxIP := addIPToContext(ctx, net.ParseIP("127.0.0.1"))

				err = apiTestCase.am.Login(ctxIP, resp, req)
				assert.NoError(t, err)

				var httpErrResponse errors.ErrorResponse
				err = json.NewDecoder(rr.Body).Decode(&httpErrResponse)
				assert.NoError(t, err)
				assert.Equal(t, httpErrResponse.Message, password.ErrPasswordLength.Msg)
				assert.Equal(t, httpErrResponse.Status, int(password.ErrPasswordLength.Code))
			},
		},
		{
			name: "should return error if user not found",
			loginReq: api.LoginRequest{
				Username: "test_username",
				Password: "test_password",
			},
			test: func(ctx context.Context, apiTestCase apiTestCase) {
				b, err := json.Marshal(apiTestCase.loginReq)
				assert.NoError(t, err)

				buf := bytes.NewBuffer(b)
				req, err := http.NewRequest("POST", "/login", buf)
				assert.NoError(t, err)

				rr := httptest.NewRecorder()
				resp := &api.Response{ResponseWriter: rr}
				ctxIP := addIPToContext(ctx, net.ParseIP("127.0.0.1"))

				err = apiTestCase.am.Login(ctxIP, resp, req)
				assert.NoError(t, err)

				var errResponse errors.ErrorResponse
				err = json.NewDecoder(rr.Body).Decode(&errResponse)
				assert.NoError(t, err)
				assert.Equal(t, errResponse.Message, user.ErrUserNotFound.Msg)
				assert.Equal(t, errResponse.Status, int(user.ErrUserNotFound.Code))
			},
		},

		{
			name: "should return valid jwt on login",
			loginReq: api.LoginRequest{
				Username: "test_username",
				Password: "test_password",
			},
			test: func(ctx context.Context, apiTestCase apiTestCase) {
				hashedPassword, err := password.NewPassword("test_password").GenerateBCrypt()
				assert.NoError(t, err)
				err = apiTestCase.db.CreateUser(ctx, app.User{
					Username: "test_username",
					Password: hashedPassword,
				})
				assert.NoError(t, err)

				b, err := json.Marshal(apiTestCase.loginReq)
				assert.NoError(t, err)

				buf := bytes.NewBuffer(b)
				req, err := http.NewRequest("POST", "/login", buf)
				assert.NoError(t, err)

				rr := httptest.NewRecorder()
				resp := &api.Response{ResponseWriter: rr}
				ctxIP := addIPToContext(ctx, net.ParseIP("127.0.0.1"))

				err = apiTestCase.am.Login(ctxIP, resp, req)
				assert.NoError(t, err)

				var httpResponse api.LoginResponse
				err = json.NewDecoder(rr.Body).Decode(&httpResponse)
				assert.NoError(t, err)
				assert.NotEmpty(t, httpResponse.Token)

				jw := jwt.Wrapper{
					Algorithm: jwtPkg.SigningMethodHS256,
					Secret:    user.JWTSecret,
				}
				token, err := jw.Decode(httpResponse.Token, &jwt.UserClaims{})
				assert.NoError(t, err)

				uc, ok := token.Claims.(*jwt.UserClaims)
				assert.True(t, ok)
				assert.Equal(t, apiTestCase.loginReq.Username, uc.Username)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			db, teardown, err := makeUserDB()
			assert.NoError(t, err)

			defer teardown()

			jwtWrapper := jwt.NewHS256Wrapper("secret")

			usr := user.NewUser(db, jwtWrapper)
			assert.NoError(t, err)

			redis, teardown, err := makeRedis()
			assert.NoError(t, err)

			defer teardown()

			metrics := makeMetrics()
			apiMux := api.NewMux(usr, jwtWrapper, geoip.GeoIP{}, redis, metrics, nil)
			test.test(ctx, apiTestCase{
				am:       apiMux,
				db:       db,
				user:     usr,
				metrics:  metrics,
				redis:    redis,
				loginReq: test.loginReq,
			})
		})
	}
}
func addIPToContext(ctx context.Context, ip net.IP) context.Context {
	return context.WithValue(ctx, api.RemoteIPContextKey, ip)
}
