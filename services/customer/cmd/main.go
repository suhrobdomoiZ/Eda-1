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

	// metrics
	m := metrics.NewMetrics("customer")
	go func() {
		http.Handle("/metrics", metrics.Handler())
		log.Printf("customer metrics server listening on :%s", cfg.Metrics.Port)
		log.Fatal(http.ListenAndServe(":"+cfg.Metrics.Port, nil))
	}()

	// postgreSQL
	pgRepo, err := repository.NewPostgresRepo(cfg.Postgres.DSN())
	if err != nil {
		log.Fatalf("postgres: %v", err)
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

	consumer := kafka.NewConsumer(*kafkaCfg, nil)
	go func() {
		consumerHandler := handlers.NewOrderConsumerHandler(customerSvc)
		if err := consumer.Start(ctx, consumerHandler); err != nil {
			log.Printf("kafka consumer stopped: %v", err)
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
		log.Fatalf("listen: %v", err)
	}

	log.Printf("customer gRPC server listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
