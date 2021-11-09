package exchange

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gocarina/gocsv"
	"github.com/google/uuid"
	"log"
	"math/big"
	"nir/clustering/blockchain"
	"nir/config"
	"nir/test/entity/account"
	"nir/test/writer"
	"os"
	"strings"
	"time"
)

type Exchange struct {
	name       string
	account    *account.Account
	clients    map[common.Address]*exchangeClient
	incomingTx chan *types.LegacyTx
	closed     chan struct{}
}

type exchangeClient struct {
	account *common.Address
	deposit *account.Account
}

func CreateExchange(ctx context.Context) (*Exchange, error) {
	acc, err := account.CreateAccount()
	if err != nil {
		return nil, err
	}

	fmt.Println("EXCHANGE CREATED", acc.GetAddress().String())
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	exch := &Exchange{
		name:       id.String(),
		account:    acc,
		clients:    make(map[common.Address]*exchangeClient),
		incomingTx: make(chan *types.LegacyTx, 1000),
		closed:     make(chan struct{}),
	}

	go exch.startGettingTxs(ctx)

	return exch, nil
}

func (exch *Exchange) startGettingTxs(ctx context.Context) {
	for tx := range exch.incomingTx {
		err := exch.GetEthFromDeposit(ctx, tx)
		if err != nil {
			log.Println("can't get eth from deposit", err)
		}
	}

	close(exch.closed)

	fmt.Println("EXCHANGE CLOSED")
}

func (exch *Exchange) Close() {
	close(exch.incomingTx)

	<-exch.closed
}

func (exch *Exchange) AddIncomingTransaction(tx *types.LegacyTx) {
	exch.incomingTx <- tx
}

func (exch *Exchange) CreateAccountIfNotExist(address *common.Address) (*common.Address, error) {
	acc, err := account.CreateAccount()
	if err != nil {
		return nil, err
	}

	exchAcc, ok := exch.clients[*address]
	if ok {
		return exchAcc.deposit.GetAddress(), nil
	}

	exch.clients[*address] = &exchangeClient{
		account: address,
		deposit: acc,
	}

	return acc.GetAddress(), nil
}

func (exch *Exchange) GetName() string {
	return exch.name
}

func (exch *Exchange) GetEthFromDeposit(ctx context.Context, tx *types.LegacyTx) error {
	for _, client := range exch.clients {
		address := client.deposit.GetAddress()
		if address != tx.To {
			continue
		}

		_, err := exch.account.SendTransaction(ctx, address, writer.GasLimit, false)
		if err != nil {
			return err
		}

		err = waitBalance(ctx, tx.Gas*tx.GasPrice.Uint64()+tx.Value.Uint64(), *tx.To)
		if err != nil {
			return err
		}

		_, err = client.deposit.SendTransaction(ctx, exch.account.GetAddress(), tx.Value.Int64(), false)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (exch *Exchange) GetAccounts() (addresses []*common.Address) {
	return []*common.Address{exch.account.GetAddress()}
}

func waitBalance(ctx context.Context, neededBalance uint64, address common.Address) error {
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

func SaveExchanges(exchanges []*Exchange) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	var exchs blockchain.Exchanges
	for _, exch := range exchanges {
		exchs = append(exchs, &blockchain.Exchange{
			Address:     strings.ToLower(exch.account.GetAddress().String()),
			Name:        exch.GetName(),
			AccountType: "eoa",
			Type:        "Exchange",
		})
	}

	f, err := os.Create(cfg.ExchangesTable)
	if err != nil {
		return err
	}

	return gocsv.Marshal(exchs, f)
}
