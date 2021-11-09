package blockchain

type Transaction struct {
	Hash                 string  `csv:"hash"`
	Nonce                uint64  `csv:"nonce"`
	BlockHash            string  `csv:"block_hash"`
	BlockNumber          uint64  `csv:"block_number"`
	TransactionIndex     uint64  `csv:"transaction_index"`
	FromAddress          string  `csv:"from_address"`
	ToAddress            string  `csv:"to_address"`
	Value                float64 `csv:"value"`
	Gas                  uint64  `csv:"gas"`
	GasPrice             uint64  `csv:"gas_price"`
	Input                string  `csv:"input"`
	BlockTimestamp       string  `csv:"block_timestamp"`
	MaxFee               uint64  `csv:"max_fee"`
	MaxFeePerGas         uint64  `csv:"max_fee_per_gas"`
	MaxPriorityFeePerGas uint64  `csv:"max_priority_fee_per_gas"`
	TransactionType      string  `csv:"transaction_type"`
}

type Transactions []*Transaction

func (transactions Transactions) GetFromAddresses() (addresses []string) {
	for _, tx := range transactions {
		addresses = append(addresses, tx.FromAddress)
	}

	return
}

func (transactions Transactions) GetTransactionsToAddresses(addresses map[string]struct{}) (txs Transactions) {
	for _, tx := range transactions {
		if _, ok := addresses[tx.ToAddress]; ok {
			txs = append(txs, tx)
		}
	}

	return
}

func (transactions Transactions) MapFromAddresses() map[string]struct{} {
	addresses := make(map[string]struct{})
	for _, exch := range transactions {
		addresses[exch.FromAddress] = struct{}{}
	}

	return addresses
}
