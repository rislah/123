package tests

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rislah/fakes/api"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/app/user"
	"github.com/rislah/fakes/internal/geoip"
	"github.com/rislah/fakes/internal/metrics"
	"github.com/rislah/fakes/internal/redis"
	"github.com/stretchr/testify/assert"
)

type apiTestCase struct {
	am      *api.Mux
	db      app.UserDB
	metrics metrics.Metrics
	redis   redis.Client
	user    user.User
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

			usr := user.NewUser(db)
			assert.NoError(t, err)

			redis, teardown, err := makeRedis()
			assert.NoError(t, err)

            defer teardown()

			metrics := makeMetrics()
			apiMux := api.NewMux(usr, geoip.GeoIP{}, redis, metrics, nil)
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

func addIPToContext(ctx context.Context, ip net.IP) context.Context {
    return context.WithValue(ctx, "remote_ip", ip)
}
