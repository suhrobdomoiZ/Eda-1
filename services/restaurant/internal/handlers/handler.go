package handlers

import (
	"context"

	pb "github.com/suhrobdomoiZ/Eda-1/pkg/api/restaurant"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/models"
	"github.com/suhrobdomoiZ/Eda-1/services/restaurant/internal/service"
	"github.com/suhrobdomoiZ/Eda-1/services/utils"
)

type Restaurant struct {
	svc *service.Restaurant
	pb.UnimplementedRestaurantServer
}

func NewRestaurant(restaurant *service.Restaurant) *Restaurant {
	return &Restaurant{svc: restaurant}
}

func (r *Restaurant) AddProduct(ctx context.Context, request *pb.AddProductRequest) (*pb.AddProductResponse, error) {
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

func (r *Restaurant) UpdateProduct(ctx context.Context, request *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
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

func (r *Restaurant) DeleteProduct(ctx context.Context, request *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
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

func (r *Restaurant) ListProducts(ctx context.Context, request *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
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

func (r *Restaurant) GetProduct(ctx context.Context, request *pb.GetProductRequest) (*pb.GetProductResponse, error) {
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

func (r *Restaurant) ChangeOrderStatus(ctx context.Context, request *pb.ChangeOrderStatusRequest) (*pb.ChangeOrderStatusResponse, error) {
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

func (r *Restaurant) ListOrders(ctx context.Context, request *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	restaurantId, err := models.ConvertListOrdersRequestToRestaurantId(request)
	if err != nil {
		return nil, utils.ToGRPC(err)
	}

	result, err := r.svc.ListOrders(ctx, restaurantId)
	if err != nil {
		return nil, utils.ToGRPC(err)
	}

	return models.ConvertSliceOfOrdersToListOrdersResponse(result), nil
}

func (r *Restaurant) ListRestaurants(ctx context.Context, request *pb.ListRestaurantsRequest) (*pb.ListRestaurantsResponse, error) {
	limit, offset := models.ConvertListRestaurantsRequestToParams(request)

	restaurants, total, err := r.svc.ListRestaurants(ctx, limit, offset)
	if err != nil {
		return nil, utils.ToGRPC(err)
	}

	return models.ConvertRestaurantsToResponse(restaurants, total), nil
}
