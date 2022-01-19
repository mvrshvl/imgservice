package database

import (
	"context"
	"fmt"
)

const (
	fromAddress = `fromAddress`
	toAddress   = `toAddress`

	selfApproveQuery = `
		SELECT transactions.hash, transactions.nonce, transactions.blockNumber,
			   transactions.transactionIndex, transactions.fromAddress, transactions.toAddress,
               transactions.value, transactions.gas, transactions.gasPrice,
			   transactions.input, transactions.contractAddress, transactions.type 
		FROM transactions
		LEFT JOIN ( SELECT %s AS address, contractAddress AS contract, COUNT(*) AS countTxs
					FROM transactions
					WHERE type = 'approve'
					AND blockNumber BETWEEN ? AND ?
					GROUP BY %s, contractAddress ) AS owner
		ON transactions.contractAddress = owner.contract AND
           transactions.%s = owner.address
		LEFT JOIN accounts
		ON 	transactions.%s = accounts.address
		WHERE owner.countTxs < ? 	
		AND transactions.blockNumber BETWEEN ? AND ?
		AND transactions.type = 'approve'
		AND NOT accounts.type IN ('exchange', 'deposit')`
)

func (db *Database) GetSelfApproveTxs(ctx context.Context, fromBlock uint64, toBlock uint64, maxApproves uint64) (Transactions, error) {
	fromApproves, err := db.getSelfApproveAddresses(ctx, fromBlock, toBlock, maxApproves, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("getting from approves failed: %w", err)
	}

	toApproves, err := db.getSelfApproveAddresses(ctx, fromBlock, toBlock, maxApproves, toAddress)
	if err != nil {
		return nil, fmt.Errorf("getting to approves failed: %w", err)
	}

	return append(fromApproves, toApproves...), nil
}

func (db *Database) getSelfApproveAddresses(ctx context.Context, fromBlock uint64, toBlock uint64, maxApproves uint64, addressField string) (Transactions, error) {
	query := fmt.Sprintf(selfApproveQuery, addressField, addressField, addressField, addressField)

	rows, err := db.connection.QueryContext(ctx, query, fromBlock, toBlock, maxApproves, fromBlock, toBlock)
	if err != nil {
		return nil, fmt.Errorf("can't get self approves: %w", err)
	}

	return scanTransactions(rows)
}
