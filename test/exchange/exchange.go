package exchange

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gocarina/gocsv"
	"github.com/google/uuid"
	"log"
	"nir/database"
	"nir/test/entity/account"
	"nir/test/writer"
	"os"
	"strings"
	"sync"
)

type Exchange struct {
	name       string
	account    *account.Account
	clients    map[common.Address]*exchangeClient
	incomingTx chan *types.LegacyTx
	closed     chan struct{}

	wg sync.WaitGroup
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
			log.Println("can't get eth from transfer", err)
		}
	}

	exch.wg.Wait()

	close(exch.closed)
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

		exch.wg.Add(1)

		go func() {
			defer exch.wg.Done()

			err = account.WaitBalance(ctx, tx.Gas*tx.GasPrice.Uint64()+tx.Value.Uint64(), *tx.To)
			if err != nil {
				fmt.Println(err)

				return
			}

			_, err = client.deposit.SendTransaction(ctx, exch.account.GetAddress(), tx.Value.Int64(), true)
			if err != nil {
				fmt.Println(err)

				return
			}
		}()

		return nil
	}

	return nil
}

func (exch *Exchange) GetAccounts() (addresses []*common.Address) {
	return []*common.Address{exch.account.GetAddress()}
}

func SaveExchanges(exchanges []*Exchange) error {
	var exchs database.Exchanges
	for _, exch := range exchanges {
		exchs = append(exchs, &database.Exchange{
			Address: strings.ToLower(exch.account.GetAddress().String()),
			Name:    exch.GetName(),
		})
	}

	f, err := os.Create("geth/data/exchanges.csv")
	if err != nil {
		return err
	}

	return gocsv.Marshal(exchs, f)
}
