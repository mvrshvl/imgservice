package blockchain

type Block struct {
	Number           uint64 `csv:"number"`
	Hash             string `csv:"hash"`
	ParentHash       string `csv:"parent_hash"`
	Nonce            uint64 `csv:"nonce"`
	SHA3Uncles       string `csv:"sha3_uncles"`
	LogsBloom        string `csv:"logs_bloom"`
	TransactionsRoot string `csv:"transactions_root"`
	StateRoot        string `csv:"state_root"`
	ReceiptsRoot     string `csv:"receipts_root"`
	Miner            string `csv:"miner"`
	Difficulty       uint64 `csv:"difficulty"`
	TotalDifficulty  uint64 `csv:"total_difficulty"`
	Size             uint64 `csv:"size"`
	ExtraData        string `csv:"extra_data"`
	GasLimit         uint64 `csv:"gas_limit"`
	GasUsed          uint64 `csv:"gas_used"`
	Timestamp        string `csv:"timestamp"`
	TransactionCount uint64 `csv:"transaction_count"`
	BaseFeePerGas    uint64 `csv:"base_fee_per_gas"`
}

type Blocks []*Block
