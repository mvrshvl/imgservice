package database

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
)

type Transaction struct {
	Hash             string  `csv:"hash" db:"hash"`
	Nonce            uint64  `csv:"nonce" db:"nonce"`
	BlockNumber      uint64  `csv:"block_number" db:"blockNumber"`
	TransactionIndex uint64  `csv:"transaction_index" db:"transactionIndex"`
	FromAddress      string  `csv:"from_address" db:"fromAddress"`
	ToAddress        string  `csv:"to_address" db:"toAddress"`
	Value            float64 `csv:"value" db:"value"`
	Gas              uint64  `csv:"gas" db:"gas"`
	GasPrice         uint64  `csv:"gas_price" db:"gasPrice"`
}

type Transactions []*Transaction

func (db *Database) AddTransaction(ctx context.Context, tx *Transaction) error {
	_, err := db.connection.ExecContext(ctx,
		`INSERT INTO transactions(hash, nonce, blockNumber, transactionIndex, fromAddress, toAddress, value, gas, gasPrice)
    			VALUES(?,?,?,?,?,?,?,?,?)`,
		common.HexToHash(tx.Hash).Bytes(), tx.Nonce, tx.BlockNumber, tx.TransactionIndex, tx.FromAddress, tx.ToAddress, tx.Value, tx.Gas, tx.GasPrice)

	return err
}

func (db *Database) AddTransactions(ctx context.Context, txs Transactions) error {
	for _, tx := range txs {
		err := db.AddTransaction(ctx, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

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
