package main

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/auth"
	courpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/courier"
	custompb "github.com/suhrobdomoiZ/Eda-1/pkg/api/customer"
	respb "github.com/suhrobdomoiZ/Eda-1/pkg/api/restaurant"

	"github.com/suhrobdomoiZ/Eda-1/services/api_gateway/internal/config"
	"github.com/suhrobdomoiZ/Eda-1/services/api_gateway/internal/handlers"
	"github.com/suhrobdomoiZ/Eda-1/services/api_gateway/internal/router"
)

func main() {
	cfg := config.Load()

	authConn := mustDial(cfg.Services.AuthAddr)
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)

	// Restaurant, Customer, Courier - опциональные (сервисы ещё не запущены)
	var (
		restaurantHandler *handlers.RestaurantHandler
		customerHandler   *handlers.CustomerHandler
		courierHandler    *handlers.CourierHandler
	)

	if cfg.Services.RestaurantAddr != "" {
		conn := mustDial(cfg.Services.RestaurantAddr)
		defer conn.Close()
		restaurantHandler = handlers.NewRestaurantHandler(respb.NewRestaurantClient(conn))
	}

	if cfg.Services.CustomerAddr != "" {
		conn := mustDial(cfg.Services.CustomerAddr)
		defer conn.Close()
		customerHandler = handlers.NewCustomerHandler(custompb.NewCustomerAPIClient(conn))
	}

	if cfg.Services.CourierAddr != "" {
		conn := mustDial(cfg.Services.CourierAddr)
		defer conn.Close()
		courierHandler = handlers.NewCourierHandler(courpb.NewCourierAPIClient(conn))
	}

	app := router.New(authClient, restaurantHandler, customerHandler, courierHandler)

	log.Printf("api gateway listening on %s", cfg.HTTP.Addr())
	if err := app.Listen(cfg.HTTP.Addr()); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func mustDial(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial %s: %v", addr, err)
	}
	return conn
}
