package common

import (
	"context"
	"nir/database"
	"nir/di"
)

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

func GetFromBlock(toBlock, diffBlock uint64) uint64 {
	if toBlock > diffBlock {
		return diffBlock - toBlock
	}

	return 0
}

func Clustering(ctx context.Context, txs database.Transactions) error {
	if len(txs) == 0 {
		return nil
	}

	return di.FromContext(ctx).Invoke(func(db *database.Database) error {
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
