package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"alpineworks.io/flux-suspension-exporter/internal/config"
	"alpineworks.io/flux-suspension-exporter/internal/flux"
	"alpineworks.io/flux-suspension-exporter/internal/logging"
	"alpineworks.io/ootel"
)

func main() {
	slogHandler := slog.NewJSONHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(slogHandler))

	slog.Info("welcome to flux-suspension-exporter!")

	c, err := config.NewConfig()
	if err != nil {
		slog.Error("could not create config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slogLevel, err := logging.LogLevelToSlogLevel(c.LogLevel)
	if err != nil {
		slog.Error("could not parse log level", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.SetLogLoggerLevel(slogLevel)

	ctx := context.Background()

	ootelClient := ootel.NewOotelClient(
		ootel.WithMetricConfig(
			ootel.NewMetricConfig(
				true,
				c.MetricsPort,
			),
		),
	)

	shutdown, err := ootelClient.Init(ctx)
	if err != nil {
		slog.Error("failed to initialize ootel client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer func() {
		_ = shutdown(ctx)
	}()

	kubernetesClient, err := flux.NewKubernetesClient()
	if err != nil {
		slog.Error("failed to create kubernetes client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	err = flux.NewMetrics(kubernetesClient)
	if err != nil {
		slog.Error("failed to create metrics", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// blocking channel to keep the program running
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
}
