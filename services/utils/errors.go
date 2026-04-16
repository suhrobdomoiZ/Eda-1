package utils

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrRestaurantIDRequired       = errors.New("restaurant_id is required")
	ErrRestaurantIdIncorrectValue = errors.New("restaurant_id is not a valid UUID")
	ErrNameRequired               = errors.New("name is required")
	ErrDescriptionTooLong         = errors.New("description exceeds 1024 characters")
	ErrPriceNegative              = errors.New("price cannot be negative")
	ErrProductAlreadyExists       = errors.New("product already exists")
	ErrValidationFailed           = errors.New("product invalid data")
	ErrInvalidRestaurantID        = errors.New("invalid restaurant id")
)

func ToGRPC(err error) error {
	switch {
	case
		errors.Is(err, ErrRestaurantIdIncorrectValue),
		errors.Is(err, ErrRestaurantIDRequired),
		errors.Is(err, ErrNameRequired),
		errors.Is(err, ErrDescriptionTooLong),
		errors.Is(err, ErrPriceNegative),
		errors.Is(err, ErrValidationFailed),
		errors.Is(err, ErrInvalidRestaurantID):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, ErrProductAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
