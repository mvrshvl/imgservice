package database

import "math/big"

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
