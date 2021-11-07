package writer

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"nir/amlerror"
)

const (
	GasLimit = 21000
	ChainID  = 45439

	errCtx = amlerror.AMLError("writer is missing in context")
)

func Connect(ctx context.Context, address string) (*ethclient.Client, error) {
	ethDirect, err := rpc.DialContext(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("dial error %w", err)
	}

	return ethclient.NewClient(ethDirect), nil
}

type writerKey struct{}

func WithWriter(ctx context.Context, writer *ethclient.Client) context.Context {
	return context.WithValue(ctx, writerKey{}, writer)
}

func FromContext(ctx context.Context) (*ethclient.Client, error) {
	writer, ok := ctx.Value(writerKey{}).(*ethclient.Client)
	if !ok {
		return nil, errCtx
	}

	return writer, nil
}
