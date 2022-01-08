package blockchain

type Block struct {
	Number           uint64 `csv:"number"`
	Hash             string `csv:"hash"`
	ParentHash       string `csv:"parent_hash"`
	Nonce            uint64 `csv:"nonce"`
	Miner            string `csv:"miner"`
	GasLimit         uint64 `csv:"gas_limit"`
	GasUsed          uint64 `csv:"gas_used"`
	Timestamp        string `csv:"timestamp"`
	TransactionCount uint64 `csv:"transaction_count"`
}

type Blocks []*Block
