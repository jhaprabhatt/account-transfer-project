# 🏦 Account Transfer System

[![Go Report Card](https://goreportcard.com/badge/github.com/jhaprabhatt/account-transfer-project)](https://goreportcard.com/report/github.com/jhaprabhatt/account-transfer-project)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/jhaprabhatt/account-transfer-project)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A production-grade, thread-safe financial transaction system built in Go.

This project demonstrates how to design and implement a **secure, ACID-compliant money transfer system** using Clean Architecture (Hexagonal Architecture / Ports & Adapters).

The system consists of:

- REST API Gateway (HTTP)
- Core Banking Service (gRPC)
- PostgreSQL (Serializable Transactions)
- Redis (Caching Layer)

It focuses on correctness, concurrency safety, and deterministic transaction ordering.

---

## 🎯 Design Goals

- Prevent double spending
- Guarantee atomic money transfers
- Avoid deadlocks under high concurrency
- Maintain strict separation of concerns
- Enable high testability via dependency inversion
- Be production-ready and extensible

---

## 🏗 Architecture Overview

The system follows **Clean Architecture principles**.

### Layers

1. **API Layer (HTTP/REST)**
    - Input validation
    - Request parsing
    - Transport mapping
    - Acts as gRPC client to Core Service

2. **Application / Service Layer**
    - Business rules
    - Transaction orchestration
    - Deadlock prevention via deterministic locking
    - Idempotency logic (extensible)

3. **Repository Layer**
    - PostgreSQL transaction management
    - Row-level locking
    - Redis caching logic

4. **Infrastructure**
    - Logging (Zap)
    - Configuration
    - gRPC transport
    - Docker orchestration

---

## 🔐 Concurrency & Deadlock Prevention

Transfers always lock accounts in deterministic order:

```go
firstID, secondID := req.SourceID, req.DestinationID
if firstID > secondID {
firstID, secondID = secondID, firstID
}
```

This prevents circular lock acquisition and eliminates database-level deadlocks under concurrent transfers.

---

## 🛠 Tech Stack

- Language: Go (1.25)
- Communication: gRPC / Protocol Buffers
- Database: PostgreSQL (Serializable Isolation)
- Caching: Redis
- Logging: Uber Zap (Structured JSON logs)
- Testing: testify, go-sqlmock, redismock
- Containerization: Docker & Docker Compose

---

## 🚀 Key Features

- ✅ ACID-compliant money transfers
- ✅ Serializable database transactions
- ✅ Deterministic lock ordering
- ✅ Redis caching for read-heavy paths
- ✅ Structured logging
- ✅ 90%+ unit test coverage
- ✅ Mocked database & cache for isolation testing
- ✅ Clean architecture for scalability

---

## 📦 Project Structure

```
.
├── cmd
│   └── server
├── internal
│   ├── config
│   ├── handler
│   ├── service
│   ├── repository
│   ├── models
│   └── proto
├── migrations
├── docker-compose.yml
├── go.mod
└── README.md
```

---

## 🛠 Getting Started

### Prerequisites

- Go 1.25
- Docker & Docker Compose

---

### Clone Repository

```bash
git clone https://github.com/jhaprabhatt/account-transfer-project.git
cd account-transfer-project
```

---

### Environment Variables

Create `.env` file:

```ini
PORT=8080
LOG_LEVEL=info

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=banking_db

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

CORE_HOST=localhost:50051
```

---

### Run with Docker

```bash
docker-compose up --build
```

---

## 📡 API Endpoints

### Create Account

POST /accounts

Request:
```json
{
"account_id": 101,
"balance": 500.00
}
```

Response:
```json
{
"success": true,
"account_id": 101
}
```

---

### Transfer Money

POST /transfer

Request:
```json
{
"source_account_id": 101,
"destination_account_id": 102,
"amount": 50.00
}
```

Response:
```json
{
"success": true,
"transaction_id": "12345"
}
```

---

## 🧪 Testing

Run all unit tests:

```bash
go test -v ./internal/...
```

Run with coverage:

```bash
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out
```

---

## ⚠️ Assumptions & Simplifications

This project intentionally focuses on core transaction correctness and concurrency safety.  
The following assumptions are made for simplicity:

### 1. Authentication & Authorization

- No authentication (AuthN) is implemented.
- No authorization (AuthZ) or role-based access control exists.
- All requests are assumed to be trusted.

### 2. Account Balance Rules

- An account balance must never go below zero.
- Transfers that would result in a negative balance are rejected.
- Overdraft protection is not supported.

### 3. Encryption

- User data is not encrypted at rest.
- Sensitive fields are not encrypted in database storage.
- Responses are not encrypted at the application layer.
- TLS termination is assumed to be handled externally (e.g., via reverse proxy).

### 4. Database

- PostgreSQL is used as the primary datastore.
- Serializable isolation level is assumed for strict consistency.
- No database sharding or replication is implemented.

### 5. Idempotency & Duplicate Detection

- Duplicate transaction detection is not implemented.
- The system assumes clients do not retry requests blindly.
- No Idempotency-Key header is supported.

### 6. Distributed System Constraints

- The system runs in a single-region environment.
- No cross-region replication is implemented.
- No message queue or eventual consistency patterns are used.

---

## 📌 Scope Clarification

This project is intentionally designed to demonstrate:

- ACID-compliant transaction handling
- Concurrency-safe account transfers
- Deadlock avoidance via deterministic locking
- Clean Architecture implementation in Go

It does **not** aim to be a fully production-ready banking system.

Security hardening, distributed scalability, and advanced compliance requirements
are intentionally excluded to keep focus on core transaction logic.


## 📈 Future Enhancements

- Idempotency-Key header support
- Prometheus metrics
- OpenTelemetry tracing
- JWT authentication
- Audit event publishing (Kafka)
- Rate limiting
- Circuit breaker for downstream services

---
