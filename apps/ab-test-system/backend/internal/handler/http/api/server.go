package api

import (
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/apikey"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/project"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/user"
)

type Server struct {
	*user.UserHandler
	*project.ProjectHandler
	*apikey.APIKeyHandler
}

func NewServer(u *user.UserHandler, p *project.ProjectHandler, k *apikey.APIKeyHandler) *Server {
	return &Server{UserHandler: u, ProjectHandler: p, APIKeyHandler: k}
}
