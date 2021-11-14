package user

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"math/rand"
	"nir/test/entity/account"
	"nir/test/exchange"
)

type User struct {
	accounts []*account.Account
	deposits map[string][]*common.Address
}

func CreateUserFromSize(size uint64) (*User, error) {
	user := &User{
		accounts: make([]*account.Account, size),
		deposits: make(map[string][]*common.Address),
	}

	for i := uint64(0); i < size; i++ {
		acc, err := account.CreateAccount()
		if err != nil {
			return nil, err
		}

		user.accounts[i] = acc
	}

	return user, nil
}

func (ent *User) SendTransaction(ctx context.Context, exchange string, amount int64) (*types.LegacyTx, error) {
	acc := ent.accounts[rand.Intn(len(ent.accounts))]
	deposit := ent.deposits[exchange][rand.Intn(len(ent.deposits[exchange]))]

	fmt.Println("TRANSFER", acc.GetAddress().String(), deposit.String())
	return acc.SendTransaction(ctx, deposit, amount, false)
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

func (ent *User) SendToken(ctx context.Context, exchange string, amount int64) (*types.LegacyTx, error) {
	acc := ent.accounts[rand.Intn(len(ent.accounts))]
	deposit := ent.deposits[exchange][rand.Intn(len(ent.deposits[exchange]))]

	fmt.Println("TRANSFER", acc.GetAddress().String(), deposit.String())
	return acc.SendTransaction(ctx, deposit, amount, false)
}

func (ent *User) CollectTokenOnOneAcc(ctx context.Context, mixDepth int, getBalance func(address *common.Address) (*big.Int, error), transferToken func(ctx context.Context, from *account.Account, toAddr *common.Address, amount *big.Int) error) error {
	if len(ent.accounts) == 1 {
		return nil
	}

	balances, err := ent.getTokenBalances(getBalance)
	if err != nil {
		return err
	}

	mainAccount := ent.randomAccount()

	err = ent.mixTokens(ctx, mainAccount, balances, mixDepth, transferToken)
	if err != nil {
		return err
	}

	for acc := range balances {
		if acc.GetAddress().String() == mainAccount.GetAddress().String() {
			continue
		}

		err := transferWithinCluster(ctx, balances, acc, mainAccount, transferToken)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ent *User) mixTokens(ctx context.Context, mainAccount *account.Account, balances map[*account.Account]int64, mixDepth int, transferToken func(ctx context.Context, from *account.Account, toAddr *common.Address, amount *big.Int) error) error {
	currentDepth := 0
	sum := int64(0)

	for _, balance := range balances {
		sum += balance
	}

	for balances[mainAccount] != sum || currentDepth < mixDepth-1 {
		for source, balance := range balances {
			target := ent.randomAccount()

			if source.GetAddress().String() == target.GetAddress().String() || balance == 0 {
				continue
			}

			fmt.Println("TRANSFER TOKEN", source.GetAddress().String(), target.GetAddress().String(), "MAIN ACC", mainAccount.GetAddress().String(), "DEPTH", currentDepth)
			err := transferWithinCluster(ctx, balances, source, target, transferToken)
			if err != nil {
				return err
			}
		}

		currentDepth++
	}

	return nil
}

func (ent *User) getTokenBalances(getBalance func(address *common.Address) (*big.Int, error)) (map[*account.Account]int64, error) {
	balances := make(map[*account.Account]int64)

	for _, acc := range ent.accounts {
		balance, err := getBalance(acc.GetAddress())
		if err != nil {
			return nil, err
		}
		balances[acc] = balance.Int64()
	}

	return balances, nil
}

func transferWithinCluster(ctx context.Context, balances map[*account.Account]int64, source, target *account.Account, transferToken func(ctx context.Context, from *account.Account, toAddr *common.Address, amount *big.Int) error) error {
	if balances[source] == 0 {
		return nil
	}

	err := transferToken(ctx, source, target.GetAddress(), big.NewInt(balances[source]))
	if err != nil {
		return err
	}

	balances[target] += balances[source]
	balances[source] = 0

	return nil
}

func (ent *User) randomAccount() *account.Account {
	return ent.accounts[rand.Intn(len(ent.accounts)-1)]
}
