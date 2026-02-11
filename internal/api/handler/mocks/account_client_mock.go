package mocks

import (
	"context"
	pb "github.com/jhaprabhatt/account-transfer-project/internal/proto"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockAccountServiceClient struct {
	mock.Mock
}

func (m *MockAccountServiceClient) CreateAccount(ctx context.Context, in *pb.CreateAccountRequest, opts ...grpc.CallOption) (*pb.CreateAccountResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreateAccountResponse), args.Error(1)
}
