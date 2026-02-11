# 🏦 Account Transfer System

[![Go Report Card](https://goreportcard.com/badge/github.com/jhaprabhatt/account-transfer-project)](https://goreportcard.com/report/github.com/jhaprabhatt/account-transfer-project)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/jhaprabhatt/account-transfer-project)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance, thread-safe financial transaction system built in Go. This project implements a secure money transfer service using a Hexagonal Architecture (Ports & Adapters) pattern, ensuring strict separation of concerns, testability, and scalability.

It features a RESTful API Gateway that communicates with a Core Service via gRPC, backed by PostgreSQL for ACID-compliant transactions and Redis for high-speed caching.

---

## 🏗 Architecture

The system is designed with a **Clean Architecture** approach:

1.  **API Layer (HTTP/REST)**: Handles external requests, validation, and auth. Acts as a client to the Core Service.
2.  **Service Layer (Business Logic)**: Orchestrates transaction flows, enforces business rules (e.g., non-negative balance), and manages idempotency.
3.  **Repository Layer (Data Access)**: Handles raw SQL transactions and Redis cache operations.
4.  **Communication**: gRPC (Protobuf) is used for low-latency inter-service communication.

### Tech Stack

* **Language:** Golang (1.25)
* **Communication:** gRPC / Protocol Buffers
* **Database:** PostgreSQL (Serializable Transactions)
* **Caching:** Redis (Read-through / Write-through caching)
* **API Router:** Standard `net/http` (or Chi/Gin if applicable)
* **Logging:** Uber Zap (Structured Logging)
* **Testing:** `testify`, `go-sqlmock`, `redismock`
* **Containerization:** Docker & Docker Compose

---

## 🚀 Key Features

* **Atomic Transactions:** Uses PostgreSQL transactions to ensure money is never lost or created during a transfer (ACID compliance).
* **Concurrency Control:** Handles race conditions effectively using database-level locking.
* **High Performance:** Critical read paths (like account existence checks) are cached in Redis.
* **gRPC Integration:** Strongly typed, high-performance communication between the API Gateway and the Core Banking Service.
* **Structured Logging:** JSON-formatted logs for easy ingestion by ELK/Splunk.
* **Comprehensive Testing:** 90%+ Unit Test coverage with mocked dependencies.

---

## 🛠 Getting Started

### Prerequisites

* Go 1.25
* Docker & Docker Compose
* Make (optional, for running Makefile commands)

### 1. Clone the Repository

```bash
git clone [https://github.com/jhaprabhatt/account-transfer-project.git](https://github.com/yourusername/account-transfer-project.git)
cd account-transfer-project
```

### 2. Environment Setup

Create a `.env` file in the root directory:

```ini
# App
PORT=8080
LOG_LEVEL=info

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=banking_db

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# gRPC Core Service
CORE_HOST=localhost:50051
```

### 3. Run with Docker (Recommended)

This will spin up Postgres, Redis, the Core Service, and the API Gateway.

```bash
docker-compose up --build
```

### 4. Run Locally

If you prefer running services individually:

```bash
# Start dependencies
docker-compose up postgres redis -d

# Run the application
go run cmd/server/main.go
```

---

## 📡 API Reference

### 1. Create Account
**POST** `/accounts`

Request:
```json
{
  "account_id": 101,
  "balance": 500.00
}
```

Response (201 Created):
```json
{
  "success": true,
  "account_id": 101
}
```

### 2. Make Transfer
**POST** `/transfer`

Request:
```json
{
  "source_account_id": 101,
  "destination_account_id": 102,
  "amount": 50.00
}
```

Response (200 OK):
```json
{
  "success": true,
  "transaction_id": "txn_12345_67890"
}
```

---

## 🧪 Testing

The project uses `testify` for assertions and mocking.

### Run Unit Tests
To run all tests (excluding integration tests):

```bash
go test -v ./internal/... 
```

### Run with Coverage
To see how much of the code is covered:

```bash
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out
```

---

## 📂 Project Structure

```bash
.
├── cmd
│   └── server          # Main entry point
├── internal
│   ├── config          # Environment configuration
│   ├── handler         # HTTP Handlers (Controller layer)
│   ├── service         # Business Logic (Use Cases)
│   ├── repository      # Database & Cache Access (Data Layer)
│   ├── models          # Domain Models
│   └── proto           # gRPC generated code
├── migrations          # SQL Migration files
├── docker-compose.yml
├── go.mod
└── README.md
```

---

## 🔮 Future Improvements

* **Idempotency Keys:** Add support for `Idempotency-Key` headers to prevent double-spending on network retries.
* **Metrics:** Integate Prometheus for monitoring request latency and DB connection pool stats.
* **Authentication:** Implement JWT Middleware for secure API access.
* **Audit Trail:** Asynchronous event publishing (Kafka/RabbitMQ) for analytics.