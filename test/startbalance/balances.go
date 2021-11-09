package startbalance

import (
	"context"
	"log"
	"nir/amlerror"
	"nir/test/entity/account"
)

const errAccount = amlerror.AMLError("account with common balance is missing")

type CommonBalances struct {
	balances []*CommonBalance
	taskPool chan func(acc *account.Account)
}

type CommonBalance struct {
	account *account.Account
}

func newCommonBalance(taskPool chan func(acc *account.Account), key string) (*CommonBalance, error) {
	acc, err := account.NewAccountFromKey(key)
	if err != nil {
		return nil, err
	}

	cb := &CommonBalance{
		account: acc,
	}

	go cb.start(taskPool)

	return cb, nil
}

func (cb *CommonBalance) start(taskPool chan func(acc *account.Account)) {
	for task := range taskPool {
		task(cb.account)
	}
}

type accountWithCommonBalanceKey struct{}

func CommonBalancesWithCtx(ctx context.Context, keys []string) (context.Context, error) {
	accs := make([]*CommonBalance, len(keys))

	taskPool := make(chan func(acc *account.Account), 100000)
	for i, key := range keys {
		acc, err := newCommonBalance(taskPool, key)
		if err != nil {
			return nil, err
		}

		accs[i] = acc
	}

	return context.WithValue(ctx, accountWithCommonBalanceKey{}, &CommonBalances{
		balances: accs,
		taskPool: taskPool,
	}), nil
}

func AddTask(ctx context.Context, fn func(acc *account.Account)) error {
	accs, ok := ctx.Value(accountWithCommonBalanceKey{}).(*CommonBalances)
	if !ok {
		return errAccount
	}

	select {
	case accs.taskPool <- fn:
	default:
		log.Println("TASKPOOL IS OVER")
	}

	return nil
}

func Close(ctx context.Context) error {
	accs, ok := ctx.Value(accountWithCommonBalanceKey{}).(*CommonBalances)
	if !ok {
		return errAccount
	}

	close(accs.taskPool)

	return nil
}
