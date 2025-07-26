package types

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// OrderSide defines if the order is a buy or sell.
type OrderSide string

// We define constants for buy and sell sides to avoid typos.
const (
	Buy  OrderSide = "buy"
	Sell OrderSide = "sell"
)

// CreateOrderData is the payload sent from the API to the engine to create an order.
type CreateOrderData struct {
	UserID   uuid.UUID       `json:"user_id"`
	Market   string          `json:"market"`
	Price    decimal.Decimal `json:"price"`
	Quantity decimal.Decimal `json:"quantity"`
	Side     OrderSide       `json:"side"`
}

// CancelOrderData is the payload sent from the API to the engine to cancel an order.
type CancelOrderData struct {
	UserID  uuid.UUID `json:"user_id"`
	OrderID uuid.UUID `json:"order_id"`
	Market  string    `json:"market"`
}

// GetDepthData is the payload sent from the API to the engine to get order book depth.
type GetDepthData struct {
	Market string `json:"market"`
	Limit  int    `json:"limit"` // Number of price levels to return
}

// APIResponse represents a response sent back to the API service.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// CreateOrderResponse is the response for a CREATE_ORDER request.
type CreateOrderResponse struct {
	OrderID uuid.UUID `json:"order_id"`
	Fills   []Fill    `json:"fills"`
}

// CancelOrderResponse is the response for a CANCEL_ORDER request.
type CancelOrderResponse struct {
	OrderID uuid.UUID `json:"order_id"`
	Success bool      `json:"success"`
}

// GetDepthResponse is the response for a GET_DEPTH request.
type GetDepthResponse struct {
	Depth DepthPayload `json:"depth"`
}

// Order represents a single order in the live order book within the matching engine.
type Order struct {
	ID       uuid.UUID       `json:"id"`
	UserID   uuid.UUID       `json:"user_id"`
	Side     OrderSide       `json:"side"`
	Price    decimal.Decimal `json:"price"`
	Quantity decimal.Decimal `json:"quantity"`
	Filled   decimal.Decimal `json:"filled"`
}

// Fill represents a single matched trade execution.
type Fill struct {
	Qty           decimal.Decimal `json:"qty"`
	Price         decimal.Decimal `json:"price"`
	TradeID       int64           `json:"trade_id"`
	MarketOrderID uuid.UUID       `json:"market_order_id"`
	OtherUserID   uuid.UUID       `json:"other_user_id"`
}

// DepthPayload represents the state of the order book for a given market.
type DepthPayload struct {
	Market string               `json:"market"`
	Bids   [][2]decimal.Decimal `json:"bids"` // Each entry is [price, quantity]
	Asks   [][2]decimal.Decimal `json:"asks"` // Each entry is [price, quantity]
}
