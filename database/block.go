package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Block struct {
	Number           uint64 `csv:"number" db:"number"`
	Hash             string `csv:"hash" db:"hash"`
	ParentHash       string `csv:"parent_hash" db:"parentHash"`
	Miner            string `csv:"miner" db:"miner"`
	GasLimit         uint64 `csv:"gas_limit" db:"gasLimit"`
	GasUsed          uint64 `csv:"gas_used" db:"gasUsed"`
	Timestamp        int64  `csv:"timestamp" db:"blockTimestamp"`
	TransactionCount uint64 `csv:"transaction_count" db:"transactionsCount"`
}

type Blocks []*Block

func (db *Database) GetLastBlock(ctx context.Context) (uint64, error) {
	blockNum := new(uint64)
	err := db.connection.GetContext(ctx, blockNum, "SELECT number FROM blocks ORDER BY number DESC LIMIT 1")
	if err != nil {
		return 0, fmt.Errorf("can't get last block: %w", err)
	}

	return *blockNum, nil
}

func (block *Block) AddBlock(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO blocks(number, hash, parentHash, miner, gasLimit, gasUsed, blockTimestamp, transactionsCount)
    			VALUES(?,?,?,?,?,?,?,?)`,
		block.Number, block.Hash, block.ParentHash, block.Miner, block.GasLimit, block.GasUsed, time.Unix(block.Timestamp, 0), block.TransactionCount)
	if err != nil {
		return fmt.Errorf("can't add block: %w", err)
	}

	account := &Account{
		Address: block.Miner,
		AccType: miner,
	}

	return account.AddAccount(ctx, tx)
}

func (blocks Blocks) AddBlocks(ctx context.Context, tx *sql.Tx) error {
	for _, b := range blocks {
		err := b.AddBlock(ctx, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) ExecuteTx(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) error) error {
	tx, err := db.connection.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("can't begin tx %w", err)
	}

	err = fn(ctx, tx)
	if err != nil {
		return fmt.Errorf("can't execute tx: %w. rollback: %v", err, tx.Rollback())
	}

	return tx.Commit()
}
