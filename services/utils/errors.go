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
	ErrProductHasDependencies    = errors.New("product has dependencies")
	ErrProductDoesNotFound       = errors.New("product does not exist")

	ErrInvalidDataSchema = errors.New("invalid data schema")
	ErrConnectionFailure = errors.New("connection failure")

	ErrOrderIDRequired    = errors.New("order id is required")
	ErrInvalidOrderID     = errors.New("invalid order id")
	ErrOrderDoesNotFound  = errors.New("order does not exist")
	ErrInvalidOrderStatus = errors.New("invalid order status")
	ErrNotAllowed         = errors.New("not allowed to change status of delivering,delivered and cancelled order")
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
		errors.Is(err, ErrProductIDRequired),
		errors.Is(err, ErrInvalidDataSchema),
		errors.Is(err, ErrOrderIDRequired),
		errors.Is(err, ErrInvalidOrderStatus),
		errors.Is(err, ErrInvalidOrderID):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, ErrProductAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, ErrProductHasDependencies):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, ErrNotAllowed):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, ErrProductDoesNotFound),
		errors.Is(err, ErrOrderDoesNotFound):
		return status.Error(codes.NotFound, err.Error())

	default:
		return status.Error(codes.Internal, err.Error())
	}
}
