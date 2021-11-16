package depositreuse

import (
	"nir/clustering"
	"nir/clustering/transfer"
)

func Find(transfers []*transfer.ExchangeTransfer) clustering.Clusters {
	depositsWithSenders := make(map[string]*clustering.Cluster)

	// Поиск депозитов и всех аккаунтов которые отправляли транзакции на данный депозит
	for _, t := range transfers {
		if depositsWithSenders[t.TxToDeposit.ToAddress] == nil {
			depositsWithSenders[t.TxToDeposit.ToAddress] = clustering.NewCluster()
		}

		AddTransfer(depositsWithSenders[t.TxToDeposit.ToAddress], t)
	}

	var clusters []*clustering.Cluster

	for deposit, cluster := range depositsWithSenders {
		MergeMatchesExchangeTransfersAccounts(cluster, deposit, depositsWithSenders)
		clusters = append(clusters, cluster)
	}

	return clusters
}

func AddTransfer(cluster *clustering.Cluster, transfer *transfer.ExchangeTransfer) {
	for _, ts := range cluster.AccountsExchangeTransfers[transfer.TxToDeposit.FromAddress] {
		if ts.TxToDeposit.Hash == transfer.TxToDeposit.Hash {
			return
		}
	}

	cluster.Accounts[transfer.TxToDeposit.FromAddress] = struct{}{}

	cluster.AccountsExchangeTransfers[transfer.TxToDeposit.FromAddress] = append(cluster.AccountsExchangeTransfers[transfer.TxToDeposit.FromAddress], transfer)
}

func AddTransfers(cluster *clustering.Cluster, transfers []*transfer.ExchangeTransfer) {
	for _, t := range transfers {
		AddTransfer(cluster, t)
	}
}

func HasAnAccounts(cluster *clustering.Cluster, accs map[string][]*transfer.ExchangeTransfer) bool {
	for acc := range accs {
		if _, ok := cluster.AccountsExchangeTransfers[acc]; ok {
			return true
		}
	}

	return false
}

// MergeMatchesExchangeTransfersAccounts Добавляет в кластер А транзакции кластера Б, если хотя бы один аккаунт кластера Б существуют в кластере А
func MergeMatchesExchangeTransfersAccounts(cluster *clustering.Cluster, currentDeposit string, depositsWithSenders map[string]*clustering.Cluster) { // todo rewrite this
	for deposit, c := range depositsWithSenders {
		if !HasAnAccounts(cluster, c.AccountsExchangeTransfers) || currentDeposit == deposit {
			continue
		}

		for _, transfers := range c.AccountsExchangeTransfers {
			AddTransfers(cluster, transfers)
		}

		delete(depositsWithSenders, deposit)
	}
}
