package interceptors

import (
	"context"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type contextKey string

const CorrelationKey contextKey = "correlation_id"

func UnaryCorrelationInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if ids := md.Get("correlation_id"); len(ids) > 0 {
				if id, err := strconv.ParseInt(ids[0], 10, 64); err == nil {
					ctx = context.WithValue(ctx, CorrelationKey, id)
				}
			}
		}
		return handler(ctx, req)
	}
}
