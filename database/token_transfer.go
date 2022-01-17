package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type TokenTransfer struct {
	ContractAddress string `csv:"token_address"`
	SourceAddress   string `csv:"from_address"`
	TargetAddress   string `csv:"to_address"`
	Value           int64  `csv:"value"`
	TxHash          string `csv:"transaction_hash"`
	LogIndex        uint64 `csv:"log_index"`
	BlockNumber     uint64 `csv:"block_number"`
}

type TokenTransfers []*TokenTransfer

func (db *Database) GetAirdrops(ctx context.Context, fromBlock uint64, toBlock uint64, minTransfers uint64) ([]*Airdrop, error) {
	queryAirdrop := `SELECT hash, nonce, blockNumber, transactionIndex, FromAddress, toAddress, value, gas, gasPrice, input, ContractAddress, type FROM transactions
    INNER JOIN (SELECT FromAddress as f, value as v, ContractAddress as c, COUNT(*) as count FROM transactions
                WHERE type = 'transfer'
                  AND blockNumber BETWEEN ? AND ?
                GROUP BY FromAddress, value) g
    ON transactions.FromAddress = g.f
        AND transactions.ContractAddress = g.c
        AND transactions.value = g.v
        WHERE g.count > ?
    	AND transactions.type = 'transfer'
        AND transactions.blockNumber BETWEEN ? AND ?
        AND NOT toAddress = ''`

	rows, err := db.connection.QueryContext(ctx, queryAirdrop, fromBlock, toBlock, minTransfers, fromBlock, toBlock)
	if err != nil {
		return nil, fmt.Errorf("can't get airdrops: %w", err)
	}

	txs, err := scanTransactions(rows)
	if err != nil {
		return nil, err
	}

	return db.FilterOwners(ctx, groupByAirdrop(txs))
}

type Airdrop struct {
	FromAddress     string
	ContractAddress string
	Txs             Transactions
}

func (db *Database) FindTransfersBetweenMembers(ctx context.Context, a *Airdrop) (Transactions, error) {
	receivers := make([]string, 0, len(a.Txs))
	for _, tx := range a.Txs {
		receivers = append(receivers, tx.ToAddress)
	}

	query := `SELECT hash, nonce, blockNumber, transactionIndex, fromAddress, toAddress, value, gas, gasPrice, input, contractAddress, type FROM transactions
				WHERE fromAddress IN ( ? )
				AND toAddress IN ( ? )
				AND contractAddress = ?`

	queryIn, args, err := sqlx.In(query, receivers, receivers, a.ContractAddress)
	if err != nil {
		return nil, fmt.Errorf("can't create IN QUERY for transfers between members: %w", err)
	}

	rows, err := db.connection.QueryContext(ctx, queryIn, args...)
	if err != nil {
		return nil, fmt.Errorf("can't find transfers between members: %w", err)
	}

	return scanTransactions(rows)
}

// группировка транзакций по каждому событию
func groupByAirdrop(txs Transactions) (airdrops []*Airdrop) {
	mappingAirdrop := make(map[string]map[string]Transactions)

	for _, tx := range txs {
		if _, ok := mappingAirdrop[tx.FromAddress]; !ok {
			mappingAirdrop[tx.FromAddress] = map[string]Transactions{*tx.ContractAddress: {tx}}

			continue
		}

		mappingAirdrop[tx.FromAddress][*tx.ContractAddress] = append(mappingAirdrop[tx.FromAddress][*tx.ContractAddress], tx)
	}

	for fromAddress, contractTxs := range mappingAirdrop {
		for contractAddress, txs := range contractTxs {
			airdrops = append(airdrops, &Airdrop{
				FromAddress:     fromAddress,
				ContractAddress: contractAddress,
				Txs:             txs,
			})
		}
	}

	return
}

// FilterOwners вернуть если addr это владелец контракта или для адреса существует approve на этот контракт
func (db *Database) FilterOwners(ctx context.Context, airdrops []*Airdrop) (filtered []*Airdrop, err error) {
	query := `SELECT hash, nonce, blockNumber, transactionIndex, fromAddress, toAddress, value, gas, gasPrice, input, contractAddress, type FROM transactions
				WHERE ContractAddress = ?
				AND ((type = 'transfer'
            	AND toAddress = ''
            	AND FromAddress = ?)
            	OR (type = 'approve'
            	AND toAddress = ?))
				LIMIT 1`

	for _, airdrop := range airdrops {
		tx := new(Transaction)

		err = db.connection.GetContext(ctx, tx, query, airdrop.ContractAddress, airdrop.FromAddress, airdrop.FromAddress)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			return nil, fmt.Errorf("can't filter contract owners %w", err)
		}

		filtered = append(filtered, airdrop)
	}

	return filtered, nil
}
