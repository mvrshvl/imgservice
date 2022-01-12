package user

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"math/rand"
	"nir/test/contract"
	"nir/test/entity/account"
	"nir/test/exchange"
	"nir/test/writer"
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

	//fmt.Println("TRANSFER", acc.GetAddress().String(), deposit.String())
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

	mainAccount := ent.RandomAccount()

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
			target := ent.RandomAccount()

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

func (ent *User) RandomAccount() *account.Account {
	if len(ent.accounts) == 1 {
		return ent.accounts[0]
	}

	return ent.accounts[rand.Intn(len(ent.accounts)-1)]
}

func (ent *User) DeployContract(ctx context.Context, tokens int) (*contract.SimpleToken, error) {
	distributor := ent.RandomAccount()

	bytecode, err := contract.BytecodeContract(contract.SimpleTokenBin, contract.SimpleTokenABI, "test", "t", big.NewInt(int64(tokens)))
	if err != nil {
		return nil, err
	}

	var gas uint64

	err = writer.Execute(ctx, func(w *writer.Writer) (innerErr error) {
		gas, innerErr = w.EstimateGas(ctx, *distributor.GetAddress(), nil, bytecode)

		return innerErr
	})
	if err != nil {
		return nil, err
	}

	err = writer.Execute(ctx, func(w *writer.Writer) error {
		b, innerErr := w.BalanceAt(ctx, *distributor.GetAddress())
		if innerErr != nil {
			return innerErr
		}

		fmt.Println("Distributor balance before deploy", b.Int64())

		return nil
	})
	if err != nil {
		return nil, err
	}

	var token *contract.SimpleToken

	err = executeToken(ctx, distributor, gas, func(auth *bind.TransactOpts, backend bind.ContractBackend) (tx *types.Transaction, innerErr error) {
		_, tx, token, innerErr = contract.DeploySimpleToken(auth, backend, "TestContract", "TC", big.NewInt(100000000000))

		return tx, innerErr
	})
	if err != nil {
		return nil, err
	}

	for _, acc := range ent.accounts {
		if distributor.GetAddress().String() == acc.GetAddress().String() {
			continue
		}

		err = executeToken(ctx, distributor, gas, func(auth *bind.TransactOpts, backend bind.ContractBackend) (*types.Transaction, error) {
			return token.Approve(auth, *acc.GetAddress(), big.NewInt(1))
		})
		if err != nil {
			return nil, err
		}
	}

	return token, err
}

func executeToken(ctx context.Context, owner *account.Account, gas uint64, fn func(auth *bind.TransactOpts, backend bind.ContractBackend) (*types.Transaction, error)) error {
	tx, err := owner.ExecuteContract(ctx, gas, fn)
	if err != nil {
		return err
	}

	return writer.Execute(ctx, func(w *writer.Writer) error {
		return w.WaitTx(ctx, tx.Hash())
	})
}

func (ent *User) ExecuteContract(ctx context.Context, gasLimit uint64, fn func(auth *bind.TransactOpts, backend bind.ContractBackend) (*types.Transaction, error)) (tx *types.Transaction, err error) {
	acc := ent.RandomAccount()

	return acc.ExecuteContract(ctx, gasLimit, fn)
}
