package api

import (
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/analytics"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/apikey"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/experiment"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/flag"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/organization"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/project"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/user"
)

type Server struct {
	*user.UserHandler
	*project.ProjectHandler
	*apikey.APIKeyHandler
	*flag.FlagHandler
	*experiment.ExperimentHandler
	*analytics.AnalyticsHandler
	*organization.Handler
}

func NewServer(
	u *user.UserHandler,
	p *project.ProjectHandler,
	k *apikey.APIKeyHandler,
	f *flag.FlagHandler,
	e *experiment.ExperimentHandler,
	a *analytics.AnalyticsHandler,
	o *organization.Handler,
) *Server {
	return &Server{
		UserHandler:       u,
		ProjectHandler:    p,
		APIKeyHandler:     k,
		FlagHandler:       f,
		ExperimentHandler: e,
		AnalyticsHandler:  a,
		Handler:           o,
	}
}
