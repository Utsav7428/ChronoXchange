package matching

import (
	"sort"
	"sync"
	"time"

	"github.com/Utsav7428/ChronoXchange/pkg/types"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Orderbook matches buy and sell orders for a single market.
type Orderbook struct {
	mu        sync.RWMutex
	market    string
	bids      map[string]*types.Order // Using order ID as key for easy access
	asks      map[string]*types.Order // Using order ID as key
	bidPrices []decimal.Decimal       // Sorted list of bid prices (high to low)
	askPrices []decimal.Decimal       // Sorted list of ask prices (low to high)
}

// NewOrderbook creates a new orderbook for a given market.
func NewOrderbook(market string) *Orderbook {
	return &Orderbook{
		market:    market,
		bids:      make(map[string]*types.Order),
		asks:      make(map[string]*types.Order),
		bidPrices: make([]decimal.Decimal, 0),
		askPrices: make([]decimal.Decimal, 0),
	}
}

// AddOrder adds a new order to the book and attempts to match it.
func (ob *Orderbook) AddOrder(orderData types.CreateOrderData) []types.Fill {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	order := &types.Order{
		ID:       uuid.New(),
		UserID:   orderData.UserID,
		Side:     orderData.Side,
		Price:    orderData.Price,
		Quantity: orderData.Quantity,
		Filled:   decimal.Zero,
	}

	var fills []types.Fill

	if order.Side == types.Buy {
		fills = ob.matchBid(order)
	} else {
		fills = ob.matchAsk(order)
	}

	// If the order is not fully filled, add it to the book.
	if order.Quantity.GreaterThan(order.Filled) {
		ob.add(order)
	}

	return fills
}

// matchBid attempts to match a buy order (bid) with existing sell orders (asks).
func (ob *Orderbook) matchBid(order *types.Order) []types.Fill {
	fills := make([]types.Fill, 0)

	// Iterate through asks from lowest price to highest
	for i := 0; i < len(ob.askPrices) && order.Filled.LessThan(order.Quantity); {
		askPrice := ob.askPrices[i]
		if order.Price.LessThan(askPrice) {
			break
		}

		ordersAtPrice := ob.getOrdersByPrice(types.Sell, askPrice)
		for _, matchedOrder := range ordersAtPrice {
			if order.Filled.Equal(order.Quantity) {
				break
			}

			qtyToFill := decimal.Min(order.Quantity.Sub(order.Filled), matchedOrder.Quantity.Sub(matchedOrder.Filled))
			order.Filled = order.Filled.Add(qtyToFill)
			matchedOrder.Filled = matchedOrder.Filled.Add(qtyToFill)

			fills = append(fills, types.Fill{
				Qty:           qtyToFill,
				Price:         matchedOrder.Price,
				TradeID:       time.Now().UnixNano(),
				MarketOrderID: matchedOrder.ID,
				OtherUserID:   matchedOrder.UserID,
			})

			if matchedOrder.Filled.Equal(matchedOrder.Quantity) {
				ob.remove(matchedOrder)
			}
		}

		if len(ob.getOrdersByPrice(types.Sell, askPrice)) == 0 {
			ob.removePrice(types.Sell, askPrice)
		} else {
			i++
		}
	}
	return fills
}

// matchAsk attempts to match a sell order (ask) with existing buy orders (bids).
func (ob *Orderbook) matchAsk(order *types.Order) []types.Fill {
	fills := make([]types.Fill, 0)

	// Iterate through bids from highest price to lowest
	for i := 0; i < len(ob.bidPrices) && order.Filled.LessThan(order.Quantity); {
		bidPrice := ob.bidPrices[i]
		if order.Price.GreaterThan(bidPrice) {
			break
		}

		ordersAtPrice := ob.getOrdersByPrice(types.Buy, bidPrice)
		for _, matchedOrder := range ordersAtPrice {
			if order.Filled.Equal(order.Quantity) {
				break
			}
			qtyToFill := decimal.Min(order.Quantity.Sub(order.Filled), matchedOrder.Quantity.Sub(matchedOrder.Filled))
			order.Filled = order.Filled.Add(qtyToFill)
			matchedOrder.Filled = matchedOrder.Filled.Add(qtyToFill)

			fills = append(fills, types.Fill{
				Qty:           qtyToFill,
				Price:         matchedOrder.Price,
				TradeID:       time.Now().UnixNano(),
				MarketOrderID: matchedOrder.ID,
				OtherUserID:   matchedOrder.UserID,
			})

			if matchedOrder.Filled.Equal(matchedOrder.Quantity) {
				ob.remove(matchedOrder)
			}
		}

		if len(ob.getOrdersByPrice(types.Buy, bidPrice)) == 0 {
			ob.removePrice(types.Buy, bidPrice)
		} else {
			i++
		}
	}
	return fills
}

// --- Helper methods for managing the orderbook state ---

func (ob *Orderbook) add(order *types.Order) {
	if order.Side == types.Buy {
		ob.bids[order.ID.String()] = order
		if !ob.hasPrice(types.Buy, order.Price) {
			ob.bidPrices = append(ob.bidPrices, order.Price)
			sort.Slice(ob.bidPrices, func(i, j int) bool {
				return ob.bidPrices[i].GreaterThan(ob.bidPrices[j]) // Sort high to low
			})
		}
	} else {
		ob.asks[order.ID.String()] = order
		if !ob.hasPrice(types.Sell, order.Price) {
			ob.askPrices = append(ob.askPrices, order.Price)
			sort.Slice(ob.askPrices, func(i, j int) bool {
				return ob.askPrices[i].LessThan(ob.askPrices[j]) // Sort low to high
			})
		}
	}
}

func (ob *Orderbook) remove(order *types.Order) {
	if order.Side == types.Buy {
		delete(ob.bids, order.ID.String())
	} else {
		delete(ob.asks, order.ID.String())
	}

	if len(ob.getOrdersByPrice(order.Side, order.Price)) == 0 {
		ob.removePrice(order.Side, order.Price)
	}
}

func (ob *Orderbook) hasPrice(side types.OrderSide, price decimal.Decimal) bool {
	prices := ob.bidPrices
	if side == types.Sell {
		prices = ob.askPrices
	}
	for _, p := range prices {
		if p.Equal(price) {
			return true
		}
	}
	return false
}

func (ob *Orderbook) removePrice(side types.OrderSide, price decimal.Decimal) {
	if side == types.Buy {
		for i, p := range ob.bidPrices {
			if p.Equal(price) {
				ob.bidPrices = append(ob.bidPrices[:i], ob.bidPrices[i+1:]...)
				return
			}
		}
	} else {
		for i, p := range ob.askPrices {
			if p.Equal(price) {
				ob.askPrices = append(ob.askPrices[:i], ob.askPrices[i+1:]...)
				return
			}
		}
	}
}

func (ob *Orderbook) getOrdersByPrice(side types.OrderSide, price decimal.Decimal) []*types.Order {
	orders := make([]*types.Order, 0)
	source := ob.bids
	if side == types.Sell {
		source = ob.asks
	}
	for _, order := range source {
		if order.Price.Equal(price) {
			orders = append(orders, order)
		}
	}
	return orders
}
