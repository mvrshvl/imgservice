package database

import (
	"context"
	"fmt"
	"nir/amlerror"
)

const errTransfer = amlerror.AMLError("can't find tx from transfer")

type ExchangeTransfer struct {
	TxExchange string `db:"txExchange"`
	TxDeposit  string `db:"txDeposit"`
}

func (db *Database) AddExchangeTransfer(ctx context.Context, txExchangeHash, txDepositHash string) error {
	_, err := db.connection.ExecContext(ctx,
		`INSERT INTO exchangeTransfers(txDeposit, txExchange)
    			VALUES(?,?)`,
		txDepositHash, txExchangeHash)
	if err != nil {
		return fmt.Errorf("can't add exchange transfer: %w", err)
	}

	return nil
}

func (db *Database) GetTransferTxs(ctx context.Context, transfer *ExchangeTransfer) (txToDeposit, txToExchange *Transaction, err error) {
	query := `SELECT hash, nonce, blockNumber, fromAddress, toAddress, value, gas, gasPrice, input FROM transactions
				WHERE hash = ?
 				   OR hash = ?`

	rows, err := db.connection.QueryContext(ctx, query, transfer.TxDeposit, transfer.TxExchange)
	if err != nil {
		return nil, nil, err
	}

	txs, err := scanTransactions(rows)
	if err != nil {
		return nil, nil, err
	}

	if len(txs) != 2 {
		return nil, nil, fmt.Errorf("%w: toDeposit %s, toExchange %s", errTransfer, transfer.TxDeposit, transfer.TxExchange)
	}

	for _, tx := range txs {
		switch tx.Hash {
		case transfer.TxDeposit:
			txToDeposit = tx
		case transfer.TxExchange:
			txToExchange = tx
		}
	}

	return
}
