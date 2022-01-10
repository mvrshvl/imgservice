package depositreuse

import (
	"context"
	"fmt"
	"nir/amlerror"
	"nir/database"
	"nir/di"
	logging "nir/log"
)

const errAccounts = amlerror.AMLError("can't get transfer accounts")

func Run(ctx context.Context, txs database.Transactions) error {
	exchangeTransfers, err := getExchangeTransfers(ctx, txs)
	if err != nil {
		return err
	}

	for _, transfer := range exchangeTransfers {

	}
	return nil
	//depositsWithSenders := make(map[string]*clustering.Cluster)
	//
	//// Поиск депозитов и всех аккаунтов которые отправляли транзакции на данный депозит
	//for _, t := range transfers {
	//	if depositsWithSenders[t.TxToDeposit.ToAddress] == nil {
	//		depositsWithSenders[t.TxToDeposit.ToAddress] = clustering.NewCluster()
	//	}
	//
	//	AddTransfer(depositsWithSenders[t.TxToDeposit.ToAddress], t)
	//}
	//
	//var clusters []*clustering.Cluster
	//
	//for deposit, cluster := range depositsWithSenders {
	//	MergeMatchesExchangeTransfersAccounts(cluster, deposit, depositsWithSenders)
	//	clusters = append(clusters, cluster)
	//}
	//
	//return clusters
}

func clustering(ctx context.Context, transfer *database.ExchangeTransfer) error {
	return di.FromContext(ctx).Invoke(func(db *database.Database) error {
		txToDeposit, _, innerErr := db.GetTransferTxs(ctx, transfer)
		if innerErr != nil {
			return innerErr
		}

		sender, deposit, innerErr := getSenderAndDeposit(ctx, db, txToDeposit.FromAddress, txToDeposit.ToAddress)

		switch true {
		case sender.Cluster != nil && deposit.Cluster == nil:
			senders, innerErr := db.GetSenders(ctx, deposit.Address, sender.Address)
			if innerErr != nil {
				return innerErr
			}

			if len(senders) > 1 {
				logging.Debugf(ctx, "should be one sender. Senders %v, deposit %s", senders, deposit.Address)
			}

			// update cluster for deposit and his sender
			return db.UpdateClusterByAddress(ctx, *sender.Cluster, append(senders, deposit.Address)...)
		case sender.Cluster == nil && deposit.Cluster != nil:
			deposits, innerErr := db.GetDeposits(ctx, sender.Address, deposit.Address)
			if innerErr != nil {
				return innerErr
			}
			// update cluster for sender and his deposits
			return db.UpdateClusterByAddress(ctx, *deposit.Cluster, append(deposits, sender.Address)...)
		case sender.Cluster != nil && deposit.Cluster != nil && *sender.Cluster != *deposit.Cluster:
			return db.MergeClusters(ctx, *deposit.Cluster, *sender.Cluster)
		//case sender.Cluster != nil && deposit.Cluster != nil && *sender.Cluster == *deposit.Cluster:
		//	return nil
		case sender.Cluster == nil && deposit.Cluster == nil:
			senders, innerErr := db.GetSenders(ctx, deposit.Address, sender.Address)
			if innerErr != nil {
				return innerErr
			}

			if len(senders) == 0 {
				return nil
			}

			id, innerErr := db.CreateCluster(ctx)
			if innerErr != nil {
				return innerErr
			}

			// todo set all senders to cluster and their deposits
		}

		return innerErr
	})
}

func getSenderAndDeposit(ctx context.Context, db *database.Database, senderAddr, depositAddr string) (sender, deposit *database.Account, err error) {
	accounts, innerErr := db.GetAccounts(ctx, senderAddr, depositAddr)
	if innerErr != nil {
		return nil, nil, err
	}

	if len(accounts) != 2 {
		return nil, nil, fmt.Errorf("%w: sender address %s, deposit address %s", errAccounts, senderAddr, depositAddr)
	}
	for _, acc := range accounts {
		switch acc.Address {
		case senderAddr:
			sender = acc
		case depositAddr:
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

//func AddTransfer(cluster *clustering.Cluster, transfer *transfer.ExchangeTransfer) {
//	for _, ts := range cluster.AccountsExchangeTransfers[transfer.TxToDeposit.FromAddress] {
//		if ts.TxToDeposit.Hash == transfer.TxToDeposit.Hash {
//			return
//		}
//	}
//
//	cluster.Accounts[transfer.TxToDeposit.FromAddress] = struct{}{}
//
//	cluster.AccountsExchangeTransfers[transfer.TxToDeposit.FromAddress] = append(cluster.AccountsExchangeTransfers[transfer.TxToDeposit.FromAddress], transfer)
//}
//
//func AddTransfers(cluster *clustering.Cluster, transfers []*transfer.ExchangeTransfer) {
//	for _, t := range transfers {
//		AddTransfer(cluster, t)
//	}
//}
//
//func HasAnAccounts(cluster *clustering.Cluster, accs map[string][]*transfer.ExchangeTransfer) bool {
//	for acc := range accs {
//		if _, ok := cluster.AccountsExchangeTransfers[acc]; ok {
//			return true
//		}
//	}
//
//	return false
//}
//
//// MergeMatchesExchangeTransfersAccounts Добавляет в кластер А транзакции кластера Б, если хотя бы одна транзакция к Exchange от одного аккаунта кластера Б существуют в кластере А
//func MergeMatchesExchangeTransfersAccounts(cluster *clustering.Cluster, currentDeposit string, depositsWithSenders map[string]*clustering.Cluster) { // todo rewrite this
//	for deposit, c := range depositsWithSenders {
//		if !HasAnAccounts(cluster, c.AccountsExchangeTransfers) || currentDeposit == deposit {
//			continue
//		}
//
//		for _, transfers := range c.AccountsExchangeTransfers {
//			AddTransfers(cluster, transfers)
//		}
//
//		delete(depositsWithSenders, deposit)
//	}
//}
