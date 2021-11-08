package user

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"math/rand"
	"nir/test/entity/account"
	"nir/test/exchange"
)

type User struct {
	accounts []*account.Account
	deposits map[string][]*common.Address
}

func CreateEntity(size uint64) (*User, error) {
	cluster := &User{
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

func (ent *User) SendTransaction(ctx context.Context, exchange string, amount int64) (*common.Address, error) {
	acc := ent.accounts[rand.Intn(len(ent.accounts))]
	deposit := ent.deposits[exchange][rand.Intn(len(ent.deposits))]

	return acc.SendTransaction(ctx, deposit, amount)
}

func (ent *User) CreateExchangeAccounts(exchange *exchange.Exchange) error {
	for _, acc := range ent.accounts {
		deposit, err := exchange.CreateAccountIfNotExist(acc.GetAddress())
		if err != nil {
			return err
		}

		ent.deposits[exchange.GetName()] = append(ent.deposits[exchange.GetName()], deposit)
	}

	return nil
}

func (ent *User) GetAccounts() (addresses []*common.Address) {
	for _, acc := range ent.accounts {
		addresses = append(addresses, acc.GetAddress())
	}

	return addresses
}
