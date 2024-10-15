package app

import (
	grpcApp "awesomeProject/internal/app/grpc"
	"awesomeProject/internal/services/auth"
	"awesomeProject/internal/storage/sqlite"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpcApp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, tokenTTL)

	gRPCApp := grpcApp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: gRPCApp,
	}
}
