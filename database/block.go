package database

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
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
		return 0, err
	}

	return *blockNum, nil
}

func (db *Database) AddBlock(ctx context.Context, block *Block) error {
	_, err := db.connection.ExecContext(ctx,
		`INSERT INTO blocks(number, hash, parentHash, nonce, miner, gasLimit, gasUsed, blockTimestamp, transactionsCount)
    			VALUES(?,?,?,?,?,?,?,?,?)`,
		block.Number, common.HexToHash(block.Hash).Bytes(), common.HexToHash(block.ParentHash), block.Nonce, common.HexToAddress(block.Miner).Bytes(), block.GasLimit, block.GasUsed, time.Unix(block.Timestamp, 0), block.TransactionCount)

	return err
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

//func cutAddress(hash string, length int) string {
//	if len(hash) == length+2 {
//		return hash[2:]
//	}
//
//	return hash
//}
