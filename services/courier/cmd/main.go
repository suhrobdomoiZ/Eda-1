package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/suhrobdomoiZ/Eda-1/pkg/api/courier"
	metrics "github.com/suhrobdomoiZ/Eda-1/pkg/metrics"
	"github.com/suhrobdomoiZ/Eda-1/services/courier/internal/config"
	"github.com/suhrobdomoiZ/Eda-1/services/courier/internal/handlers"
	"github.com/suhrobdomoiZ/Eda-1/services/courier/internal/repository"
	service "github.com/suhrobdomoiZ/Eda-1/services/courier/internal/services"
)

func main() {
	cfg := config.Load()

	m := metrics.NewMetrics("courier")
	go func() {
		http.Handle("/metrics", metrics.Handler())
		log.Printf("courier metrics server listening on :%s", cfg.Metrics.Port)
		log.Fatal(http.ListenAndServe(":"+cfg.Metrics.Port, nil))
	}()

	pgRepo, err := repository.NewPostgresRepo(cfg.Postgres.DSN())
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pgRepo.Close()

	courierSvc := service.NewCourierService(pgRepo, m)
	courierHandler := handlers.NewCourierHandler(courierSvc)

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
