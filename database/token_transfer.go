package database

import (
	"context"
	"database/sql"
	"errors"
)

const address0 = "0x0000000000000000000000000000000000000000"

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
	queryAirdrop := `SELECT hash, nonce, blockNumber, transactionIndex, fromAddress, toAddress, value, gas, gasPrice, input, contractAddress, type FROM transactions
    INNER JOIN (SELECT fromAddress as f, value as v, contractAddress as c, COUNT(*) as count FROM transactions
                WHERE type = 'transfer'
                  AND blockNumber BETWEEN ? AND ?
                GROUP BY fromAddress, value) g
    ON transactions.fromAddress = g.f
        AND transactions.contractAddress = g.c
        AND transactions.value = g.v
        WHERE g.count > ?
    	AND transactions.type = 'transfer'
        AND transactions.blockNumber BETWEEN ? AND ?
        AND NOT toAddress = ''`

	rows, err := db.connection.QueryContext(ctx, queryAirdrop, fromBlock, toBlock, minTransfers, fromBlock, toBlock)
	if err != nil {
		return nil, err
	}

	txs, err := scanTransactions(rows)
	if err != nil {
		return nil, err
	}

	return db.FilterOwners(ctx, groupByAirdrop(txs))
}

type Airdrop struct {
	fromAddress     string
	contractAddress string
	txs             Transactions
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
				fromAddress:     fromAddress,
				contractAddress: contractAddress,
				txs:             txs,
			})
		}
	}

	return
}

// FilterOwners вернуть если addr это владелец контракта или для адреса существует approve на этот контракт
func (db *Database) FilterOwners(ctx context.Context, airdrops []*Airdrop) (filtered []*Airdrop, err error) {
	query := `SELECT * FROM transactions
				WHERE contractAddress = ?
				AND ((type = 'transfer'
            	AND toAddress = ''
            	AND fromAddress = ?)
            	OR (type = 'approve'
            	AND toAddress = ?))
				LIMIT 1`

	for _, airdrop := range airdrops {
		var tx Transaction

		err = db.connection.GetContext(ctx, &tx, query, airdrop.contractAddress, airdrop.fromAddress, airdrop.fromAddress)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			return nil, err
		}

		filtered = append(filtered, airdrop)
	}

	return
}
