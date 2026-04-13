package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/suhrobdomoiZ/Eda-1/services/customer/internal/api"
	"github.com/suhrobdomoiZ/Eda-1/services/customer/internal/config"
	"github.com/suhrobdomoiZ/Eda-1/services/customer/internal/handlers"
	"github.com/suhrobdomoiZ/Eda-1/services/customer/internal/repository"
	service "github.com/suhrobdomoiZ/Eda-1/services/customer/internal/services"
)

func main() {
	cfg := config.Load()

	pgRepo, err := repository.NewPostgresRepo(cfg.Postgres.DSN())
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pgRepo.Close()

	customerSvc := service.NewCustomerService(pgRepo)
	customerHandler := handlers.NewCustomerHandler(customerSvc)

	grpcServer := grpc.NewServer()
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
