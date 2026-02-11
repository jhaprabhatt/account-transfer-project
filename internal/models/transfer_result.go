package models

import (
	"time"
)

type TransferResult struct {
	AuditID           int64
	CorrelationID     int64
	Status            string
	SourcePostBalance string
	CreatedAt         time.Time
}
