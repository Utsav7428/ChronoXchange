package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/Utsav7428/ChronoXchange/internal/hub"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity
	},
}

func serveWs(h *hub.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &hub.Client{Hub: h, Conn: conn, Send: make(chan []byte, 256)}
	client.Hub.Register <- client
	go client.WritePump()
	go client.ReadPump()
}

// listenToRedis subscribes to the Redis topic and forwards messages to the hub.
func listenToRedis(ctx context.Context, h *hub.Hub) {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found")
	}
	redisURL := os.Getenv("REDIS_URL")
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		slog.Error("could not parse redis url", "error", err)
		return
	}
	redisClient := redis.NewClient(opts)

	slog.Info("Subscribing to ws-messages topic on Redis")
	pubsub := redisClient.Subscribe(ctx, "ws-messages")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		// When a message is received from Redis, send it to the hub's broadcast channel.
		slog.Info("received message from redis, broadcasting to clients", "msg", msg.Payload)
		h.Broadcast <- []byte(msg.Payload)
	}
}

func main() {
	h := hub.NewHub()
	ctx := context.Background()

	go h.Run()
	go listenToRedis(ctx, h) // Start the Redis listener

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(h, w, r)
	})

	log.Println("WebSocket server starting on :8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
