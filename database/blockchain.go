package database

import (
	"context"
	"database/sql"
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
		return db.ExecuteTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
			err := nb.Blocks.AddBlocks(ctx, tx)
			if err != nil {
				return err
			}

			err = nb.Transactions.AddTransactions(ctx, tx)
			if err != nil {
				return err
			}

			err = nb.TokenTransfers.UpdateTokenTransfers(ctx, tx)
			if err != nil {
				return err
			}

			return nb.Approves.UpdateApproves(ctx, tx)
		})
	})
}
