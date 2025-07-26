package types

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// This file defines the messages that are sent to the db-processor queue.

// GenericMessage is used to unmarshal the message first to find its type.
type GenericMessage struct {
	Type string `json:"type"`
}

// DBTradeMessage is the payload for a new trade to be saved.
type DBTradeMessage struct {
	Type          string    `json:"type"`
	ID            uuid.UUID `json:"id"`
	IsBuyerMaker  bool      `json:"is_buyer_maker"`
	Price         string    `json:"price"`
	Quantity      string    `json:"quantity"`
	QuoteQuantity string    `json:"quote_quantity"`
	Timestamp     int64     `json:"timestamp"` // Unix milliseconds
	Market        string    `json:"market"`
}

// DBOrderMessage is the payload for an order update to be saved.
type DBOrderMessage struct {
	OrderID     uuid.UUID       `json:"order_id"`
	ExecutedQty decimal.Decimal `json:"executed_qty"`
	Market      string          `json:"market"`
	Price       string          `json:"price"`
	Quantity    string          `json:"quantity"`
	Side        OrderSide       `json:"side"`
}
