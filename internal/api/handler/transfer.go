package handler

import (
	"encoding/json"
	"github.com/jhaprabhatt/account-transfer-project/internal/api/middleware"
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jhaprabhatt/account-transfer-project/internal/models"
	pb "github.com/jhaprabhatt/account-transfer-project/internal/proto"
)

type TransactionHandler struct {
	client pb.TransferServiceClient
	log    *zap.Logger
}

func NewTransactionHandler(client pb.TransferServiceClient, log *zap.Logger) *TransactionHandler {
	return &TransactionHandler{
		client: client,
		log:    log,
	}
}

func (h *TransactionHandler) MakeTransfer(w http.ResponseWriter, r *http.Request) {

	correlationID, _ := r.Context().Value(middleware.CorrelationKey).(string)

	var req models.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("Failed to decode transfer request", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		h.log.Warn("Invalid transfer request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	grpcReq := &pb.TransferRequest{
		SourceId:      req.SourceID,
		DestinationId: req.DestinationID,
		Amount:        req.Amount.String(),
	}

	h.log.Info("Initiating transfer",
		zap.String("correlation_id", correlationID),
		zap.Int64("source", req.SourceID),
		zap.Int64("destination", req.DestinationID),
	)

	resp, err := h.client.MakeTransfer(r.Context(), grpcReq)
	if err != nil {
		st, _ := status.FromError(err)

		h.log.Error("Transfer failed via gRPC",
			zap.String("correlation_id", correlationID),
			zap.String("grpc_code", st.Code().String()),
			zap.Error(err),
		)

		switch st.Code() {
		case codes.NotFound:
			http.Error(w, st.Message(), http.StatusNotFound)
		case codes.FailedPrecondition:
			http.Error(w, st.Message(), http.StatusUnprocessableEntity)
		case codes.InvalidArgument:
			http.Error(w, st.Message(), http.StatusBadRequest)
		case codes.AlreadyExists:
			http.Error(w, st.Message(), http.StatusConflict)
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error("Failed to write response", zap.Error(err))
	}
}
