package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type AccountType string

const (
	eoa      AccountType = "eoa"
	miner    AccountType = "miner"
	deposit  AccountType = "deposit"
	exchange AccountType = "exchange"
)

type Account struct {
	Address string      `db:"Address"`
	AccType AccountType `db:"accountType"`
	Cluster *uint64     `db:"Cluster"`
}

func (db *Database) AddAccount(ctx context.Context, account *Account) error {
	_, err := db.connection.ExecContext(ctx,
		`INSERT IGNORE INTO accounts (Address, accountType)
				VALUES (?, ?)`,
		account.Address, account.AccType)
	if err != nil {
		return fmt.Errorf("can't add account: %w", err)
	}

	return nil
}

func (db *Database) UpdateAccountType(ctx context.Context, address string, accType AccountType) error {
	_, err := db.connection.ExecContext(ctx,
		`UPDATE accounts SET accountType = ? WHERE address = ?`,
		accType, address)
	if err != nil {
		return fmt.Errorf("can't update account type: %w", err)
	}

	return nil
}

func (db *Database) GetAccounts(ctx context.Context, addresses ...string) ([]*Account, error) {
	query := `SELECT * FROM accounts
				WHERE address IN ( ? )`

	queryIn, args, err := sqlx.In(query, addresses)
	if err != nil {
		return nil, fmt.Errorf("can't create IN QUERY: %w", err)
	}

	rows, err := db.connection.QueryContext(ctx, queryIn, args...)
	if err != nil {
		return nil, fmt.Errorf("can't get exchange transfers: %w", err)
	}

	defer rows.Close()

	return scanAccounts(rows)
}

func scanAccounts(rows *sql.Rows) ([]*Account, error) {
	var accounts []*Account

	for rows.Next() {
		var acc Account

		err := rows.Scan(
			&acc.Address,
			&acc.AccType,
			&acc.Cluster,
		)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, &acc)
	}

	return accounts, nil
}

func (db *Database) GetSenders(ctx context.Context, address string, excludeAddresses ...string) ([]string, error) {
	query := `SELECT fromAddress FROM transactions
				WHERE toAddress = ?`

	return db.getAddresses(ctx, query, excludeAddresses, address)
}

func (db *Database) GetDeposits(ctx context.Context, address string, excludeAddresses ...string) ([]string, error) {
	query := `SELECT toAddress FROM transactions
				WHERE fromAddress = ?
					AND accountType = ?`

	return db.getAddresses(ctx, query, excludeAddresses, address, deposit)
}

func (db *Database) getAddresses(ctx context.Context, query string, excludeAddresses []string, args ...interface{}) ([]string, error) {
	rows, err := db.connection.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("can't get exchange transfers: %w", err)
	}

	defer rows.Close()

	var senders []string

	for rows.Next() {
		var sender string

		err := rows.Scan(
			&sender,
		)
		if err != nil {
			return nil, err
		}

		if isExcluded(sender, excludeAddresses) {
			continue
		}

		senders = append(senders, sender)
	}

	return senders, nil
}

func isExcluded(sender string, excludeAddresses []string) bool {
	for _, excludeAddr := range excludeAddresses {
		if sender == excludeAddr {
			return true
		}
	}

	return false
}

func (db *Database) UpdateClusterByAddress(ctx context.Context, cluster uint64, addresses ...string) error {
	query := `UPDATE accounts SET cluster = ? WHERE address IN ( ? )`

	queryIn, args, err := sqlx.In(query, cluster, addresses)
	if err != nil {
		return fmt.Errorf("can't create IN QUERY for cluster updating: %w", err)
	}

	_, err = db.connection.ExecContext(ctx, queryIn, args...)
	if err != nil {
		return fmt.Errorf("can't update cluster by address: %w", err)
	}

	return nil
}

func (db *Database) UpdateClusterByCluster(ctx context.Context, src, dst uint64) error {
	query := `UPDATE accounts SET cluster = ? WHERE cluster = ?`

	_, err := db.connection.ExecContext(ctx, query, dst, src)
	if err != nil {
		return fmt.Errorf("can't update cluster by cluster: %w", err)
	}

	return nil
}

func (db *Database) CreateCluster(ctx context.Context) (int64, error) {
	query := `INSERT INTO clusters() VALUES ()`

	res, err := db.connection.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("can't create cluster: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (db *Database) DeleteCluster(ctx context.Context, id uint64) error {
	query := `DELETE FROM clusters WHERE id = ?`

	_, err := db.connection.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("can't delete cluster: %w", err)
	}

	return nil
}

func (db *Database) MergeClusters(ctx context.Context, src, dst uint64) error {
	err := db.UpdateClusterByCluster(ctx, src, dst)
	if err != nil {
		return err
	}

	return db.DeleteCluster(ctx, src)
}
