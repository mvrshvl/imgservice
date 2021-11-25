package blockchain

type Blockchain struct {
	Transactions   Transactions
	Blocks         Blocks
	Exchanges      Exchanges
	TokenTransfers TokenTransfers
	Approves       ERC20Approves
}

func New(transactions []*Transaction, blocks []*Block, exchanges []*Exchange, tokenTransfers TokenTransfers, logs Logs) (*Blockchain, error) {
	approves, err := logs.ToApproves(transactions)
	if err != nil {
		return nil, err
	}

	return &Blockchain{
		Transactions:   transactions,
		Blocks:         blocks,
		Exchanges:      exchanges,
		TokenTransfers: tokenTransfers,
		Approves:       approves,
	}, nil
}
