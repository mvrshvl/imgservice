package selfauth

import (
	"context"
	"nir/clustering/common"
	"nir/database"
	"nir/di"
)

const (
	maxApproves = uint64(10)
	blockDiff   = uint64(1000)
)

func Find(ctx context.Context, toBlock uint64) error {
	txs, err := getSelfApproves(ctx, common.GetFromBlock(toBlock, blockDiff), toBlock, maxApproves)
	if err != nil {
		return err
	}

	return common.Clustering(ctx, txs)
}

func getSelfApproves(ctx context.Context, fromBlock uint64, toBlock uint64, maxApproves uint64) (txs database.Transactions, err error) {
	err = di.FromContext(ctx).Invoke(func(db *database.Database) (innerErr error) {
		txs, innerErr = db.GetSelfApproveTxs(ctx, fromBlock, toBlock, maxApproves)

		return innerErr
	})

	return txs, err
}
