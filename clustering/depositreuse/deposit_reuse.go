package depositreuse

import (
	"nir/clustering"
	"nir/clustering/transfer"
)

func Find(transfers []*transfer.ExchangeTransfer) clustering.Clusters {
	depositsWithSenders := make(map[string]*clustering.Cluster)

	for _, t := range transfers {
		if depositsWithSenders[t.TxToDeposit.ToAddress] == nil {
			depositsWithSenders[t.TxToDeposit.ToAddress] = clustering.NewCluster()
		}

		depositsWithSenders[t.TxToDeposit.ToAddress].AddTransfer(t)
	}

	noMatches := make(map[string]*clustering.Cluster)
	for deposit, cluster := range depositsWithSenders {
		noMatches[deposit] = cluster
	}

	for deposit, cluster := range noMatches {
		cluster.MergeMatches(deposit, noMatches)
	}

	var clusters []*clustering.Cluster
	for _, cluster := range noMatches {
		clusters = append(clusters, cluster)
	}

	return clusters
}
