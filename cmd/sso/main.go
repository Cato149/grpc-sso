package main

import (
	"awesomeProject/internal/app"
	"awesomeProject/internal/config"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(cfg.Env)
	log := startupLogger(cfg.Env)
	log.Info("Sturtup app",
		slog.Any("cfg", cfg),
	)

	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)
	go application.GRPCServer.Run()

	// TODO: Start-up gRPC-service
	// TODO: Make pretty logger with colorful text

	// Idea: we started server in goroutine and now waiting OS signal for shut down
	stop := make(chan os.Signal, 1)

	// TODO: Read about SIGINT and SIGTERM more
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// We got signal and then stop our app
	syssig := <-stop

	log.Info("Shutting down", syssig)
	//TODO: Wrap SQL server into app
	application.GRPCServer.Stop()
	log.Info("Application stopped")
}

func startupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	}
	return log
}
