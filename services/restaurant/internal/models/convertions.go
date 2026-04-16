package models

import (
	"github.com/google/uuid"
	"github.com/suhrobdomoiZ/Eda-1/services/api"
	"github.com/suhrobdomoiZ/Eda-1/services/utils"
)

func ConvertAddProductRequestToProductInfo(recent *api.AddProductRequest) (*ProductInfo, error) {
	stringId := recent.ProductInfo.RestaurantId
	if stringId == "" {
		return nil, utils.ErrRestaurantIDInvalid
	}
	restaurantId, err := uuid.Parse(stringId)
	if err != nil {
		return nil, utils.ErrRestaurantIDInvalid
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
