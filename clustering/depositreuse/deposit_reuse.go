package depositreuse

import (
	"context"
	"fmt"
	"nir/amlerror"
	"nir/database"
	"nir/di"
	logging "nir/log"
	"sync/atomic"
)

const errAccounts = amlerror.AMLError("can't get transfer accounts")

func Run(ctx context.Context, txs database.Transactions) error {
	exchangeTransfers, err := getExchangeTransfers(ctx, txs)
	if err != nil {
		return err
	}

	for _, transfer := range exchangeTransfers {
		err := clustering(ctx, transfer)
		if err != nil {
			return err
		}
	}

	return nil
}

func clustering(ctx context.Context, transfer *database.ExchangeTransfer) error {
	return di.FromContext(ctx).Invoke(func(db *database.Database) error {
		sender, deposit, innerErr := getSenderAndDeposit(ctx, db, transfer)
		if innerErr != nil {
			return innerErr
		}

		return findRelations(ctx, db, sender, deposit)
	})
}

func findRelations(ctx context.Context, db *database.Database, sender, deposit *database.Account) error {
	switch true {
	case sender.Cluster != nil && deposit.Cluster == nil:
		return includeDepositToCluster(ctx, db, sender, deposit)
	case sender.Cluster == nil && deposit.Cluster != nil:
		return includeSenderToCluster(ctx, db, sender, deposit)
	case sender.Cluster != nil && deposit.Cluster != nil && atomic.LoadUint64(sender.Cluster) != atomic.LoadUint64(deposit.Cluster):
		return db.MergeClusters(ctx, *deposit.Cluster, *sender.Cluster)
	case sender.Cluster != nil && deposit.Cluster != nil && atomic.LoadUint64(sender.Cluster) == atomic.LoadUint64(deposit.Cluster):
		return nil
	case sender.Cluster == nil && deposit.Cluster == nil:
		senders, innerErr := db.GetDepositSenders(ctx, deposit.Address, sender.Address)
		if innerErr != nil {
			return innerErr
		}

		if len(senders) == 0 {
			return nil
		}

		return createCluster(ctx, db, sender, deposit, senders)
	default:
		logging.Debugf(ctx, "Unexpected way to find relations: sender %+v, deposit %+v", sender, deposit)

		return nil
	}
}

func createCluster(ctx context.Context, db *database.Database, sender, deposit *database.Account, senders []string) error {
	deposits, innerErr := db.GetDepositsByAddresses(ctx, senders, deposit.Address)
	if innerErr != nil {
		return innerErr
	}

	id, innerErr := db.CreateCluster(ctx)
	if innerErr != nil {
		return innerErr
	}

	accountsToUpdate := append([]string{
		sender.Address,
		deposit.Address,
	}, deposits...)

	return db.UpdateClusterByAddress(ctx, uint64(id), append(accountsToUpdate, senders...)...)
}

func includeDepositToCluster(ctx context.Context, db *database.Database, sender, deposit *database.Account) error {
	senders, innerErr := db.GetDepositSenders(ctx, deposit.Address, sender.Address)
	if innerErr != nil {
		return innerErr
	}

	return db.UpdateClusterByAddress(ctx, *sender.Cluster, append(senders, deposit.Address)...)
}

func includeSenderToCluster(ctx context.Context, db *database.Database, sender, deposit *database.Account) error {
	deposits, innerErr := db.GetDepositsByAddresses(ctx, []string{sender.Address}, deposit.Address)
	if innerErr != nil {
		return innerErr
	}

	return db.UpdateClusterByAddress(ctx, *deposit.Cluster, append(deposits, sender.Address)...)
}

func getSenderAndDeposit(ctx context.Context, db *database.Database, transfer *database.ExchangeTransfer) (sender, deposit *database.Account, err error) {
	txToDeposit, _, innerErr := db.GetTransferTxs(ctx, transfer)
	if innerErr != nil {
		return nil, nil, innerErr
	}

	accounts, innerErr := db.GetAccounts(ctx, txToDeposit.FromAddress, txToDeposit.ToAddress)
	if innerErr != nil {
		return nil, nil, err
	}

	if len(accounts) != 2 {
		return nil, nil, fmt.Errorf("%w: sender address %s, deposit address %s", errAccounts, txToDeposit.FromAddress, txToDeposit.ToAddress)
	}

	for _, acc := range accounts {
		switch acc.Address {
		case txToDeposit.FromAddress:
			sender = acc
		case txToDeposit.ToAddress:
			deposit = acc
		}
	}

	return sender, deposit, nil
}

func getExchangeTransfers(ctx context.Context, txs database.Transactions) (exchangeTransfers []*database.ExchangeTransfer, err error) {
	err = di.FromContext(ctx).Invoke(func(db *database.Database) error {
		txsToExchange, innerErr := db.GetTxsToExchange(ctx, txs)
		if innerErr != nil {
			return innerErr
		}

		exchangeTransfers, innerErr = db.GetExchangeTransfer(ctx, txsToExchange, 10000, 1.5)

		return innerErr
	})

	return
}
