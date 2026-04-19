package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	commonpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	authpb "github.com/suhrobdomoiZ/Eda-1/services/api"
	"github.com/suhrobdomoiZ/Eda-1/services/api_gateway/internal/handlers"
	"github.com/suhrobdomoiZ/Eda-1/services/api_gateway/internal/middleware"
)

func New(
	authClient authpb.AuthServiceClient,
	restaurantHandler *handlers.RestaurantHandler,
	customerHandler *handlers.CustomerHandler,
	courierHandler *handlers.CourierHandler,
) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "${method} ${path} ${status} ${latency}\n",
	}))

	authHandler := handlers.NewAuthHandler(authClient)
	authMW := middleware.Auth(authClient)

	// Auth
	auth := app.Group("/api/v1/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)
	auth.Get("/profile", authMW, authHandler.Profile)

	// Restaurant
	rest := app.Group("/api/v1/restaurant")
	restMW := middleware.RequireRole(commonpb.UserRole_USER_ROLE_RESTAURANT)

	// Public
	rest.Get("/menu/:restaurant_id", restaurantHandler.ListProducts)
	rest.Get("/menu/:restaurant_id/product/:id", restaurantHandler.GetProduct)

	// Only restaurant
	rest.Post("/menu", authMW, restMW, restaurantHandler.AddProduct)
	rest.Put("/menu/:id", authMW, restMW, restaurantHandler.UpdateProduct)
	rest.Delete("/menu/:id", authMW, restMW, restaurantHandler.DeleteProduct)
	rest.Get("/orders", authMW, restMW, restaurantHandler.ListOrders)
	rest.Put("/orders/:id/status", authMW, restMW, restaurantHandler.ChangeOrderStatus)

	// Customer
	cust := app.Group("/api/v1/customer")
	custMW := middleware.RequireRole(commonpb.UserRole_USER_ROLE_CUSTOMER)

	// Public
	cust.Get("/restaurants", customerHandler.ListRestaurants)
	cust.Get("/restaurants/:restaurant_id/menu", customerHandler.GetRestaurantMenu)

	// Only client
	cust.Post("/orders", authMW, custMW, customerHandler.CreateOrder)
	cust.Get("/orders", authMW, custMW, customerHandler.ListMyOrders)
	cust.Get("/orders/:order_id", authMW, custMW, customerHandler.GetOrder)
	cust.Delete("/orders/:order_id", authMW, custMW, customerHandler.CancelOrder)

	// Courier
	cour := app.Group("/api/v1/courier")
	courMW := middleware.RequireRole(commonpb.UserRole_USER_ROLE_COURIER)

	cour.Get("/orders/available", authMW, courMW, courierHandler.GetAvailableOrders)
	cour.Post("/orders/:order_id/accept", authMW, courMW, courierHandler.AcceptOrder)
	cour.Get("/orders", authMW, courMW, courierHandler.GetMyOrders)
	cour.Post("/orders/:order_id/pickup", authMW, courMW, courierHandler.PickUpOrder)
	cour.Post("/orders/:order_id/deliver", authMW, courMW, courierHandler.DeliverOrder)

	return app
}
