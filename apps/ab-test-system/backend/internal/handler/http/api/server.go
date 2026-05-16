package api

import (
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/project"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/user"
)

type Server struct {
	*user.UserHandler
	*project.ProjectHandler
}

func NewServer(u *user.UserHandler, p *project.ProjectHandler) *Server {
	return &Server{UserHandler: u, ProjectHandler: p}
}
