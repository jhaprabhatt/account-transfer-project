package handler

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"account-transfer-project/internal/models"
	pb "account-transfer-project/internal/proto"
)

type AccountHandler struct {
	client pb.AccountServiceClient
	log    *zap.Logger
}

func NewAccountHandler(client pb.AccountServiceClient, log *zap.Logger) *AccountHandler {
	return &AccountHandler{client: client, log: log}
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("Failed to decode JSON", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	grpcReq := &pb.CreateAccountRequest{AccountId: req.ID, Balance: req.Balance.String()}

	h.log.Info("Forwarding creation request to Core", zap.Int64("account_id", req.ID))

	resp, err := h.client.CreateAccount(r.Context(), grpcReq)

	if err != nil {

		st, ok := status.FromError(err)

		if ok && st.Code() == codes.AlreadyExists {
			h.log.Warn("Duplicate account creation attempt", zap.Error(err))
			http.Error(w, "Account already exists", http.StatusConflict)
			return
		}

		h.log.Error("gRPC call failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.log.Info("Account created successfully",
		zap.Int64("account_id", req.ID),
		zap.Bool("success", resp.Success),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error("Failed to write response", zap.Error(err))
	}
}
