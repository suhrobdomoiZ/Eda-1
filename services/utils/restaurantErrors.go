package utils

import "errors"

var (
	ErrRestaurantIDRequired = errors.New("restaurant_id is required")
	ErrRestaurantIDInvalid  = errors.New("restaurant_id is not a valid UUID")
	ErrNameRequired         = errors.New("name is required")
	ErrDescriptionTooLong   = errors.New("description exceeds 1024 characters")
	ErrPriceNegative        = errors.New("price cannot be negative")
)
