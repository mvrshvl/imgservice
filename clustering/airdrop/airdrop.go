package airdrop

import (
	"context"
	"nir/database"
	"nir/di"
)

const (
	minAirdropAccounts = uint64(5)
	blockDiff          = uint64(1000)
)

func Run(ctx context.Context, toBlock uint64) error {
	airdrops, err := getAirdrops(ctx, getFromBlock(toBlock, blockDiff), toBlock, minAirdropAccounts)
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

		if len(txs) == 0 {
			return nil
		}

		for _, tx := range txs {
			sender, receiver, err := db.GetSenderAndReceiver(ctx, tx.FromAddress, tx.ToAddress)
			if err != nil {
				return err
			}

			err = db.UpdateCluster(ctx, sender, receiver, clusteringBySender, clusteringByReceiver, createCluster)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func includeAccountToCluster(ctx context.Context, db *database.Database, toInclude *database.Account, byInclude *database.Account) error {
	deposits, innerErr := db.GetDepositsByAddresses(ctx, []string{toInclude.Address})
	if innerErr != nil {
		return innerErr
	}

	return db.UpdateClusterByAddress(ctx, *byInclude.Cluster, append(deposits, toInclude.Address)...)
}

func clusteringBySender(ctx context.Context, db *database.Database, from, to *database.Account) error {
	return includeAccountToCluster(ctx, db, to, from)
}

func clusteringByReceiver(ctx context.Context, db *database.Database, from, to *database.Account) error {
	return includeAccountToCluster(ctx, db, from, to)
}

func createCluster(ctx context.Context, db *database.Database, sender, receiver *database.Account) error {
	id, err := db.CreateCluster(ctx)
	if err != nil {
		return err
	}

	byInclude := &database.Account{
		Cluster: &id,
	}

	err = includeAccountToCluster(ctx, db, sender, byInclude)
	if err != nil {
		return err
	}

	return includeAccountToCluster(ctx, db, receiver, byInclude)
}

func getAirdrops(ctx context.Context, fromBlock, toBlock, minSize uint64) (airdrops []*database.Airdrop, err error) {
	err = di.FromContext(ctx).Invoke(func(db *database.Database) (innerErr error) {
		airdrops, innerErr = db.GetAirdrops(ctx, fromBlock, toBlock, minSize)

		return innerErr
	})

	return
}

func getFromBlock(toBlock, diffBlock uint64) uint64 {
	if toBlock > diffBlock {
		return diffBlock - toBlock
	}

	return 0
}
