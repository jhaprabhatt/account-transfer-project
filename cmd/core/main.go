package main

import (
	"context"
	"database/sql"
	"github.com/jhaprabhatt/account-transfer-project/internal/config"
	"github.com/jhaprabhatt/account-transfer-project/internal/core/handler"
	"github.com/jhaprabhatt/account-transfer-project/internal/core/interceptors"
	"github.com/jhaprabhatt/account-transfer-project/internal/logger"
	pb "github.com/jhaprabhatt/account-transfer-project/internal/proto"
	"github.com/jhaprabhatt/account-transfer-project/internal/repository"
	"github.com/jhaprabhatt/account-transfer-project/internal/service"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	log := logger.InitLogger("account-transfer-core", "info")

	defer func() {
		_ = log.Sync()
	}()

	dbConfig := config.LoadDatabaseConfig()

	log.Info("Connecting to database",
		zap.String("host", dbConfig.Host),
		zap.String("port", dbConfig.Port),
	)

	db, err := sql.Open("pgx", dbConfig.ConnectionString())
	if err != nil {
		log.Fatal("Failed to open DB connection", zap.Error(err))
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Error("Error closing DB connection")
		}
	}(db)

	if err := db.Ping(); err != nil {

		log.Fatal("Failed to ping DB", zap.Error(err))
	}

	accRepo := repository.NewAccountRepository(db, log)
	transferRepo := repository.NewTransferRepository(db, log)
	cache := repository.NewAccountCache()
	accSvc := service.NewAccountService(accRepo, cache, log)
	txSvc := service.NewTransferService(transferRepo, cache, log)

	log.Info("Starting Cache Warm-up...")
	ctx := context.Background()
	if err := accSvc.LoadAllAccountsToCache(ctx); err != nil {
		log.Fatal("Failed to warm up cache", zap.Error(err))
	}
	log.Info("Cache Warm-up Complete")

	grpcHandler := handler.NewGrpcHandler(accSvc, txSvc, log)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("Failed to listen on port 50051", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.UnaryCorrelationInterceptor()),
	)

	pb.RegisterAccountServiceServer(grpcServer, grpcHandler)
	pb.RegisterTransferServiceServer(grpcServer, grpcHandler)

	log.Info("Core Service listening via gRPC", zap.String("address", ":50051"))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve gRPC", zap.Error(err))
	}
}
