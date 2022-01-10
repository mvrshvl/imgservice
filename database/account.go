package database

import (
	"context"
	"fmt"
)

type AccountType string

const (
	eoa   AccountType = "eoa"
	miner AccountType = "miner"
)

type Account struct {
	address string      `db:"address"`
	accType AccountType `db:"accountType"`
	cluster uint64      `db:"cluster"`
}

func (db *Database) AddAccount(ctx context.Context, account *Account) error {
	_, err := db.connection.ExecContext(ctx,
		`INSERT IGNORE INTO accounts (address, accountType)
				VALUES (?, ?)`,
		account.address, account.accType)
	if err != nil {
		return fmt.Errorf("can't add account: %w", err)
	}

	return nil
}
