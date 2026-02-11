package mocks

import (
	pb "account-transfer-project/internal/proto"
	"context"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockTransferServiceClient struct {
	mock.Mock
}

func (m *MockTransferServiceClient) MakeTransfer(ctx context.Context, in *pb.TransferRequest, opts ...grpc.CallOption) (*pb.TransferResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.TransferResponse), args.Error(1)
}
