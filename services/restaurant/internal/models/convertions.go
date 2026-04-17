package models

import (
	"github.com/google/uuid"
	"github.com/suhrobdomoiZ/Eda-1/services/api"
	"github.com/suhrobdomoiZ/Eda-1/services/utils"
)

func ConvertAddProductRequestToProductInfo(recent *api.AddProductRequest) (*ProductInfo, error) {
	stringId := recent.ProductInfo.RestaurantId
	if stringId == "" {
		return nil, utils.ErrRestaurantIDRequired
	}
	restaurantId, err := uuid.Parse(stringId)
	if err != nil {
		return nil, utils.ErrRestaurantIdIsIncorrectValue
	}

	name := recent.ProductInfo.Name
	if name == "" {
		return nil, utils.ErrNameRequired
	}

	description := recent.ProductInfo.Description //TODO: подумать над пустым description
	if len([]rune(description)) > 1024 {
		return nil, utils.ErrDescriptionTooLong
	}

	price := recent.ProductInfo.Price
	if price < 0 {
		return nil, utils.ErrPriceNegative
	}

	return &ProductInfo{
		RestaurantId: restaurantId,
		Name:         name,
		Description:  description,
		Price:        price,
	}, nil
}

func ConvertUUIDToAddProductResponse(uuid uuid.UUID) *api.AddProductResponse {
	return &api.AddProductResponse{
		Id:     uuid.String(),
		Status: utils.StatusOK,
	}
}

func ConvertUpdateProductRequestToFullProduct(recent *api.UpdateProductRequest) (*FullProduct, error) {
	stringProductId := recent.Id
	if stringProductId == "" {
		return nil, utils.ErrProductIDRequired
	}
	productId, err := uuid.Parse(stringProductId)
	if err != nil {
		return nil, utils.ErrProductIdIsIncorrectValue
	}

	stringRestaurantId := recent.ProductInfo.RestaurantId
	if stringRestaurantId == "" {
		return nil, utils.ErrRestaurantIDRequired
	}
	restaurantId, err := uuid.Parse(stringRestaurantId)
	if err != nil {
		return nil, utils.ErrRestaurantIdIsIncorrectValue
	}

	name := recent.ProductInfo.Name
	if name == "" {
		return nil, utils.ErrNameRequired
	}

	description := recent.ProductInfo.Description //TODO: подумать над пустым description
	if len([]rune(description)) > 1024 {
		return nil, utils.ErrDescriptionTooLong
	}

	price := recent.ProductInfo.Price
	if price < 0 {
		return nil, utils.ErrPriceNegative
	}

	return &FullProduct{
		Id:           productId,
		RestaurantId: restaurantId,
		Name:         name,
		Description:  description,
		Price:        price,
	}, nil
}

func ConvertUUIDTOUpdateProductResponse(uuid uuid.UUID) *api.UpdateProductResponse {
	return &api.UpdateProductResponse{
		Id:     uuid.String(),
		Status: utils.StatusOK,
	}
}

func ConvertDeleteProductRequestToUUID(recent *api.DeleteProductRequest) (*ProductId, error) {
	stringId := recent.Id
	if stringId == "" {
		return nil, utils.ErrProductIDRequired
	}

	productId, err := uuid.Parse(stringId)
	if err != nil {
		return nil, utils.ErrProductIdIsIncorrectValue
	}
	return &ProductId{Id: productId}, nil
}

func ConvertStatusToDeleteProductResponse() *api.DeleteProductResponse {
	return &api.DeleteProductResponse{
		Status: utils.StatusOK,
	}
}
