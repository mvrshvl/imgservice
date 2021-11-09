package blockchain

type Blockchain struct {
	Transactions Transactions
	Blocks       Blocks
	Exchanges    Exchanges
}

func New(transactions []*Transaction, blocks []*Block, exchanges []*Exchange) *Blockchain {
	return &Blockchain{
		Transactions: transactions,
		Blocks:       blocks,
		Exchanges:    exchanges,
	}
}
