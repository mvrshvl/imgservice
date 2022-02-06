package database

import (
	"context"
	"fmt"
)

type Entity struct {
	Name     string
	Accounts []*Account
}

func (db *Database) GetEntity(ctx context.Context, address string) (*Entity, error) {
	acc, err := db.GetAccount(ctx, address)
	if err != nil {
		return nil, err
	}

	entity := &Entity{}

	if *acc.Cluster == 0 {
		entity.Name = "Single account"

		return entity, nil
	}

	entity.Accounts, err = db.getEntityAccounts(ctx, *acc.Cluster)
	if err != nil {
		return nil, fmt.Errorf("get enity accounts failed: %w", err)
	}

	return entity, err
}

func (db *Database) getEntityAccounts(ctx context.Context, cluster uint64) ([]*Account, error) {
	query := `SELECT * FROM accounts
				WHERE cluster = ?`

	rows, err := db.connection.QueryContext(ctx, query, cluster)
	if err != nil {
		return nil, fmt.Errorf("can't get cluster accounts: %w", err)
	}

	defer rows.Close()

	return scanAccounts(rows)
}
