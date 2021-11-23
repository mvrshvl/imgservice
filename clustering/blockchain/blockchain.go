package blockchain

import "encoding/hex"

type Blockchain struct {
	Transactions   Transactions
	Blocks         Blocks
	Exchanges      Exchanges
	TokenTransfers TokenTransfers
	Approves       ERC20Approves
}

func New(transactions []*Transaction, blocks []*Block, exchanges []*Exchange, tokenTransfers TokenTransfers, logs Logs) (*Blockchain, error) {
	for _, l := range logs {
		data, err := hex.DecodeString(l.Data)
		if err != nil {
			return nil, err
		}

		l.Data = string(data)
	}

	return &Blockchain{
		Transactions:   transactions,
		Blocks:         blocks,
		Exchanges:      exchanges,
		TokenTransfers: tokenTransfers,
		Approves:       logs.ToApproves(),
	}, nil
}
