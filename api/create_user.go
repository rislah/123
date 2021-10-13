package api

import (
	"context"
	"encoding/json"
	app "github.com/rislah/fakes/internal"
	"net/http"
)

func (s *Server) CreateUser(ctx context.Context, req *http.Request) (interface{}, error) {
	var user app.User
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		return nil, err
	}

	if err := s.userService.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return nil, nil
}
