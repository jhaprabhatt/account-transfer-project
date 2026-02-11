package constants

type TransferStatus int

const (
	StatusPending TransferStatus = iota + 1
	StatusCompleted
	StatusFailed
)
