package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/suhrobdomoiZ/Eda-1/pkg/api/courier"
	"github.com/suhrobdomoiZ/Eda-1/pkg/kafka"
	metrics "github.com/suhrobdomoiZ/Eda-1/pkg/metrics"
	"github.com/suhrobdomoiZ/Eda-1/services/courier/internal/config"
	"github.com/suhrobdomoiZ/Eda-1/services/courier/internal/handlers"
	"github.com/suhrobdomoiZ/Eda-1/services/courier/internal/repository"
	service "github.com/suhrobdomoiZ/Eda-1/services/courier/internal/services"
)

func main() {
	cfg := config.Load()

	// metrics
	m := metrics.NewMetrics("courier")
	go func() {
		http.Handle("/metrics", metrics.Handler())
		log.Printf("courier metrics server listening on :%s", cfg.Metrics.Port)
		log.Fatal(http.ListenAndServe(":"+cfg.Metrics.Port, nil))
	}()

	// postgreSQL
	pgRepo, err := repository.NewPostgresRepo(cfg.Postgres.DSN())
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pgRepo.Close()

	// Kafka Producer
	kafkaCfg := kafka.Load()
	producer := kafka.NewProducer(*kafkaCfg)
	defer producer.Close()

	// service
	courierSvc := service.NewCourierService(pgRepo, producer, m)
	courierHandler := handlers.NewCourierHandler(courierSvc)

	// Kafka Consumer
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	consumer := kafka.NewConsumer(*kafkaCfg, nil)
	go func() {
		consumerHandler := handlers.NewOrderConsumerHandler(courierSvc)
		if err := consumer.Start(ctx, consumerHandler); err != nil {
			log.Printf("kafka consumer stopped: %v", err)
		}
	}()
	defer consumer.Close()

	// gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(m.UnaryServerInterceptor()),
	)
	pb.RegisterCourierAPIServer(grpcServer, courierHandler)
	reflection.Register(grpcServer)

	addr := fmt.Sprintf(":%s", cfg.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	log.Printf("courier gRPC server listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
