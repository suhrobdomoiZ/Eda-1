package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/suhrobdomoiZ/Eda-1/pkg/closer"
	"github.com/suhrobdomoiZ/Eda-1/pkg/config"
	"github.com/suhrobdomoiZ/Eda-1/pkg/kafka"
	pb "github.com/suhrobdomoiZ/Eda-1/services/api"
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

	clsr := closer.New(*logger)

	dbDSN := buildDSN()
	pool, err := pgxpool.New(ctx, dbDSN)
	if err != nil {
		logger.Error("restaurant service: failed to create pool", "error", err)
		os.Exit(1)
	}

	if err := pool.Ping(ctx); err != nil {
		logger.Error("restaurant service: db ping failed", "error", err)
		os.Exit(1)
	}

	clsr.AddFunc("pool", pool.Close)

	logger.Info("restaurant service: connected to db", "host", config.Key("DATABASE_HOST").Get(""))

	kafkaCfg := kafka.Load()
	producer := kafka.NewProducer(*kafkaCfg)

	clsr.AddFunc("kafka producer", func() {
		_ = producer.Close()
	})

	repo := repository.NewRestaurant(pool)
	svc := service.NewRestaurant(repo, producer)
	grpcHandler := handlers.NewRestaurant(svc)

	consumer := kafka.NewConsumer(*kafkaCfg, logger)

	go func() {
		consumerHandler := handlers.NewOrderConsumerHandler(svc, logger)
		err = consumer.Start(ctx, consumerHandler)
		if err := consumer.Start(ctx, consumerHandler); err != nil {
			logger.Error("kafka consumer stopped unexpectedly", "error", err)
		}
	}()

	clsr.AddFunc("kafka consumer", func() {
		_ = consumer.Close()
	})

	grpcPort := config.Key("RESTAURANT_GRPC_PORT").MustGet()
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Error("restaurant service: failed to listen", "port", grpcPort, "error", err)
		os.Exit(1)
	}
	clsr.AddFunc("grpc listener", func() {
		_ = lis.Close()
	})

	grpcServer := grpc.NewServer()
	pb.RegisterRestaurantServer(grpcServer, grpcHandler)

	clsr.Add("grpc server", func(ctx context.Context) error {
		done := make(chan struct{})

		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			return nil
		case <-ctx.Done():
			grpcServer.Stop()
			<-done
			return ctx.Err()
		}
	})

	errCh := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			errCh <- fmt.Errorf("server.Run: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		logger.Error("restaurant service: error occurred", "error", err)
		os.Exit(1)
	case sig := <-sigCh:
		logger.Info("restaurant service: received signal", "signal", sig.String())
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := clsr.Close(shutdownCtx); err != nil && errors.Is(err, context.DeadlineExceeded) {
			logger.Error("restaurant service: graceful shutdown", "error", err)
		}
	}

	logger.Info("server stopped")
	os.Exit(0)
}
