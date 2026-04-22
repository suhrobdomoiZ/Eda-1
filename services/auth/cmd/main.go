package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/suhrobdomoiZ/Eda-1/pkg/api/auth"
	metrics "github.com/suhrobdomoiZ/Eda-1/pkg/metrics"
	api "github.com/suhrobdomoiZ/Eda-1/services/auth/internal/api/server"
	"github.com/suhrobdomoiZ/Eda-1/services/auth/internal/config"
	"github.com/suhrobdomoiZ/Eda-1/services/auth/internal/repository"
	service "github.com/suhrobdomoiZ/Eda-1/services/auth/internal/services"
)

func main() {
	cfg := config.Load()

	// PostgreSQL
	pgRepo, err := repository.NewPostgresRepo(cfg.Postgres.DSN())
	if err != nil {
		log.Fatalf("postgres connection failed: %v", err)
	}

	// Metrics
	m := metrics.NewMetrics("auth")
	go func() {
		http.Handle("/metrics", metrics.Handler())
		log.Printf("auth metrics server listening on port %s", cfg.Metrics.Port)
		if err := http.ListenAndServe(":"+cfg.Metrics.Port, nil); err != nil {
			log.Fatalf("metrics server failed: %v", err)
		}
	}()

	// Redis
	redisRepo, err := repository.NewRedisRepo(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}

	// JWT Service
	jwtSvc := service.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	// Auth Service
	authSvc := service.NewAuthService(pgRepo, redisRepo, jwtSvc, m)
	authServer := api.NewServer(authSvc)

	// gRPC Server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(m.UnaryServerInterceptor()),
	)
	pb.RegisterAuthServiceServer(grpcServer, authServer)
	reflection.Register(grpcServer)

	addr := fmt.Sprintf(":%s", cfg.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("auth gRPC server listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
