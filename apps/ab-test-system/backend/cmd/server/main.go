package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/infra/gcp"
	gcpauth "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/infra/gcp/auth"
	infrakafka "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/infra/kafka"
	infrapostgres "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/infra/postgres"
	infraredis "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/infra/redis"

	httphandler "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http"
	httpapi "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api"
	apikeyhandler "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/apikey"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
	sdkgen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen/sdk"
	experimenthandler "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/experiment"
	flaghandler "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/flag"
	projecthandler "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/project"
	sdkhandler "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/sdk"
	userhandler "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/user"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"

	postgresapikey "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/infra/postgres/apikey"
	postgresexperiment "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/infra/postgres/experiment"
	postgresflag "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/infra/postgres/flag"
	postgresproject "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/infra/postgres/project"
	postgresuser "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/infra/postgres/user"
	apikeyservice "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/apikey"
	eventservice "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/event"
	experimentservice "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/experiment"
	flagservice "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/flag"
	projectservice "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/project"
	sdkservice "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/sdk"
	userservice "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/user"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/pkg/config"
)

type Config struct {
	HTTPServer httphandler.Config   `yaml:"http_server"`
	GCP        gcp.Config           `yaml:"gcp_auth"`
	Database   infrapostgres.Config `yaml:"database"`
	Redis      infraredis.Config    `yaml:"redis"`
	Kafka      infrakafka.Config    `yaml:"kafka"`
}

var configPath = flag.String("config", "./config/example.yaml", "Path to config file")

func main() {
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var cfg Config
	if err := config.Load(*configPath, &cfg); err != nil {
		log.Fatal().Err(err).Msg("loading config")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := infrapostgres.New(ctx, cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("connecting to postgres")
	}
	defer db.Close()
	log.Info().Msg("postgres connected")

	rdb, err := infraredis.New(ctx, cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("connecting to redis")
	}
	defer func() {
		if err := rdb.Close(); err != nil {
			log.Error().Err(err).Msg("closing redis")
		}
	}()
	log.Info().Msg("redis connected")

	fbAuth, err := gcpauth.New(ctx, cfg.GCP)
	if err != nil {
		log.Fatal().Err(err).Msg("initialising firebase auth")
	}
	log.Info().Msg("firebase auth initialised")

	// Repos → Services → Handlers
	userRepo := postgresuser.NewRepo(db)
	userSvc := userservice.NewService(userRepo)
	userH := userhandler.NewHandler(userSvc)

	projectRepo := postgresproject.NewRepo(db)
	projectSvc := projectservice.NewService(projectRepo)
	projectH := projecthandler.NewHandler(userSvc, projectSvc)

	apiKeyRepo := postgresapikey.NewRepo(db)
	apiKeySvc := apikeyservice.NewService(apiKeyRepo)
	apiKeyH := apikeyhandler.NewHandler(apiKeySvc)

	redisCacheClient := infraredis.NewCache(rdb)

	flagRepo := postgresflag.NewRepo(db)
	cachedFlagRepo := postgresflag.NewCachedRepo(flagRepo, redisCacheClient, log.Logger)
	flagSvc := flagservice.NewService(cachedFlagRepo)
	flagH := flaghandler.NewHandler(flagSvc)

	experimentRepo := postgresexperiment.NewRepo(db)
	cachedExperimentRepo := postgresexperiment.NewCachedRepo(experimentRepo, redisCacheClient, log.Logger)
	experimentSvc := experimentservice.NewService(cachedExperimentRepo)
	experimentH := experimenthandler.NewHandler(experimentSvc)

	producers := infrakafka.NewProducerRegistry(cfg.Kafka.Brokers, cfg.Kafka.Producers)
	defer func() {
		if err := producers.Close(); err != nil {
			log.Error().Err(err).Msg("closing kafka producers")
		}
	}()

	sdkSvc := sdkservice.NewService(cachedFlagRepo, cachedExperimentRepo)
	eventSvc := eventservice.NewService(producers.Get("events"))
	sdkH := sdkhandler.NewHandler(sdkSvc, eventSvc)

	server := httpapi.NewServer(userH, projectH, apiKeyH, flagH, experimentH)

	router := mux.NewRouter()

	// Control-plane routes — Firebase JWT auth.
	strict := gen.NewStrictHandler(server, nil)
	gen.HandlerWithOptions(strict, gen.GorillaServerOptions{
		BaseURL:     "/api/v1",
		BaseRouter:  router,
		Middlewares: []gen.MiddlewareFunc{middleware.Auth(fbAuth)},
	})

	// SDK routes — API key auth applied per-operation via HandlerWithOptions.
	sdkStrict := sdkgen.NewStrictHandler(sdkH, nil)
	sdkgen.HandlerWithOptions(sdkStrict, sdkgen.GorillaServerOptions{
		BaseURL:     "/api/v1",
		BaseRouter:  router,
		Middlewares: []sdkgen.MiddlewareFunc{middleware.ApiKeyAuth(apiKeySvc)},
	})

	httpSrv := httphandler.NewServer(cfg.HTTPServer, router)

	// Start
	go func() {
		log.Info().Int("port", cfg.HTTPServer.Port).Msg("HTTP server starting")
		if err := httpSrv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("HTTP server shutdown error")
	}

	log.Info().Msg("stopped")
}
