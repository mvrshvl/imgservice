package depositreuse

import (
	"context"
	"nir/database"
	"nir/di"
)

func Run(ctx context.Context, txs database.Transactions) error {
	_, err := getExchangeTransfers(ctx, txs)
	if err != nil {
		return err
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
