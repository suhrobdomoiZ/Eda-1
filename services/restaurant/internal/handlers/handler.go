package handlers

import (
	"context"

	"github.com/suhrobdomoiZ/Eda-1/services/api"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/models"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/service"
	"github.com/suhrobdomoiZ/Eda-1/services/utils"
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
		return nil, utils.ToGRPC(err)
	}

	result, err := r.svc.AddProduct(ctx, productInfo)
	if err != nil {
		return nil, utils.ToGRPC(err)
	}

	return models.ConvertUUIDToAddProductResponse(result), nil
}

func (r *Restaurant) UpdateProduct(ctx context.Context, request *api.UpdateProductRequest) (*api.UpdateProductResponse, error) {
	product, err := models.ConvertUpdateProductRequestToFullProduct(request)
	if err != nil {
		return nil, utils.ToGRPC(err)
	}
	result, err := r.svc.UpdateProduct(ctx, product)
	if err != nil {
		return nil, utils.ToGRPC(err)
	}
	return models.ConvertUUIDTOUpdateProductResponse(result), nil
}

func (r *Restaurant) DeleteProduct(ctx context.Context, request *api.DeleteProductRequest) (*api.DeleteProductResponse, error) {
	productId, err := models.ConvertDeleteProductRequestToUUID(request)
	if err != nil {
		return nil, utils.ToGRPC(err)
	}

	err = r.svc.DeleteProduct(ctx, productId)
	if err != nil {
		return nil, utils.ToGRPC(err)
	}

	return models.ConvertStatusToDeleteProductResponse(), nil
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
