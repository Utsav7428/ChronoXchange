package api

import (
	"net/http"
	"os"
	"time"

	"github.com/Utsav7428/ChronoXchange/internal/database"

	"context"
	"encoding/json"
	"log/slog"

	"github.com/Utsav7428/ChronoXchange/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

type signupRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func Signup(c *gin.Context) {
	var req signupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := database.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if result := database.DB.Create(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user database.User
	if result := database.DB.Where("email = ?", req.Email).First(&user); result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Create JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	jwtSecret := os.Getenv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

type createOrderRequest struct {
	Market   string          `json:"market" binding:"required"`
	Price    decimal.Decimal `json:"price" binding:"required"`
	Quantity decimal.Decimal `json:"quantity" binding:"required"`
	Side     types.OrderSide `json:"side" binding:"required"`
}

// APIRequestWrapper is the message format sent to the engine's queue.
type APIRequestWrapper struct {
	ClientID string          `json:"client_id"`
	UserID   uuid.UUID       `json:"user_id"`
	Message  json.RawMessage `json:"message"`
}

// APIMessage is the inner message payload.
type APIMessage struct {
	Type string                `json:"type"`
	Data types.CreateOrderData `json:"data"`
}

func CreateOrder(c *gin.Context) {
	// 1. Get UserID from middleware and parse the request body
	userID, _ := c.Get("userID")

	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Connect to Redis
	redisURL := os.Getenv("REDIS_URL")
	opts, _ := redis.ParseURL(redisURL)
	redisClient := redis.NewClient(opts)
	ctx := context.Background()

	// 3. Prepare the command for the engine
	orderData := types.CreateOrderData{
		UserID:   userID.(uuid.UUID),
		Market:   req.Market,
		Price:    req.Price,
		Quantity: req.Quantity,
		Side:     req.Side,
	}
	apiMsg := APIMessage{Type: "CREATE_ORDER", Data: orderData}
	messagePayload, _ := json.Marshal(apiMsg)

	// This is the unique channel the API will listen on for a response.
	responseChannel := uuid.New().String()

	wrappedReq := APIRequestWrapper{
		ClientID: responseChannel,
		UserID:   userID.(uuid.UUID),
		Message:  messagePayload,
	}

	wrappedPayload, _ := json.Marshal(wrappedReq)

	// 4. Subscribe to the response channel BEFORE sending the command
	pubsub := redisClient.Subscribe(ctx, responseChannel)
	defer pubsub.Close()

	// 5. Push the command to the engine's queue
	if err := redisClient.LPush(ctx, "messages", wrappedPayload).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send order to engine"})
		return
	}

	// 6. Wait for a response from the engine (with a timeout)
	slog.Info("waiting for response on channel", "channel", responseChannel)
	// In a real app, you would receive the engine's response here.
	// For now, we'll just send a success message.
	// _, err := pubsub.ReceiveMessage(ctx)
	// if err != nil { ... }

	c.JSON(http.StatusOK, gin.H{"message": "Order submitted successfully"})
}
