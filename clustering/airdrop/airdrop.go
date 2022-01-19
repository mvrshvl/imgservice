package airdrop

import (
	"context"
	"nir/clustering/common"
	"nir/database"
	"nir/di"
)

const (
	minAirdropAccounts = uint64(5)
	blockDiff          = uint64(1000)
)

func Run(ctx context.Context, toBlock uint64) error {
	airdrops, err := getAirdrops(ctx, common.GetFromBlock(toBlock, blockDiff), toBlock, minAirdropAccounts)
	if err != nil {
		return err
	}

	for _, airdrop := range airdrops {
		err = clustering(ctx, airdrop)
		if err != nil {
			return err
		}
	}

	return nil
}

func clustering(ctx context.Context, airdrop *database.Airdrop) error {
	return di.FromContext(ctx).Invoke(func(db *database.Database) error {
		txs, err := db.FindTransfersBetweenMembers(ctx, airdrop)
		if err != nil {
			return err
		}

		return common.Clustering(ctx, txs)
	})
}

func getAirdrops(ctx context.Context, fromBlock, toBlock, minSize uint64) (airdrops []*database.Airdrop, err error) {
	err = di.FromContext(ctx).Invoke(func(db *database.Database) (innerErr error) {
		airdrops, innerErr = db.GetAirdrops(ctx, fromBlock, toBlock, minSize)

		return innerErr
	})

	return
}
