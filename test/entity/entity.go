package entity

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"math/rand"
	"nir/test/entity/account"
	"nir/test/exchange"
)

type Entity struct {
	accounts []*account.Account
	deposits map[string][]*common.Address
}

func CreateEntity(size uint64) (*Entity, error) {
	cluster := &Entity{
		accounts: make([]*account.Account, size),
		deposits: make(map[string][]*common.Address),
	}

	for i := uint64(0); i < size; i++ {
		acc, err := account.CreateAccount()
		if err != nil {
			return nil, err
		}

		cluster.accounts[i] = acc
	}

	return cluster, nil
}

func (cluster *Entity) SendTransaction(ctx context.Context, exchange string, amount int64) error {
	acc := cluster.accounts[rand.Intn(len(cluster.accounts))]
	deposit := cluster.deposits[exchange][rand.Intn(len(cluster.deposits))]

	return acc.SendTransaction(ctx, deposit, amount)
}

func (cluster *Entity) CreateExchangeAccounts(exchange *exchange.Exchange) error {
	for _, acc := range cluster.accounts {
		deposit, err := exchange.CreateAccountIfNotExist(acc.GetAddress())
		if err != nil {
			return err
		}

		cluster.deposits[exchange.GetName()] = append(cluster.deposits[exchange.GetName()], deposit)
	}

	return nil
}

func (cluster *Entity) GetAccounts() []*account.Account {
	return cluster.accounts
}
