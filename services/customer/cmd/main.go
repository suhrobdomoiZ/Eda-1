package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/suhrobdomoiZ/Eda-1/pkg/api/customer"
	"github.com/suhrobdomoiZ/Eda-1/pkg/kafka"
	metrics "github.com/suhrobdomoiZ/Eda-1/pkg/metrics"
	"github.com/suhrobdomoiZ/Eda-1/services/customer/internal/config"
	"github.com/suhrobdomoiZ/Eda-1/services/customer/internal/handlers"
	"github.com/suhrobdomoiZ/Eda-1/services/customer/internal/repository"
	service "github.com/suhrobdomoiZ/Eda-1/services/customer/internal/services"
)

func main() {
	// config
	cfg := config.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// metrics
	m := metrics.NewMetrics("customer")
	go func() {
		http.Handle("/metrics", metrics.Handler())
		logger.Info("customer metrics server listening", "port", cfg.Metrics.Port)
		logger.Error(http.ListenAndServe(":"+cfg.Metrics.Port, nil).Error())
	}()

	// postgreSQL
	pgRepo, err := repository.NewPostgresRepo(cfg.Postgres.DSN())
	if err != nil {
		logger.Error("postgres: %v", err)
	}
	defer pgRepo.Close()

	// Kafka producer
	kafkaCfg := kafka.Load()
	producer := kafka.NewProducer(*kafkaCfg)
	defer producer.Close()

	// service
	customerSvc := service.NewCustomerService(pgRepo, producer, m)
	customerHandler := handlers.NewCustomerHandler(customerSvc)

	// Kafka Consumer
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	consumer := kafka.NewConsumer(*kafkaCfg, logger)
	go func() {
		consumerHandler := handlers.NewOrderConsumerHandler(customerSvc)
		if err := consumer.Start(ctx, consumerHandler); err != nil {
			logger.Info("kafka consumer stopped:", "error", err)
		}
	}()
	defer consumer.Close()

	// gRPC srver
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(m.UnaryServerInterceptor()),
	)
	pb.RegisterCustomerAPIServer(grpcServer, customerHandler)
	reflection.Register(grpcServer)

	addr := fmt.Sprintf(":%s", cfg.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("listen: %v", err)
	}

	logger.Info("customer gRPC server listening", "port", addr)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("serve: %v", err)
	}
}
