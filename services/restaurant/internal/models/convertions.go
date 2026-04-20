package models

import (
	"github.com/google/uuid"
	"github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
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

func ConvertListProductsRequestToRestaurantId(recent *api.ListProductsRequest) (*RestaurantId, error) {
	stringRestaurantId := recent.RestaurantId
	if stringRestaurantId == "" {
		return nil, utils.ErrRestaurantIDRequired
	}

	restaurantId, err := uuid.Parse(recent.RestaurantId)
	if err != nil {
		return nil, utils.ErrRestaurantIdIsIncorrectValue
	}

	return &RestaurantId{Id: restaurantId}, nil
}

func ConvertSliceOfProductsToListProductsResponse(products []FullProduct) *api.ListProductsResponse {
	pbProducts := make([]*api.FullProduct, len(products))
	for i, p := range products {
		pbProducts[i] = &api.FullProduct{
			Id: p.Id.String(),
			Info: &api.ProductInfo{
				RestaurantId: p.RestaurantId.String(),
				Name:         p.Name,
				Description:  p.Description,
				Price:        p.Price,
			},
		}
	}

	return &api.ListProductsResponse{
		Status:   utils.StatusOK,
		Products: pbProducts,
	}
}

func ConvertGetProductRequestToProductID(recent *api.GetProductRequest) (*ProductId, error) {
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

func ConvertFullProductToGetProductResponse(product *FullProduct) *api.GetProductResponse {
	return &api.GetProductResponse{
		Status: utils.StatusOK,
		Product: &api.FullProduct{
			Id: product.Id.String(),
			Info: &api.ProductInfo{
				RestaurantId: product.RestaurantId.String(),
				Name:         product.Name,
				Description:  product.Description,
				Price:        product.Price,
			},
		},
	}
}

func ConvertChangeOrderStatusRequestToOrderIDWithStatus(recent *api.ChangeOrderStatusRequest) (*OrderIdWithStatus, error) {
	stringId := recent.Id
	if stringId == "" {
		return nil, utils.ErrProductIDRequired
	}

	orderId, err := uuid.Parse(stringId)
	if err != nil {
		return nil, utils.ErrProductIdIsIncorrectValue
	}

	rawStatus := recent.Status
	if !IsValidOrderStatus(rawStatus) {
		return nil, utils.ErrInvalidOrderStatus
	}

	return &OrderIdWithStatus{OrderId: orderId, Status: rawStatus}, nil
}

func ConvertCommonOrderStatusToDBStatus(recent common.OrderStatus) (DBOrderStatus, error) {
	switch recent {
	case common.OrderStatus_ORDER_STATUS_CREATED:
		return "created", nil
	case common.OrderStatus_ORDER_STATUS_COOKING:
		return "cooking", nil
	case common.OrderStatus_ORDER_STATUS_READY:
		return "ready", nil
	case common.OrderStatus_ORDER_STATUS_DELIVERING:
		return "delivering", nil
	case common.OrderStatus_ORDER_STATUS_DELIVERED:
		return "delivered", nil
	case common.OrderStatus_ORDER_STATUS_CANCELED:
		return "cancelled", nil
	default:
		return "", utils.ErrInvalidOrderStatus
	}
}

func ConvertDBStatusToCommonOrderStatus(s string) common.OrderStatus {
	switch s {
	case "cooking":
		return common.OrderStatus_ORDER_STATUS_COOKING
	case "ready":
		return common.OrderStatus_ORDER_STATUS_READY
	case "delivering":
		return common.OrderStatus_ORDER_STATUS_DELIVERING
	case "delivered":
		return common.OrderStatus_ORDER_STATUS_DELIVERED
	case "cancelled":
		return common.OrderStatus_ORDER_STATUS_CANCELED
	default:
		return common.OrderStatus_ORDER_STATUS_CREATED
	}
}

func ConvertChangedOrderIdToChangeOrderStatusResponse(changedOrderId *ChangedOrderId) *api.ChangeOrderStatusResponse {
	return &api.ChangeOrderStatusResponse{
		Id:     changedOrderId.OrderId.String(),
		Status: utils.StatusOK,
	}
}

func ConvertListOrdersRequestToRestaurantId(recent *api.ListOrdersRequest) (*RestaurantId, error) {
	stringId := recent.Id
	if stringId == "" {
		return nil, utils.ErrRestaurantIDRequired
	}
	restaurantId, err := uuid.Parse(stringId)

	if err != nil {
		return nil, utils.ErrRestaurantIdIsIncorrectValue
	}

	return &RestaurantId{Id: restaurantId}, nil
}

func ConvertOrderedProductToCommonOrderedItem(orderedProduct OrderedProduct) *common.OrderItem {
	return &common.OrderItem{
		ProductId: orderedProduct.ProductId.String(),
		Quantity:  orderedProduct.Quantity,
		Price:     orderedProduct.Price,
		Name:      orderedProduct.Name,
	}
}

func ConvertOrderToCommonOrder(order *Order) *common.Order {

	var orderItems []*common.OrderItem

	for _, item := range order.OrderedItems {
		orderItems = append(orderItems, ConvertOrderedProductToCommonOrderedItem(item))
	}

	return &common.Order{
		Id:           order.Id.String(),
		RestaurantId: order.RestaurantId.String(),
		CourierId:    order.CourierId.String(),
		ClientId:     order.ClientId.String(),
		TotalPrice:   order.TotalPrice,
		Status:       order.OrderStatus,
		Address:      order.Address,
		Items:        orderItems,
	}
}

func ConvertSliceOfOrdersToListOrdersResponse(array []Order) *api.ListOrdersResponse {
	var orders []*common.Order

	for _, order := range array {
		orders = append(orders, ConvertOrderToCommonOrder(&order))
	}

	return &api.ListOrdersResponse{
		Status: utils.StatusOK,
		Orders: orders,
	}
}
