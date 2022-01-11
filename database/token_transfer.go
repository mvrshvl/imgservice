package database

import (
	"context"
	"fmt"
	"math/big"
)

type TokenTransfer struct {
	ContractAddress string  `csv:"token_address"`
	SourceAddress   string  `csv:"from_address"`
	TargetAddress   string  `csv:"to_address"`
	Value           big.Int `csv:"value"`
	TxHash          string  `csv:"transaction_hash"`
	LogIndex        uint64  `csv:"log_index"`
	BlockNumber     uint64  `csv:"block_number"`
}

type TokenTransfers []*TokenTransfer

func (db *Database) GetAirdrop(ctx context.Context, fromBlock uint64, toBlock uint64, minTransfers uint64) (Transactions, error) {
	queryAirdrop := `SELECT hash, COUNT(*) as countFromAddr FROM transactions
						WHERE type = 'transfer'
						AND block BETWEEN ? AND ?
						AND countFromAddr > 5
						GROUP BY fromAddress, value`

	queryTxs := fmt.Sprintf("SELECT * FROM transactions RIGHT JOIN ( %s ) as airdrop ON transactions.hash = airdrop.hash", queryAirdrop)

	rows, err := db.connection.QueryContext(ctx, queryTxs, fromBlock, toBlock, minTransfers)
	if err != nil {
		return nil, err
	}

	return scanTransactions(rows)
}
