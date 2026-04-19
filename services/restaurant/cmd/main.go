package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/suhrobdomoiZ/Eda-1/pkg/config"
	"github.com/suhrobdomoiZ/Eda-1/pkg/kafka"
	"github.com/suhrobdomoiZ/Eda-1/services/api"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/handlers"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/repository"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/service"
	"google.golang.org/grpc"
)

func buildDSN() string {
	user := config.Key("DATABASE_USER").MustGet()
	pass := config.Key("DATABASE_PASSWORD").MustGet()
	name := config.Key("DATABASE_NAME").MustGet()
	host := config.Key("DATABASE_HOST").MustGet()
	port := config.Key("DATABASE_PORT").MustGet()

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, name,
	)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	dbDSN := buildDSN()
	pool, err := pgxpool.New(ctx, dbDSN)
	if err != nil {
		logger.Error("restaurant service: failed to create pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("restaurant service: db ping failed", "error", err)
		os.Exit(1)
	}

	logger.Info("restaurant service: connected to db", "host", config.Key("DATABASE_HOST").Get(""))

	kafkaCfg := kafka.Load()
	producer := kafka.NewProducer(*kafkaCfg)
	defer producer.Close()

	consumer := kafka.NewConsumer(*kafkaCfg, logger)

	repo := repository.NewRestaurant(pool)
	svc := service.NewRestaurant(repo, producer)
	grpcHandler := handlers.NewRestaurant(*svc)

	grpcPort := config.Key("RESTAURANT_GRPC_PORT").MustGet()
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Error("restaurant service: failed to listen", "port", grpcPort, "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	api.RegisterRestaurantServer(grpcServer, grpcHandler)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("restaurant service: ready to serve", "port", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server error", "error", err)
		}
	}()

	<-stop
	logger.Info("shutdown signal received")

	grpcServer.GracefulStop()
	if err := consumer.Close(); err != nil {
		logger.Error("failed to close kafka consumer", "error", err)
	}

	logger.Info("server stopped gracefully")
}
