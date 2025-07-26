package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/Utsav7428/ChronoXchange/internal/matching"
	"github.com/Utsav7428/ChronoXchange/pkg/types"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

const (
	apiQueue         = "messages"
	dbProcessorQueue = "db_processor"
)

// APIRequestWrapper corresponds to the `MessageWrapper` in the Rust engine.
type APIRequestWrapper struct {
	ClientID string          `json:"client_id"` // The channel to send the API response back on
	UserID   uuid.UUID       `json:"user_id"`
	Message  json.RawMessage `json:"message"` // The actual command payload
}

// APIMessage corresponds to the `MessageFromApi` enum.
type APIMessage struct {
	Type string                `json:"type"`
	Data types.CreateOrderData `json:"data"`
}

func main() {
	// 1. Initialize Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using environment variables")
	}

	// 2. Connect to Redis
	redisURL := os.Getenv("REDIS_URL")
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		slog.Error("could not parse redis url", "error", err)
		os.Exit(1)
	}
	redisClient := redis.NewClient(opts)
	ctx := context.Background()

	// 3. Initialize the Orderbook
	// For now, we hardcode one market like in the Rust example.
	orderbook := matching.NewOrderbook("SOL_USDC")
	slog.Info("Matching engine started", "market", "SOL_USDC")

	// 4. Main Loop: Listen for API commands
	for {
		// Block until a command is available in the 'messages' queue.
		result, err := redisClient.BRPop(ctx, 0, apiQueue).Result()
		if err != nil {
			slog.Error("error popping from api queue", "error", err)
			continue
		}

		// Unmarshal the outer wrapper to get the client_id and the message payload.
		var wrappedReq APIRequestWrapper
		if err := json.Unmarshal([]byte(result[1]), &wrappedReq); err != nil {
			slog.Error("could not unmarshal request wrapper", "error", err)
			continue
		}

		slog.Info("processing request", "client_id", wrappedReq.ClientID, "user_id", wrappedReq.UserID)

		// Unmarshal the inner message to determine the command type.
		var apiMsg APIMessage
		if err := json.Unmarshal(wrappedReq.Message, &apiMsg); err != nil {
			slog.Error("could not unmarshal api message", "error", err)
			continue
		}

		// 5. Process the command
		switch apiMsg.Type {
		case "CREATE_ORDER":
			// The AddOrder method returns the trades (fills) that resulted from the new order.
			fills := orderbook.AddOrder(apiMsg.Data)

			// After processing, publish results to other services.
			publishResults(ctx, redisClient, fills, "SOL_USDC")

			slog.Info("order processed", "fills", len(fills))

			// TODO: Send a confirmation back to the API service on the `wrappedReq.ClientID` channel.

		// TODO: Add cases for CANCEL_ORDER, GET_DEPTH, etc.

		default:
			slog.Warn("received unknown message type", "type", apiMsg.Type)
		}
	}
}

// publishResults sends data to the db-processor and WebSocket services via Redis.
func publishResults(ctx context.Context, rdb *redis.Client, fills []types.Fill, market string) {
	for _, fill := range fills {
		// --- Task 1: Publish to DB Processor ---
		dbTradeMsg := types.DBTradeMessage{
			Type:          "TRADE_ADDED",
			ID:            uuid.New(),
			IsBuyerMaker:  false, // Simplified for now
			Price:         fill.Price.String(),
			Quantity:      fill.Qty.String(),
			QuoteQuantity: fill.Price.Mul(fill.Qty).String(),
			Timestamp:     time.Now().UnixMilli(),
			Market:        market,
		}
		dbPayload, _ := json.Marshal(dbTradeMsg)
		if err := rdb.LPush(ctx, dbProcessorQueue, dbPayload).Err(); err != nil {
			slog.Error("failed to push to db processor queue", "error", err)
		}

		// --- Task 2: Publish to WebSocket Hub ---
		wsTradeData := types.TradeData{
			EventType: "trade",
			TradeID:   fill.TradeID,
			Price:     fill.Price,
			Quantity:  fill.Qty,
			Market:    market,
		}
		wsMsg := types.WsMessage{
			Stream: "trades@" + market,
			Data:   wsTradeData,
		}
		wsPayload, _ := json.Marshal(wsMsg)
		// Publish to a general topic that the WebSocket server will listen to.
		if err := rdb.Publish(ctx, "ws-messages", wsPayload).Err(); err != nil {
			slog.Error("failed to publish to ws topic", "error", err)
		}
	}
}
