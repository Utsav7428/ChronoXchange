package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/Utsav7428/ChronoXchange/internal/database"
	"github.com/Utsav7428/ChronoXchange/pkg/types"

	"github.com/redis/go-redis/v9"
)

const dbProcessorQueue = "db_processor"

func main() {
	// 1. Initialize Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. Connect to the Database
	// This uses the Connect function and models we created earlier.
	database.Connect()

	// 3. Connect to Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		slog.Error("REDIS_URL not set, using default")
		redisURL = "redis://localhost:6379/0"
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		slog.Error("could not parse redis url", "error", err)
		os.Exit(1)
	}
	redisClient := redis.NewClient(opts)
	ctx := context.Background()

	slog.Info("DB processor started, waiting for messages...")

	// 4. Main Loop: Listen for and process messages indefinitely
	for {
		// Perform a blocking pop from the Redis list.
		// It waits for 0 seconds (forever) until a message is available.
		result, err := redisClient.BRPop(ctx, 0, dbProcessorQueue).Result()
		if err != nil {
			slog.Error("error popping from redis queue", "error", err)
			time.Sleep(1 * time.Second) // Avoid spamming logs on persistent error
			continue
		}

		// result is a slice where result[0] is the queue name and result[1] is the message.
		messageData := []byte(result[1])
		slog.Info("received message from queue", "data", string(messageData))

		// Determine message type and process accordingly.
		var genericMsg types.GenericMessage
		if err := json.Unmarshal(messageData, &genericMsg); err != nil {
			slog.Error("could not unmarshal generic message", "error", err)
			continue
		}

		switch genericMsg.Type {
		case "TRADE_ADDED":
			var msg types.DBTradeMessage
			if err := json.Unmarshal(messageData, &msg); err != nil {
				slog.Error("could not unmarshal trade message", "error", err)
				continue
			}
			handleTradeAdded(msg)

		case "ORDER_UPDATE":
			var msg types.DBOrderMessage
			if err := json.Unmarshal(messageData, &msg); err != nil {
				slog.Error("could not unmarshal order message", "error", err)
				continue
			}
			handleOrderUpdate(msg)

		default:
			slog.Warn("received unknown message type", "type", genericMsg.Type)
		}
	}
}

func handleTradeAdded(msg types.DBTradeMessage) {
	slog.Info("processing TRADE_ADDED message", "trade_id", msg.ID)
	trade := database.Trade{
		ID:            msg.ID,
		IsBuyerMaker:  msg.IsBuyerMaker,
		Price:         msg.Price,
		Quantity:      msg.Quantity,
		QuoteQuantity: msg.QuoteQuantity,
		Timestamp:     time.UnixMilli(msg.Timestamp),
		Market:        msg.Market,
	}

	if result := database.DB.Create(&trade); result.Error != nil {
		slog.Error("failed to create trade in db", "error", result.Error)
	}
}

func handleOrderUpdate(msg types.DBOrderMessage) {
	slog.Info("processing ORDER_UPDATE message", "order_id", msg.OrderID)
	order := database.Order{
		ID:          msg.OrderID,
		ExecutedQty: msg.ExecutedQty,
		Market:      msg.Market,
		Price:       msg.Price,
		Quantity:    msg.Quantity,
		Side:        string(msg.Side),
	}

	// This assumes an order is created on first update.
	// You might want to use `FirstOrCreate` or `Updates` for more complex logic.
	if result := database.DB.Create(&order); result.Error != nil {
		slog.Error("failed to create order in db", "error", result.Error)
	}
}
