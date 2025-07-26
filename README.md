# ChronoXchange

This is a high-performance cryptocurrency exchange matching engine built entirely in Go. It's designed with a microservices architecture to be scalable, concurrent, and maintainable.

---

## 🏛️ Architecture

The system is composed of four main microservices that communicate asynchronously via a Redis message broker.

* **API Server**: The main entry point for users. A REST API that handles user authentication, login, and order submission.
* **Matching Engine**: The core of the system. It hosts the orderbook, processes incoming orders, matches trades, and publishes results.
* **DB Processor**: A dedicated service that listens for trade and order data from the engine and persists it to the PostgreSQL database.
* **WebSocket Server**: Provides real-time updates (like new trades) to connected clients.

### Technology Stack

* **Language**: Go
* **Services**: PostgreSQL, Redis
* **Containerization**: Docker & Docker Compose
* **API Framework**: Gin
* **Database ORM**: GORM
* **Real-time Communication**: Gorilla WebSocket

---

## ✨ Features

* Full user authentication with JWT.
* Order submission (Create Order).
* A real-time matching engine with a price-time priority orderbook.
* Asynchronous data persistence.
* Real-time trade updates via WebSockets.

---

## 🚀 Getting Started

### Prerequisites

* [Go (1.21+)](https://go.dev/dl/) installed.
* [Docker](https://www.docker.com/products/docker-desktop/) installed and running.

### Installation

1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/your-username/ChronoXchange.git](https://github.com/your-username/ChronoXchange.git)
    cd ChronoXchange
    ```

2.  **Create the environment file:**
    Create a `.env` file in the root of the project and add your configuration.
    ```env
    DATABASE_URL="host=localhost user=your_user password=your_password dbname=exchange port=5432 sslmode=disable"
    REDIS_URL="redis://localhost:6379/0"
    JWT_SECRET="your-super-secret-key"
    ```

3.  **Start backend services:**
    Run the PostgreSQL and Redis containers using Docker Compose.
    ```bash
    docker-compose up -d
    ```

4.  **Enable DB Extension (First Time Only):**
    Connect to the database and enable the UUID extension.
    ```bash
    docker exec -it exchange_postgres psql -U your_user -d exchange
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    \q
    ```

5.  **Run the Go services:**
    Open four separate terminals in the project root and run each service.
    ```bash
    # Terminal 1
    go run ./cmd/api

    # Terminal 2
    go run ./cmd/engine

    # Terminal 3
    go run ./cmd/db-processor

    # Terminal 4
    go run ./cmd/ws
    ```
The system is now fully running.

---

## 🛠️ API Usage Example

1.  **Sign up:**
    ```bash
    curl -X POST http://localhost:8080/api/v1/auth/signup \
    -H "Content-Type: application/json" \
    -d '{"username": "test", "email": "test@test.com", "password": "password123"}'
    ```

2.  **Log in to get a token:**
    ```bash
    curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email": "test@test.com", "password": "password123"}'
    ```

3.  **Place an order:**
    ```bash
    curl -X POST http://localhost:8080/api/v1/orders \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <YOUR_TOKEN>" \
    -d '{"market": "SOL_USDC", "price": "150", "quantity": "10", "side": "buy"}'
    ```
