package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	infrakafka "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/infra/kafka"
	infrapostgres "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/infra/postgres"
	postgresevent "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/infra/postgres/event"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/pkg/config"
)

type Config struct {
	Database infrapostgres.Config `yaml:"database"`
	Kafka    infrakafka.Config    `yaml:"kafka"`
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

	eventsCfg := cfg.Kafka.Consumers["events"]

	eventRepo := postgresevent.NewRepo(db)
	consumer := infrakafka.NewConsumer(cfg.Kafka.Brokers, eventsCfg, eventRepo, log.Logger)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Info().Msg("shutting down consumer")
		cancel()
	}()

	log.Info().Strs("brokers", cfg.Kafka.Brokers).Str("topic", eventsCfg.Topic).Msg("consumer starting")
	if err := consumer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal().Err(err).Msg("consumer error")
	}

	log.Info().Msg("stopped")
}
