package database

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// This file maps Go structs to your PostgreSQL tables using GORM tags.

// User maps to the "users" table.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Username     string    `gorm:"unique"`
	Email        string    `gorm:"unique"`
	PasswordHash string
	CreatedAt    time.Time `gorm:"not null;default:current_timestamp"`
	UpdatedAt    time.Time `gorm:"not null;default:current_timestamp"`
}

// Order maps to the "orders" table.
type Order struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExecutedQty decimal.Decimal `gorm:"type:numeric"`
	Market      string
	Price       string
	Quantity    string
	Side        string
	CreatedAt   time.Time `gorm:"not null;default:current_timestamp"`
}

// Trade maps to the "trades" table.
type Trade struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	IsBuyerMaker  bool
	Price         string
	Quantity      string
	QuoteQuantity string
	Timestamp     time.Time `gorm:"not null;default:current_timestamp"`
	Market        string
}
