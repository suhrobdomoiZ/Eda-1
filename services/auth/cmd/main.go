package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/suhrobdomoiZ/Eda-1/services/auth/internal/api/gen"
	api "github.com/suhrobdomoiZ/Eda-1/services/auth/internal/api/server"
	"github.com/suhrobdomoiZ/Eda-1/services/auth/internal/config"
	"github.com/suhrobdomoiZ/Eda-1/services/auth/internal/repository"
	service "github.com/suhrobdomoiZ/Eda-1/services/auth/internal/services"
)

func main() {
	cfg := config.Load()

	pgRepo, err := repository.NewPostgresRepo(cfg.Postgres.DSN())
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	redisRepo, err := repository.NewRedisRepo(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	jwtSvc := service.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	authSvc := service.NewAuthService(pgRepo, redisRepo, jwtSvc)
	authServer := api.NewServer(authSvc)

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authServer)
	reflection.Register(grpcServer)

	addr := fmt.Sprintf(":%s", cfg.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	log.Printf("auth gRPC server listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
