package geth

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"nir/config"
)

type Worker struct {
	connect *ethclient.Client
}

func New(cfg *config.Config) (*Worker, error) {
	ethDirect, err := rpc.Dial(cfg.Ethereum.Address)
	if err != nil {
		return nil, fmt.Errorf("dial error %w", err)
	}

	return &Worker{connect: ethclient.NewClient(ethDirect)}, nil
}

func (w *Worker) GetLastBlock(ctx context.Context) (uint64, error) {
	return w.connect.BlockNumber(ctx)
}
