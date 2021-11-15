package blockchain

type Blockchain struct {
	Transactions   Transactions
	Blocks         Blocks
	Exchanges      Exchanges
	TokenTransfers TokenTransfers
}

func New(transactions []*Transaction, blocks []*Block, exchanges []*Exchange, tokenTransfers TokenTransfers) *Blockchain {
	return &Blockchain{
		Transactions:   transactions,
		Blocks:         blocks,
		Exchanges:      exchanges,
		TokenTransfers: tokenTransfers,
	}
}
