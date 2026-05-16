package api

import (
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/user"
)

type Server struct {
	*user.Handler
}

func NewServer(u *user.Handler) *Server {
	return &Server{Handler: u}
}
