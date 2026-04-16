package utils

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrRestaurantIDRequired         = errors.New("restaurant id is required")
	ErrRestaurantIdIsIncorrectValue = errors.New("restaurant id is not a valid UUID")
	ErrNameRequired                 = errors.New("name is required")
	ErrDescriptionTooLong           = errors.New("description exceeds 1024 characters")
	ErrPriceNegative                = errors.New("price cannot be negative")
	ErrProductAlreadyExists         = errors.New("product already exists")
	ErrValidationFailed             = errors.New("product invalid data")
	ErrInvalidRestaurantID          = errors.New("invalid restaurant id")

	ErrProductIDRequired         = errors.New("product id is required")
	ErrProductIdIsIncorrectValue = errors.New("product id is not a valid UUID")
	ErrInvalidProductID          = errors.New("invalid product id")
)

func ToGRPC(err error) error {
	switch {
	case
		errors.Is(err, ErrRestaurantIdIsIncorrectValue),
		errors.Is(err, ErrRestaurantIDRequired),
		errors.Is(err, ErrNameRequired),
		errors.Is(err, ErrDescriptionTooLong),
		errors.Is(err, ErrPriceNegative),
		errors.Is(err, ErrValidationFailed),
		errors.Is(err, ErrInvalidRestaurantID),
		errors.Is(err, ErrProductIdIsIncorrectValue),
		errors.Is(err, ErrProductIDRequired):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, ErrProductAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
