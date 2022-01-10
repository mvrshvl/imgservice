package transfer

//import (
//	"nir/database"
//)

//func GetExchangeTransfers(ctx context.Context, txs database.Transactions, blockDiff uint64) []*ExchangeTransfer {
//	err := di.FromContext(ctx).Invoke(func(db *database.Database) error {
//		transfersToExchange, innerErr := db.GetTxsToExchange(ctx, txs)
//		if innerErr != nil {
//			return innerErr
//		}
//
//
//	})
//	if err != nil {
//		return nil, err
//	}
//	txsToExchanges := chain.Transactions.GetTransactionsToAddresses(chain.Exchanges.MapAddresses())
//	txsToDeposits := chain.Transactions.GetTransactionsToAddresses(txsToExchanges.MapFromAddresses())
//
//	return mergeTransactions(txsToExchanges, txsToDeposits, blockDiff)
//}
//
//func mergeTransactions(txsToExchange database.Transactions, txsToDeposit database.Transactions, maxBlockDiff uint64) []*ExchangeTransfer {
//	var exchangeTransfers []*ExchangeTransfer
//
//	for _, txToExchange := range txsToExchange {
//		for _, txToDeposit := range txsToDeposit {
//			blockDiff := txToExchange.BlockNumber - txToDeposit.BlockNumber
//
//			if txToExchange.FromAddress == txToDeposit.ToAddress &&
//				txToExchange.Value == txToDeposit.Value && // is always equal?
//				blockDiff > 0 && blockDiff < maxBlockDiff {
//				exchangeTransfers = append(exchangeTransfers, &ExchangeTransfer{
//					TxToExchange: txToExchange,
//					TxToDeposit:  txToDeposit,
//				})
//			}
//
//		}
//	}
//
//	return exchangeTransfers
//}
