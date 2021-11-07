package test

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
	"nir/test/entity"
	"nir/test/entity/account"
	"nir/test/exchange"
	"nir/test/writer"
	"sync/atomic"
	"testing"
	"time"
)

const (
	errECDSA = amlerror.AMLError("error casting public key to ECDSA")
)

func TestGenerate(t *testing.T) {
	ctx := context.Background()

	wr, err := writer.Connect(ctx, writerNode)
	if err != nil {
		t.Fatal(err)
	}

	ctx = writer.WithWriter(ctx, wr)

	exchanges, err := createExchanges(countExchanges)

	clusters, err := createEntities(countCluster, maxCountAccountsInCluster)
	if err != nil {
		t.Fatal(err)
	}

	accounts, err := createEntities(countAccounts-(countCluster*maxCountAccountsInCluster), 1)
	if err != nil {
		t.Fatal(err)
	}

	entities := append(accounts, clusters...)

	for _, ent := range entities {
		err = CreateExchangesAccounts(exchanges, ent)
		if err != nil {
			t.Fatal(err)
		}

		err := addEtherToEntity(ctx, wr, ent, commonAmount)
		if err != nil {
			t.Fatal(err)
		}
	}

	SendTransactions(ctx, entities, exchanges, countTransactions)
}

func addEtherToEntity(ctx context.Context, wr *ethclient.Client, entity *entity.Entity, amount int64) error {
	for _, acc := range entity.GetAccounts() {
		err := addEtherToAccount(ctx, wr, acc, amount)
		if err != nil {
			return err
		}

		fmt.Println("ADDED", acc.GetAddress().String())
	}

	return nil
}

func addEtherToAccount(ctx context.Context, wr *ethclient.Client, to *account.Account, amount int64) error {
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
		To:       to.GetAddress(),
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

	txReceipt, err := waitTx(ctx, wr, signedTx.Hash())
	if err != nil {
		return err
	}

	if txReceipt.Status == types.ReceiptStatusFailed {
		return err
	}

	return nil
}

func waitTx(ctx context.Context, client *ethclient.Client, hash common.Hash) (*types.Receipt, error) {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for range tick.C {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout")
		default:
			txReceipt, err := client.TransactionReceipt(ctx, hash)
			if err != nil {
				if errors.Is(err, ethereum.NotFound) {
					continue
				}

				return nil, err
			}

			return txReceipt, nil
		}
	}

	return nil, nil
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

func createEntities(count, size uint64) (clusters []*entity.Entity, err error) {
	clusters = make([]*entity.Entity, count)

	for i := uint64(0); i < count; i++ {
		clusters[i], err = entity.CreateEntity(size)
		if err != nil {
			return nil, err
		}
	}

	return clusters, nil
}

func createExchanges(count uint64) (exchanges []*exchange.Exchange, err error) {
	exchanges = make([]*exchange.Exchange, count)

	for i := uint64(0); i < count; i++ {
		exchanges[i], err = exchange.CreateExchange()
		if err != nil {
			return nil, err
		}
	}

	return exchanges, nil
}

func CreateExchangesAccounts(exchanges []*exchange.Exchange, entity *entity.Entity) error {
	for _, exch := range exchanges {
		err := entity.CreateExchangeAccounts(exch)
		if err != nil {
			return err
		}
	}

	return nil
}

func SendTransactions(ctx context.Context, entities []*entity.Entity, exchanges []*exchange.Exchange, countTxs int32) {
	count := new(int32)

	for {
		for j := 0; j < transactonsPerSecond; j++ {
			go func() {
				currentTxNumber := atomic.AddInt32(count, 1)

				currentEntity := entities[currentTxNumber%int32(len(entities))]
				currentExchange := exchanges[currentTxNumber%int32(len(exchanges))]

				err := currentEntity.SendTransaction(ctx, currentExchange.GetName(), 1)
				if err != nil {
					log.Printf("can't send %d tranasction: %v\n", currentTxNumber, err)

					return
				}
			}()
		}

		if atomic.LoadInt32(count) == countTxs {
			return
		}

		countSent := atomic.LoadInt32(count)
		log.Printf("Sent %d tranasctions\n", countSent)

		time.Sleep(time.Second)
	}
}
