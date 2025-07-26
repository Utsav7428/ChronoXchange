package database

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is a global variable to hold the database connection pool.
var DB *gorm.DB

// Connect initializes the connection to the database and runs migrations.
func Connect() {
	// Load the .env file during development.
	// In production, environment variables are typically set directly.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Read the database URL from the environment.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	// Open a connection to the database.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	log.Println("Database connection successfully established.")

	// Run database migrations to create the tables.
	// This ensures your schema is up-to-date every time the app starts.
	err = db.AutoMigrate(&User{}, &Order{}, &Trade{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	log.Println("Database successfully migrated.")

	// Assign the connected database instance to the global variable.
	DB = db
}