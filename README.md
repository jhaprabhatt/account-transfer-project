# 🏦 Account Transfer System

[![Go Report Card](https://goreportcard.com/badge/github.com/jhaprabhatt/account-transfer-project?refresh=1)](https://goreportcard.com/report/github.com/jhaprabhatt/account-transfer-project)
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

## 🏗 Architecture & Request Flow

The system follows a two-service architecture:

1. REST API Service
2. Core Banking Service (gRPC)

The REST service acts as a transport layer and delegates all business logic to the Core service.

---

### 🔁 High-Level Flow

Client → REST API → gRPC → Core Service → PostgreSQL + Redis → Response

---

## 1️⃣ REST Service (Transport Layer)

Responsibilities:

- Accept HTTP requests
- Parse JSON payload
- Perform basic input validation
- Return early on validation failure
- Forward validated requests to Core service via gRPC

Flow:

1. Receive HTTP request
2. Validate request payload (required fields, format, basic checks)
3. If validation fails:
   - Return 4xx response immediately
4. If validation succeeds:
   - Construct gRPC request
   - Send request to Core service
   - Await response
5. Translate gRPC response → HTTP response

The REST layer contains **no business logic**.

---

## 2️⃣ Core Service (Business Logic Layer)

The Core service is responsible for:

- Business validation
- Balance checks
- Transaction processing
- Database operations
- Redis cache interaction
- Concurrency control
- Deadlock avoidance

---

### 🔐 Core Service Processing Flow

1. Receive gRPC request
2. Perform validation using Redis cache
   - Validate account existence
   - Validate account state
3. If validation fails:
   - Return error response
4. If validation succeeds:
   - Start PostgreSQL transaction
   - Lock accounts in deterministic order
   - Perform balance checks
   - Update balances
   - Commit transaction
5. Update Redis cache accordingly
6. Return success response

---

## 🧠 Redis Usage Strategy

Redis is used as:

- Fast validation layer
- Account existence lookup
- Read-optimization layer

### Cache Initialization

- When Core service boots up:
   - It loads relevant account metadata into Redis
   - This enables fast validation during transfers

### Cache Update Strategy

- Redis is updated within the transaction flow
- Database remains the source of truth
- Cache is kept consistent after DB commit

Redis is treated as an optimization layer, not a system of record.

---

## 🗄 Database (Source of Truth)

- PostgreSQL is the primary datastore
- Serializable isolation level is used
- Row-level locking prevents race conditions
- Deterministic lock ordering prevents deadlocks

---

## 🛡 Concurrency Model

Transfers lock accounts in sorted order:

```go
firstID, secondID := req.SourceID, req.DestinationID
if firstID > secondID {
firstID, secondID = secondID, firstID
}
```

This guarantees:

- No circular lock dependencies
- No deadlocks
- Deterministic transaction ordering

---

## 📌 Separation of Concerns

REST Service:
- Transport
- Input validation
- Protocol translation

Core Service:
- Business rules
- Transaction management
- Data integrity
- Cache management

This separation enables:

- Independent scaling
- Clear ownership boundaries
- Clean architecture principles
- Easier testability

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
│   ├── api                 # REST API entrypoint
│   │   └── main.go
│   └── core                # Core gRPC service entrypoint
│       └── main.go
│
├── deploy
│   └── postgres
│       └── init.sql        # DB initialization scripts
│
├── internal
│   ├── api
│   │   ├── handler         # HTTP handlers (transport layer)
│   │   └── middleware      # HTTP middleware
│   │
│   ├── config              # Configuration loading & environment parsing
│   ├── constants           # Application-wide constants
│   │
│   ├── core
│   │   ├── handler         # gRPC handlers
│   │   └── interceptors    # gRPC interceptors (tracing)
│   │
│   ├── grpcclient          # gRPC client used by API service
│   ├── logger              # Structured logging setup (Zap)
│   ├── models              # Domain models / entities
│   ├── pkg                 # Shared internal utilities
│   ├── proto               # Protobuf definitions / generated files
│   ├── repository          # PostgreSQL + Redis data access
│   └── service             # Business logic (use cases)
│
├── .env
├── docker-compose.yml
├── Dockerfile.api
├── Dockerfile.core
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

## 🧪 Post Deployment Verification (PDV)

This project includes a Post Deployment Verification (PDV) script to validate the system after deployment.

The PDV ensures:

- Account creation works
- Duplicate account detection works
- Negative balance validation works
- Transfer succeeds
- Invalid account IDs are rejected
- Same source/destination transfer is rejected
- Insufficient funds is rejected
- Negative transfer amount is rejected

---

### 🔎 What PDV Validates

| Scenario | Expected HTTP Code |
|----------|-------------------|
| Account Created | 201 |
| Duplicate Account | 409 |
| Negative Balance | 400 |
| Successful Transfer | 200 |
| Invalid Account ID | 400 |
| Same Account Transfer | 400 |
| Insufficient Funds | 422 |
| Negative Transfer Amount | 400 |

---

### ▶️ How to Run PDV

Make sure all services are running:

```bash
docker compose up -d
```

Then execute:

```bash
chmod +x scripts/post_deployment_verification.sh
BASE_URL=http://localhost:8080 ./scripts/post_deployment_verification.sh
```

---

### 📌 Example Output

```
$ ./post_deployment_verification.sh
Running Post Deployment Verification against: http://localhost:8080

✅ PASS: POST /accounts (HTTP 201, correlation_id=2021546445781340160)
✅ PASS: POST /accounts (HTTP 201, correlation_id=2021546446972522496)
✅ PASS: POST /accounts (HTTP 409, correlation_id=2021546448159510528)
✅ PASS: POST /accounts (HTTP 400, correlation_id=2021546449510076416)
✅ PASS: POST /accounts (HTTP 201, correlation_id=2021546450688675840)
✅ PASS: POST /transfers (HTTP 200, correlation_id=2021546451988910080)
✅ PASS: POST /transfers (HTTP 400, correlation_id=2021546453335281664)
✅ PASS: POST /transfers (HTTP 400, correlation_id=2021546454564212736)
✅ PASS: POST /transfers (HTTP 400, correlation_id=2021546456011247616)
✅ PASS: POST /transfers (HTTP 400, correlation_id=2021546457382785024)
✅ PASS: POST /transfers (HTTP 400, correlation_id=2021546458607521792)
✅ PASS: POST /transfers (HTTP 422, correlation_id=2021546460167802880)
✅ PASS: POST /transfers (HTTP 400, correlation_id=2021546461455454208)

🎉 Post Deployment Verification completed successfully.


```

---

### 🧠 Why PDV Is Important

The PDV script validates both:

- Transport layer correctness (HTTP responses)
- Business rule enforcement
- Status code semantics
- System wiring (API ↔ Core ↔ DB ↔ Redis)

It ensures the deployment is healthy beyond simple health checks.

---

### ⚠️ Notes

- The PDV script generates unique account IDs to avoid collisions.
- If you want a clean database state before running PDV:

```bash
docker compose down -v
docker compose up -d
```

This removes the PostgreSQL volume and reinitializes the database.

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

POST /transfers

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
