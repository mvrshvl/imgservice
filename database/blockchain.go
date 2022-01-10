package database

import (
	"context"
	"nir/di"
)

type NewBlocks struct {
	Transactions   Transactions
	Blocks         Blocks
	Exchanges      Exchanges
	TokenTransfers TokenTransfers
	Approves       ERC20Approves
}

func GetNewBlocks(transactions []*Transaction, blocks []*Block, exchanges []*Exchange, tokenTransfers TokenTransfers, logs Logs) (*NewBlocks, error) {
	approves, err := logs.ToApproves(transactions)
	if err != nil {
		return nil, err
	}

	return &NewBlocks{
		Transactions:   transactions,
		Blocks:         blocks,
		Exchanges:      exchanges,
		TokenTransfers: tokenTransfers,
		Approves:       approves,
	}, nil
}

func (nb *NewBlocks) Save(ctx context.Context) error {
	return di.FromContext(ctx).Invoke(func(db *Database) error {
		err := db.AddBlocks(ctx, nb.Blocks)
		if err != nil {
			return err
		}

		err = db.AddTransactions(ctx, nb.Transactions)
		if err != nil {
			return err
		}

		err = db.AddExchanges(ctx, nb.Exchanges)
		if err != nil {
			return err
		}

		err = db.UpdateTokenTransfers(ctx, nb.TokenTransfers)
		if err != nil {
			return err
		}

		err = db.UpdateApproves(ctx, nb.Approves)
		if err != nil {
			return err
		}

		return nil
	})
}
