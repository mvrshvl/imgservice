package airdrop

import (
	"nir/clustering"
	"nir/clustering/blockchain"
	"nir/clustering/transfer"
)

const address0 = "0x0000000000000000000000000000000000000000"

//todo сделать для нескольких токенов
func Find(tokenTransfers blockchain.TokenTransfers) (clusters clustering.Clusters) {
	owner := getOwner(tokenTransfers)

	ownerTransfers := getAccountsByTransfers(tokenTransfers, owner)

	accountsTransfers := getAirdropAccountsWithTransfers(tokenTransfers, ownerTransfers)

	for target, sources := range accountsTransfers {
		cluster := clustering.NewCluster()

		AddTransfersToCluster(cluster, sources)

		search(accountsTransfers, target, sources, cluster)

		//for source := range sources {
		//	if source == target {
		//		continue
		//	}
		//
		//	copySources, ok := accountsTransfers[source]
		//	if !ok{
		//		continue
		//	}
		//
		//	for sourceCopySources, transferCopySources := range copySources {
		//		AddTransferToClusterAccount(cluster, source, transferCopySources)
		//
		//		accountsTransfers[target][sourceCopySources] = transferCopySources
		//	}
		//
		//	delete(accountsTransfers, source)
		//}

		if len(cluster.AccountsExchangeTransfers) >= 2 {
			clusters = append(clusters, cluster)
		}

	}

	return clusters
}

func search(accountsTransfers map[string]map[string]*blockchain.TokenTransfer, target string, sources map[string]*blockchain.TokenTransfer, cluster *clustering.Cluster) {
	for source := range sources {
		if source == target {
			continue
		}

		copySources, ok := accountsTransfers[source]
		if !ok {
			continue
		}

		AddTransfersToCluster(cluster, copySources)
		search(accountsTransfers, source, copySources, cluster)
		//for sourceCopySources, transferCopySources := range copySources {
		//	AddTransferToClusterAccount(cluster, source, transferCopySources)
		//
		//	accountsTransfers[target][sourceCopySources] = transferCopySources
		//}

		delete(accountsTransfers, source)
	}
}
func AddTransferToClusterAccount(cluster *clustering.Cluster, account string, t *blockchain.TokenTransfer) {
	cluster.AccountsTokenTransfers[account] = append(cluster.AccountsTokenTransfers[account], &transfer.TokenTransfer{
		TokenAddress: t.ContractAddress,
		FromAddress:  t.SourceAddress,
		ToAddress:    t.TargetAddress,
		Value:        t.Value,
	})
}

func AddTransfersToCluster(cluster *clustering.Cluster, ts map[string]*blockchain.TokenTransfer) {
	for _, t := range ts {
		AddTransferToClusterAccount(cluster, t.TargetAddress, t)
	}
}

// todo сделать на выходе нормальные связи ?
func getOwner(tokenTransfers blockchain.TokenTransfers) string {
	for _, tokenTransfer := range tokenTransfers {
		if tokenTransfer.SourceAddress == address0 {
			return tokenTransfer.TargetAddress
		}
	}

	return ""
}

func getAccountsByTransfers(tokenTransfers blockchain.TokenTransfers, distributor string) map[string]*blockchain.TokenTransfer {
	addresses := make(map[string]*blockchain.TokenTransfer)

	for _, tokenTransfer := range tokenTransfers {
		if tokenTransfer.SourceAddress == distributor {
			addresses[tokenTransfer.TargetAddress] = tokenTransfer
		}
	}

	return addresses
}

func getAirdropAccountsWithTransfers(tokenTransfers blockchain.TokenTransfers, airdropAccounts map[string]*blockchain.TokenTransfer) map[string]map[string]*blockchain.TokenTransfer {
	accountsTransfers := make(map[string]map[string]*blockchain.TokenTransfer)

	for acc := range airdropAccounts {
		accountsTransfers[acc] = map[string]*blockchain.TokenTransfer{airdropAccounts[acc].SourceAddress: airdropAccounts[acc]} //add transfer from distributor
	}

	for acc := range airdropAccounts {
		targetsTransfers := getAccountsByTransfers(tokenTransfers, acc)
		for target, tr := range targetsTransfers {
			accountsTransfers[target][acc] = tr
		}
	}

	return accountsTransfers
}
