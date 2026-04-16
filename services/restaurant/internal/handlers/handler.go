package handlers

import (
	"context"
	"fmt"

	"github.com/suhrobdomoiZ/Eda-1/services/api"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/models"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Restaurant struct {
	svc service.Restaurant
	api.UnimplementedRestaurantServer
}

func NewRestaurant(restaurant service.Restaurant) *Restaurant {
	return &Restaurant{svc: restaurant}
}

func (r *Restaurant) AddProduct(ctx context.Context, request *api.AddProductRequest) (*api.AddProductResponse, error) {

	productInfo, err := models.ConvertAddProductRequestToProductInfo(request)
	if err != nil {
		return nil, fmt.Errorf("handlers.AddProduct: %w", err)
	}

	return nil, status.Error(codes.Unimplemented, "")
}

func (r *Restaurant) UpdateProduct(context.Context, *api.UpdateProductRequest) (*api.UpdateProductResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (r *Restaurant) DeleteProduct(context.Context, *api.DeleteProductRequest) (*api.DeleteProductResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (r *Restaurant) ListProducts(context.Context, *api.ListProductsRequest) (*api.ListProductsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (r *Restaurant) GetProduct(context.Context, *api.GetProductRequest) (*api.GetProductResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (r *Restaurant) ChangeOrderStatus(context.Context, *api.ChangeOrderStatusRequest) (*api.ChangeOrderStatusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (r *Restaurant) ListOrders(context.Context, *api.ListOrdersRequest) (*api.ListOrdersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
