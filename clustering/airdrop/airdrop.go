package airdrop

import (
	"nir/amlerror"
	"nir/clustering"
	"nir/clustering/blockchain"
	"nir/clustering/transfer"
)

const (
	address0            = "0x0000000000000000000000000000000000000000"
	errRecursiveCounter = amlerror.AMLError("Recursion exceeded the allowed rate")
)

func Find(tokenTransfers blockchain.TokenTransfers) (clusters clustering.Clusters, err error) {
	owners := GetOwners(tokenTransfers)

	for _, owner := range owners {
		ownerTransfers := getAccountsByTransfers(tokenTransfers, owner)

		remainingAccountsTransfers := getAirdropAccountsWithTransfers(tokenTransfers, ownerTransfers)

		for target, sources := range remainingAccountsTransfers {
			cluster := clustering.NewCluster()

			AddTransfersToCluster(cluster, sources)

			err = merge(remainingAccountsTransfers, 0, target, sources, cluster)
			if err != nil {
				return nil, err
			}

			if len(cluster.AccountsTokenTransfers) >= 2 {
				clusters = append(clusters, cluster)
			}

		}

		for _, ts := range remainingAccountsTransfers {
			cluster := clustering.NewCluster()
			AddTransfersToCluster(cluster, ts)

			clusters = append(clusters, cluster)
		}
	}

	return
}

func merge(accountsTransfers map[string]map[string]*blockchain.TokenTransfer, counter uint64, target string, sources map[string]*blockchain.TokenTransfer, cluster *clustering.Cluster) error {
	if counter == 100000 {
		return errRecursiveCounter
	}

	for source := range sources {
		copySources, ok := accountsTransfers[source]
		if !ok {
			continue
		}

		AddTransfersToCluster(cluster, copySources)

		delete(accountsTransfers, target)
		delete(accountsTransfers, source)

		err := merge(accountsTransfers, counter+1, source, copySources, cluster)
		if err != nil {
			return err
		}
	}

	return nil
}

func AddTransferToClusterAccount(cluster *clustering.Cluster, account string, t *blockchain.TokenTransfer) {
	cluster.Accounts[account] = struct{}{}

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

func GetOwners(tokenTransfers blockchain.TokenTransfers) (owners []string) {
	for _, tokenTransfer := range tokenTransfers {
		if tokenTransfer.SourceAddress == address0 {
			owners = append(owners, tokenTransfer.TargetAddress)
		}
	}

	return owners
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
