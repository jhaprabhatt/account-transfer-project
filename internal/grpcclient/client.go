package grpcclient

import (
	"account-transfer-project/internal/config"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewConnection(log *zap.Logger) *grpc.ClientConn {
	coreHost := config.GetEnv("CORE_HOST", "localhost:50051")

	log.Info("Attempting to dial Core Service", zap.String("host", coreHost))

	conn, err := grpc.NewClient(coreHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Could not connect to Core BE")
	}

	return conn
}
