package account

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"
	"nir/test/writer"
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

func (acc *Account) SendTransaction(ctx context.Context, to *common.Address, amount int64) error {
	wr, err := writer.FromContext(ctx)
	if err != nil {
		return err
	}

	gasPrice, err := wr.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}

	//gas, err := wr.EstimateGas(ctx, ethereum.CallMsg{
	//	From:       *acc.address,
	//	To:         to,
	//	GasPrice:   gasPrice,
	//	Value:      big.NewInt(amount),
	//})
	//if err != nil {
	//	return err
	//}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    acc.nonce,
		GasPrice: gasPrice,
		Gas:      writer.GasLimit,
		To:       to,
		Value:    big.NewInt(amount),
	})

	signedTx, err := types.SignTx(tx, types.NewEIP2930Signer(big.NewInt(writer.ChainID)), acc.key)

	err = wr.SendTransaction(ctx, signedTx)
	if err != nil {
		return err
	}

	acc.nonce++

	balance, err := wr.BalanceAt(ctx, *acc.address, nil)
	if err != nil {
		return err
	}

	fmt.Println("BALANCE", balance)

	return nil
}

func (acc *Account) GetAddress() *common.Address {
	return acc.address
}
