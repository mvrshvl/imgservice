package account

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"nir/amlerror"
	"nir/test/writer"
	"time"
)

const (
	commonBalanceKey = "29f4455cd82770e305096f33c2a53f13efed2974873c883a6d7ca7d1bdcdf0c7"

	errECDSA = amlerror.AMLError("error casting public key to ECDSA")
)

type Account struct {
	address *common.Address
	nonce   uint64
	key     *ecdsa.PrivateKey
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

func (acc *Account) SendTransaction(ctx context.Context, to *common.Address, amount int64) (*common.Address, error) {
	wr, err := writer.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	gasPrice, err := wr.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    acc.nonce,
		GasPrice: gasPrice,
		Gas:      writer.GasLimit,
		To:       to,
		Value:    big.NewInt(amount),
	})

	signedTx, err := types.SignTx(tx, types.NewEIP2930Signer(big.NewInt(writer.ChainID)), acc.key)
	if err != nil {
		return nil, err
	}

	err = wr.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	acc.nonce++

	return to, WaitTx(ctx, wr, signedTx.Hash())
}

func (acc *Account) GetAddress() *common.Address {
	return acc.address
}

func WaitTx(ctx context.Context, client *ethclient.Client, hash common.Hash) error {
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
}

func AddEtherToAccount(ctx context.Context, wr *ethclient.Client, to *common.Address, amount int64) error {
	publicKey, privateKey, err := getAccountWithCommonBalance()

	nonce, err := wr.PendingNonceAt(ctx, crypto.PubkeyToAddress(*publicKey))
	if err != nil {
		return err
	}

	gasPrice, err := wr.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      writer.GasLimit,
		To:       to,
		Value:    big.NewInt(amount),
	})

	signedTx, err := types.SignTx(tx, types.NewEIP2930Signer(big.NewInt(writer.ChainID)), privateKey)
	if err != nil {
		return err
	}

	err = wr.SendTransaction(ctx, signedTx)
	if err != nil {
		return err
	}

	err = WaitTx(ctx, wr, signedTx.Hash())
	if err != nil {
		return err
	}

	return nil
}

func getAccountWithCommonBalance() (*ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	privateKey, err := crypto.HexToECDSA(commonBalanceKey)
	if err != nil {
		return nil, nil, err
	}

	cryptoPublicKey := privateKey.Public()
	publicKey, ok := cryptoPublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, errECDSA
	}

	return publicKey, privateKey, nil
}
