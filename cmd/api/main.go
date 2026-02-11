package main

import (
	"account-transfer-project/internal/api/handler"
	atm "account-transfer-project/internal/api/middleware"
	"account-transfer-project/internal/grpcclient"
	"account-transfer-project/internal/logger"
	"account-transfer-project/internal/pkg/idgen"
	pb "account-transfer-project/internal/proto"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	log := logger.InitLogger("account-transfer-api", "info")

	defer func() {
		_ = log.Sync()
	}()

	if err := idgen.Init(1, log); err != nil {
		log.Fatal("Failed to initialize snowflake: %v", zap.Error(err))
	}

	conn := grpcclient.NewConnection(log)

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Error("Failed to get gRPC client connection", zap.Error(err))
		}
	}(conn)

	accountClient := pb.NewAccountServiceClient(conn)
	transferClient := pb.NewTransferServiceClient(conn)

	accountHandler := handler.NewAccountHandler(accountClient, log)
	transferHandler := handler.NewTransactionHandler(transferClient, log)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(atm.GRPCCorrelationMiddleware)

	r.Post("/accounts", accountHandler.CreateAccount)
	r.Post("/transfers", transferHandler.MakeTransfer)
	log.Info("Server Listening", zap.Int("port", 8080))

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Server crashed", zap.Error(err))
	}
}
