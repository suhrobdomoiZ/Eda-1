package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/suhrobdomoiZ/Eda-1/pkg/api/customer"
	metrics "github.com/suhrobdomoiZ/Eda-1/pkg/metrics"
	"github.com/suhrobdomoiZ/Eda-1/services/customer/internal/config"
	"github.com/suhrobdomoiZ/Eda-1/services/customer/internal/handlers"
	"github.com/suhrobdomoiZ/Eda-1/services/customer/internal/repository"
	service "github.com/suhrobdomoiZ/Eda-1/services/customer/internal/services"
)

func main() {
	cfg := config.Load()

	m := metrics.NewMetrics("customer")
	go func() {
		http.Handle("/metrics", metrics.Handler())
		log.Printf("customer metrics server listening on :%s", cfg.Metrics.Port)
		log.Fatal(http.ListenAndServe(":"+cfg.Metrics.Port, nil))
	}()

	pgRepo, err := repository.NewPostgresRepo(cfg.Postgres.DSN())
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pgRepo.Close()

	customerSvc := service.NewCustomerService(pgRepo, m)
	customerHandler := handlers.NewCustomerHandler(customerSvc)

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
