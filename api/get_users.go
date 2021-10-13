package api

import (
	"net/http"

	"github.com/rislah/fakes/internal/encoder"
)

func (s *Server) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.userService.GetUsers(r.Context())
	if err != nil {
		if err := encoder.ServeError(w, r, err); err != nil {
			s.logger.LogError(err, "error sending response")
			return
		}
		return
	}
	encoder.ServeJSON(w, users, 200)
}
