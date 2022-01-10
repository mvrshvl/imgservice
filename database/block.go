package database

import (
	"context"
	"fmt"
	"time"
)

type Block struct {
	Number           uint64 `csv:"number" db:"number"`
	Hash             string `csv:"hash" db:"hash"`
	ParentHash       string `csv:"parent_hash" db:"parentHash"`
	Nonce            uint64 `csv:"nonce" db:"nonce"`
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

func (db *Database) AddBlock(ctx context.Context, block *Block) error {
	_, err := db.connection.ExecContext(ctx,
		`INSERT INTO blocks(number, hash, parentHash, nonce, miner, gasLimit, gasUsed, blockTimestamp, transactionsCount)
    			VALUES(?,?,?,?,?,?,?,?,?)`,
		block.Number, block.Hash, block.ParentHash, block.Nonce, block.Miner, block.GasLimit, block.GasUsed, time.Unix(block.Timestamp, 0), block.TransactionCount)
	if err != nil {
		return fmt.Errorf("can't add block: %w", err)
	}

	return db.AddAccount(ctx, &Account{
		Address: block.Miner,
		AccType: miner,
		Cluster: 0,
	})
}

func (db *Database) AddBlocks(ctx context.Context, blocks Blocks) error {
	for _, b := range blocks {
		err := db.AddBlock(ctx, b)
		if err != nil {
			return err
		}
	}

	return nil
}
