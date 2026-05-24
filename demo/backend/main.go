package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drakoRRR/ab-test-system/sdk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("config error")
	}

	client, err := sdk.New(sdk.Config{
		APIKey:     cfg.APIKey,
		ServiceURL: cfg.ServiceURL,
		Logger:     log.Logger,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("sdk init failed")
	}
	defer client.Close()

	h := &Handler{sdk: client, cfg: cfg}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      newRouter(h, cfg.StaticDir),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("demo server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}
	log.Info().Msg("demo server stopped")
}
