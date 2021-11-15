package airdrop

import (
	"nir/clustering"
	"nir/clustering/blockchain"
)

const address0 = "0x0000000000000000000000000000000000000000"

//todo сделать для нескольких токенов
func Find(tokenTransfers blockchain.TokenTransfers) clustering.Clusters {
	owner := getOwner(tokenTransfers)

	ownerTransfers := getAccountsByTransfers(tokenTransfers, owner)

	accountsTransfers := getTransfersToAccounts(tokenTransfers, ownerTransfers)

	for target, sources := range accountsTransfers {
		copyAccountsTransfers := make(map[string]map[string]*blockchain.TokenTransfer)

		for targetCopy, sourcesCopy := range accountsTransfers {
			copyAccountsTransfers[targetCopy] = sourcesCopy
		}

		// todo сделать ранний выход по количеству входов
		for copyTarget, copySources := range copyAccountsTransfers {
			for source := range sources {
				if source != copyTarget {
					continue
				}

				for sourceCopySources, transferCopySources := range copySources {

					accountsTransfers[target][sourceCopySources] = transferCopySources
				}

				delete(copyAccountsTransfers, copyTarget)
			}
		}

		accountsTransfers = copyAccountsTransfers
	}

	return nil
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

func getTransfersToAccounts(tokenTransfers blockchain.TokenTransfers, accounts map[string]*blockchain.TokenTransfer) map[string]map[string]*blockchain.TokenTransfer {
	accountsTransfers := make(map[string]map[string]*blockchain.TokenTransfer)

	for distibutor := range accounts {
		accountsTransfers[distibutor] = make(map[string]*blockchain.TokenTransfer)
	}

	for distributor := range accounts {
		targetsTransfers := getAccountsByTransfers(tokenTransfers, distributor)
		for target, tr := range targetsTransfers {
			if _, ok := accountsTransfers[target]; !ok {
				accountsTransfers[target] = make(map[string]*blockchain.TokenTransfer)
			}

			accountsTransfers[target][distributor] = tr
		}
	}

	return accountsTransfers
}
