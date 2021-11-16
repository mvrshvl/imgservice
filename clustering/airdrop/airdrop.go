package airdrop

import (
	"fmt"
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
		copyAccountsTransfers := make(map[string]map[string]*blockchain.TokenTransfer)

		cluster := clustering.NewCluster()

		AddTransfersToCluster(cluster, sources)

		for targetCopy, sourcesCopy := range accountsTransfers {
			copyAccountsTransfers[targetCopy] = sourcesCopy
		}

		// todo сделать ранний выход по количеству входов
		for copyTarget, copySources := range copyAccountsTransfers {
			if copyTarget == target {
				continue
			}

			for source := range sources {
				fmt.Println("NIL CHECK", sources, copySources)

				if source != copyTarget {
					continue
				}

				for sourceCopySources, transferCopySources := range copySources {
					AddTransferToClusterAccount(cluster, copyTarget, transferCopySources)

					accountsTransfers[target][sourceCopySources] = transferCopySources
				}

				delete(accountsTransfers, copyTarget)
			}
		}

		clusters = append(clusters, cluster)
	}

	return clusters
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
