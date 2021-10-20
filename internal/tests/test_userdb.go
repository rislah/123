package tests

import (
	"context"
	"testing"
	"time"

	app "github.com/rislah/fakes/internal"

	"github.com/stretchr/testify/assert"
)

type MakeUserDB func() (app.UserDB, func() error, error)

func TestUserDB(t *testing.T, makeUserDB MakeUserDB) {
	tests := []struct {
		name  string
		users []app.User
		test  func(ctx context.Context, t *testing.T, db app.UserDB, users ...app.User)
	}{
		{
			name: "create a user and read it back",
			users: []app.User{
				{
					Username: "user",
					Password: "pw",
					Role:     "guest",
				},
			},
			test: func(ctx context.Context, t *testing.T, db app.UserDB, users ...app.User) {
				if err := db.CreateUser(ctx, users[0]); err != nil {
					t.Fatal(err)
				}

				usersCreated, err := db.GetUsers(ctx)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, len(usersCreated), 1)
				assert.Equal(t, usersCreated[0].Username, users[0].Username)
				assert.Equal(t, usersCreated[0].Password, users[0].Password)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			db, teardown, err := makeUserDB()
			if err != nil {
				t.Fatal(err)
			}

			defer func() {
				err := teardown()
				assert.NoError(t, err)
			}()

			test.test(ctx, t, db, test.users...)
		})
	}
}
