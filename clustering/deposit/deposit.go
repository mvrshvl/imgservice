package deposit

import (
	"nir/clustering/blockchain"
)

type ExchangeTransfer struct {
	txToExchange *blockchain.Transaction
	txToDeposit  *blockchain.Transaction
}

func GetExchangeTransfers(chain *blockchain.Blockchain, blockDiff uint64) []*ExchangeTransfer {
	txsToExchanges := chain.Transactions.GetTransactionsToAddresses(chain.Exchanges.MapAddresses())
	txsToDeposits := chain.Transactions.GetTransactionsToAddresses(txsToExchanges.MapFromAddresses())

	return mergeTransactions(txsToExchanges, txsToDeposits, blockDiff)
}

func mergeTransactions(txsToExchange blockchain.Transactions, txsToDeposit blockchain.Transactions, maxBlockDiff uint64) []*ExchangeTransfer {
	var exchangeTransfers []*ExchangeTransfer

	for _, txToExchange := range txsToExchange {
		for _, txToDeposit := range txsToDeposit {
			blockDiff := txToExchange.BlockNumber - txToDeposit.BlockNumber

			if txToExchange.FromAddress == txToDeposit.ToAddress &&
				txToExchange.Value == txToDeposit.Value && // is always equal?
				blockDiff > 0 && blockDiff < maxBlockDiff {
				exchangeTransfers = append(exchangeTransfers, &ExchangeTransfer{
					txToExchange: txToExchange,
					txToDeposit:  txToDeposit,
				})
			}

		}
	}

	return exchangeTransfers
}
