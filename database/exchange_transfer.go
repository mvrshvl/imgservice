package database

import (
	"context"
	"fmt"
)

type ExchangeTransfer struct {
	TxExchange string `db:"txExchange"`
	TxDeposit  string `db:"txDeposit"`
}

func (db *Database) AddExchangeTransfer(ctx context.Context, txExchangeHash, txDepositHash string) error {
	_, err := db.connection.ExecContext(ctx,
		`INSERT INTO transactions(txDeposit, txExchange)
    			VALUES(?,?)`,
		txDepositHash, txExchangeHash)

	if err != nil {
		return fmt.Errorf("can't add exchange transfer: %w", err)
	}

	return nil
}
