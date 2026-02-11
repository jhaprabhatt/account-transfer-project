package middleware

import (
	"context"
	"strconv"

	"github.com/jhaprabhatt/account-transfer-project/internal/pkg/idgen"

	"google.golang.org/grpc/metadata"

	"net/http"
)

type key int

const CorrelationKey key = 0

func GRPCCorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := idgen.NextId()
		idStr := strconv.FormatInt(id, 10)

		md := metadata.Pairs("correlation_id", idStr)

		ctx := metadata.NewOutgoingContext(r.Context(), md)

		ctx = context.WithValue(ctx, CorrelationKey, idStr)

		w.Header().Set("X-Correlation-ID", idStr)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
