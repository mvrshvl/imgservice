package database

import "context"

type Block struct {
	Number           uint64 `csv:"number" db:"number"`
	Hash             string `csv:"hash" db:"hash"`
	ParentHash       string `csv:"parent_hash" db:"parentHash"`
	Nonce            uint64 `csv:"nonce" db:"nonce"`
	Miner            string `csv:"miner" db:"miner"`
	GasLimit         uint64 `csv:"gas_limit" db:"gasLimit"`
	GasUsed          uint64 `csv:"gas_used" db:"gasUsed"`
	Timestamp        string `csv:"timestamp" db:"blockTimestamp"`
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
		block.Number, block.Hash, block.ParentHash, block.Nonce, block.Miner, block.GasLimit, block.GasUsed, block.Timestamp, block.TransactionCount)

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
