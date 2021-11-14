package account

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"nir/test/writer"
	"sync"
	"time"
)

type Account struct {
	address *common.Address
	nonce   uint64
	key     *ecdsa.PrivateKey

	mux sync.RWMutex
}

func CreateAccount() (*Account, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	address := crypto.PubkeyToAddress(key.PublicKey)

	return &Account{
		address: &address,
		key:     key,
	}, nil
}

func (acc *Account) SendTransaction(ctx context.Context, to *common.Address, amount int64, wait bool) (*types.LegacyTx, error) {
	var (
		gasPrice *big.Int
	)

	err := writer.Execute(ctx, func(w *writer.Writer) (innerErr error) {
		gasPrice, innerErr = w.SuggestGasPrice(ctx)

		return innerErr
	})
	if err != nil {
		return nil, err
	}

	acc.mux.Lock()
	defer acc.mux.Unlock()

	legacyTx := &types.LegacyTx{
		Nonce:    acc.nonce,
		GasPrice: gasPrice,
		Gas:      writer.GasLimit,
		To:       to,
		Value:    big.NewInt(amount),
	}

	tx := types.NewTx(legacyTx)

	signedTx, err := types.SignTx(tx, types.NewEIP2930Signer(big.NewInt(writer.ChainID)), acc.key)
	if err != nil {
		return nil, err
	}

	err = writer.Execute(ctx, func(w *writer.Writer) (innerErr error) {
		return w.SendTransaction(ctx, signedTx)
	})
	if err != nil {
		return nil, fmt.Errorf("cant send %s %w", crypto.PubkeyToAddress(acc.key.PublicKey).String(), err)
	}

	acc.nonce++

	if wait {
		err = writer.Execute(ctx, func(w *writer.Writer) (innerErr error) {
			return w.WaitTx(ctx, signedTx.Hash())
		})
		if err != nil {
			return nil, err
		}
	}

	time.Sleep(50 * time.Millisecond)

	return legacyTx, nil
}

func (acc *Account) GetAddress() *common.Address {
	return acc.address
}

func (acc *Account) ExecuteContract(ctx context.Context, gasLimit uint64, fn func(auth *bind.TransactOpts, backend bind.ContractBackend) (*types.Transaction, error)) (tx *types.Transaction, err error) {
	err = writer.Execute(ctx, func(w *writer.Writer) (innerErr error) {
		auth, innerErr := acc.getAuth(ctx, gasLimit, w)
		if innerErr != nil {
			return innerErr
		}

		tx, innerErr = w.ExecuteContract(auth, fn)
		if innerErr != nil {
			return innerErr
		}

		return nil
	})

	acc.mux.Lock()
	acc.nonce++
	acc.mux.Unlock()

	return tx, err
}

func (acc *Account) getAuth(ctx context.Context, gasLimit uint64, w *writer.Writer) (*bind.TransactOpts, error) {
	gasPrice, err := w.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(acc.key, big.NewInt(writer.ChainID))
	if err != nil {
		return nil, err
	}

	acc.mux.RLock()
	auth.Nonce = big.NewInt(int64(acc.nonce))
	acc.mux.RUnlock()

	auth.Value = big.NewInt(0)
	auth.GasLimit = gasLimit * 20
	auth.GasPrice = gasPrice

	return auth, nil
}

func NewAccountFromKey(key string) (*Account, error) {
	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return nil, err
	}

	return &Account{
		address: nil,
		nonce:   0,
		key:     privateKey,
	}, nil
}

func WaitBalance(ctx context.Context, neededBalance uint64, address common.Address) error {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	var balance *big.Int

	for range tick.C {
		err := writer.Execute(ctx, func(w *writer.Writer) (innerErr error) {
			balance, innerErr = w.BalanceAt(ctx, address)
			return innerErr
		})
		if err != nil {
			return err
		}

		if balance.Uint64() >= neededBalance {
			break
		}
	}

	return nil
}
