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

func (r *Restaurant) ListProducts(ctx context.Context, request *api.ListProductsRequest) (*api.ListProductsResponse, error) {
	restaurantId, err := models.ConvertListProductsRequestToRestaurantId(request)
	if err != nil {
		return nil, utils.ToGRPC(err)
	}
	result, err := r.svc.ListProducts(ctx, restaurantId)

	if err != nil {
		return nil, utils.ToGRPC(err)
	}

	return models.ConvertSliceOfProductsToListProductsResponse(result), nil
}

func (r *Restaurant) GetProduct(ctx context.Context, request *api.GetProductRequest) (*api.GetProductResponse, error) {
	productId, err := models.ConvertGetProductRequestToProductID(request)
	if err != nil {
		return nil, utils.ToGRPC(err)
	}
	result, err := r.svc.GetProduct(ctx, productId)
	if err != nil {
		return nil, utils.ToGRPC(err)
	}

	return models.ConvertFullProductToGetProductResponse(result), err
}

func (r *Restaurant) ChangeOrderStatus(ctx context.Context, request *api.ChangeOrderStatusRequest) (*api.ChangeOrderStatusResponse, error) {
	order, err := models.ConvertChangeOrderStatusRequestToOrderIDWithStatus(request)
	if err != nil {
		return nil, utils.ToGRPC(err)
	}
	result, err := r.svc.ChangeOrderStatus(ctx, order)
	if err != nil {
		return nil, utils.ToGRPC(err)
	}
	
	return models.ConvertChangedOrderIdToChangeOrderStatusResponse(result), nil
}

func (r *Restaurant) ListOrders(context.Context, *api.ListOrdersRequest) (*api.ListOrdersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
