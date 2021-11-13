package writer

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"nir/amlerror"
	"sync"
	"time"
)

const (
	GasLimit = 21000
	ChainID  = 45439

	errCtx = amlerror.AMLError("writer is missing in context")
)

type Writer struct {
	connects []*ethclient.Client
}

func (w Writer) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return w.executeAll(func(client *ethclient.Client) error {
		return client.SendTransaction(ctx, tx)
	})
}

func (w Writer) DeployContract(auth *bind.TransactOpts, deployFunc func(auth *bind.TransactOpts, backend bind.ContractBackend) error) error {
	return w.executeAll(func(client *ethclient.Client) error {
		return deployFunc(auth, client)
	})
}

func (w Writer) BalanceAt(ctx context.Context, addr common.Address) (balance *big.Int, err error) {
	err = w.executeAll(func(client *ethclient.Client) (innerErr error) {
		balance, innerErr = client.BalanceAt(ctx, addr, nil)
		return innerErr
	})

	return balance, err
}

func (w Writer) BlockNumber(ctx context.Context) (number uint64, err error) {
	err = w.executeAll(func(client *ethclient.Client) (innerErr error) {
		number, innerErr = client.BlockNumber(ctx)
		return innerErr
	})

	return number, err
}

func (w Writer) SuggestGasPrice(ctx context.Context) (gasPrice *big.Int, err error) {
	err = w.executeAll(func(client *ethclient.Client) (innerErr error) {
		gasPrice, innerErr = client.SuggestGasPrice(ctx)

		return innerErr
	})

	return gasPrice, err
}

func (w Writer) WaitTx(ctx context.Context, hash common.Hash) error {
	return w.executeAll(func(client *ethclient.Client) error {
		tick := time.NewTicker(time.Second)
		defer tick.Stop()

		for range tick.C {
			select {
			case <-ctx.Done():
				return fmt.Errorf("timeout")
			default:
				_, err := client.TransactionReceipt(ctx, hash)
				if err != nil {
					if errors.Is(err, ethereum.NotFound) {
						continue
					}

					return err
				}

				return nil
			}
		}

		return nil
	})
}

func (w Writer) executeAll(fn func(client *ethclient.Client) error) error {
	var (
		wg    sync.WaitGroup
		errCh = make(chan error, len(w.connects))
	)

	for _, eth := range w.connects {
		wg.Add(1)

		eth := eth
		go func() {
			defer wg.Done()

			err := fn(eth)
			if err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()

	if len(errCh) == len(w.connects) {
		return <-errCh
	}

	return nil
}

func Connect(ctx context.Context, addresses []string) (*Writer, error) {
	w := &Writer{}

	for _, addr := range addresses {
		ethDirect, err := rpc.DialContext(ctx, addr)
		if err != nil {
			return nil, fmt.Errorf("dial error %w", err)
		}

		w.connects = append(w.connects, ethclient.NewClient(ethDirect))
	}

	return w, nil
}

type writerKey struct{}

func WithWriter(ctx context.Context, writer *Writer) context.Context {
	return context.WithValue(ctx, writerKey{}, writer)
}

func FromContext(ctx context.Context) (*Writer, error) {
	writer, ok := ctx.Value(writerKey{}).(*Writer)
	if !ok {
		return nil, errCtx
	}

	return writer, nil
}

func Execute(ctx context.Context, fn func(w *Writer) error) error {
	writer, err := FromContext(ctx)
	if err != nil {
		return err
	}

	return fn(writer)
}
