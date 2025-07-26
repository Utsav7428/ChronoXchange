package types

import "github.com/shopspring/decimal"

// WsMessage is the standard wrapper for all messages sent to clients.
type WsMessage struct {
	Stream string      `json:"stream"` // e.g., "trades@SOL_USDC"
	Data   interface{} `json:"data"`
}

// TradeData is the payload for a trade update.
type TradeData struct {
	EventType string          `json:"e"` // "trade"
	TradeID   int64           `json:"t"`
	Price     decimal.Decimal `json:"p"`
	Quantity  decimal.Decimal `json:"q"`
	Market    string          `json:"s"`
}
