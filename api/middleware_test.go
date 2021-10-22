package api_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rislah/fakes/api"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/jwt"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticationMiddleware(t *testing.T) {
	tests := []struct {
		scenario     string
		rolesAllowed []string
		test         func(url string, jwtWrapper jwt.Wrapper)
	}{
		{
			scenario:     "no auth bearer token",
			rolesAllowed: []string{"admin"},
			test: func(url string, jwtWrapper jwt.Wrapper) {
				req, err := http.NewRequest("GET", url, nil)
				assert.NoError(t, err)

				client := &http.Client{}
				resp, err := client.Do(req)
				assert.NoError(t, err)

				var httpErrResponse errors.ErrorResponse
				err = json.NewDecoder(resp.Body).Decode(&httpErrResponse)
				assert.NoError(t, err)

				assert.Equal(t, http.StatusUnauthorized, httpErrResponse.Status)
			},
		},
		{
			scenario:     "insufficient role",
			rolesAllowed: []string{"admin"},
			test: func(url string, jwtWrapper jwt.Wrapper) {
				req, err := http.NewRequest("GET", url, nil)
				assert.NoError(t, err)

				tokenStr, err := jwtWrapper.Encode(jwt.NewUserClaims("jaja", "asd"))
				assert.NoError(t, err)
				req.Header.Add("Authorization", "Bearer "+tokenStr)

				client := &http.Client{}
				resp, err := client.Do(req)
				assert.NoError(t, err)

				var httpErrResponse errors.ErrorResponse
				err = json.NewDecoder(resp.Body).Decode(&httpErrResponse)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusUnauthorized, httpErrResponse.Status)
			},
		},
		{
			scenario:     "sufficient role",
			rolesAllowed: []string{"asd"},
			test: func(url string, jwtWrapper jwt.Wrapper) {
				req, err := http.NewRequest("GET", url, nil)
				assert.NoError(t, err)

				tokenStr, err := jwtWrapper.Encode(jwt.NewUserClaims("jaja", "asd"))
				assert.NoError(t, err)
				req.Header.Add("Authorization", "Bearer "+tokenStr)

				client := &http.Client{}
				resp, err := client.Do(req)
				assert.NoError(t, err)

				b, err := ioutil.ReadAll(resp.Body)
				assert.NoError(t, err)
				assert.Equal(t, "ok", string(b))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			jwtWrapper := jwt.NewHS256Wrapper("secret")
			handler := api.AuthenticationMiddleware(testHandler(), jwtWrapper, test.rolesAllowed...)
			srv := httptest.NewServer(handler)
			defer srv.Close()
			test.test(srv.URL, jwtWrapper)
		})
	}
}

func testHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("ok"))
	}
}
