package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	logging "nir/log"
)

type TxType string

const (
	TxApprove  TxType = "approve"
	TxTransfer TxType = "transfer"
)

type Transaction struct {
	Hash             string  `csv:"hash" db:"hash"`
	Nonce            uint64  `csv:"nonce" db:"nonce"`
	BlockNumber      uint64  `csv:"block_number" db:"blocknumber"`
	TransactionIndex uint64  `csv:"transaction_index" db:"transactionindex"`
	FromAddress      string  `csv:"from_address" db:"fromaddress"`
	ToAddress        string  `csv:"to_address" db:"toaddress"`
	Value            float64 `csv:"value" db:"value"`
	Gas              uint64  `csv:"gas" db:"gas"`
	GasPrice         uint64  `csv:"gas_price" db:"gasprice"`
	Input            string  `csv:"input" db:"input"`
	ContractAddress  string  `db:"contractaddress"`
	Type             TxType  `db:"type"`
}

type Transactions []*Transaction

func (db *Database) AddTransaction(ctx context.Context, tx *Transaction) error {
	_, err := db.connection.ExecContext(ctx,
		`INSERT INTO transactions(hash, nonce, blockNumber, transactionIndex, fromAddress, toAddress, value, gas, gasPrice, input)
    			VALUES(?,?,?,?,?,?,?,?,?,?)`,
		tx.Hash, tx.Nonce, tx.BlockNumber, tx.TransactionIndex, tx.FromAddress, tx.ToAddress, tx.Value, tx.Gas, tx.GasPrice, tx.Input)

	if err != nil {
		return fmt.Errorf("can't add tx: %w", err)
	}

	err = db.AddAccount(ctx, &Account{
		address: tx.FromAddress,
		accType: eoa,
		cluster: 0,
	})
	if err != nil {
		return err
	}

	return db.AddAccount(ctx, &Account{
		address: tx.ToAddress,
		accType: eoa,
		cluster: 0,
	})
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

func (db *Database) UpdateTxType(ctx context.Context, hash, contractAddr string, t TxType) error {
	_, err := db.connection.ExecContext(ctx,
		`UPDATE transactions SET contractAddress = ?, type = ? WHERE hash = ?`,
		contractAddr, t, hash)

	if err != nil {
		return fmt.Errorf("can't update tx: %w", err)
	}

	return nil
}

func (db *Database) UpdateTokenTransfers(ctx context.Context, txs TokenTransfers) error {
	for _, tx := range txs {
		err := db.UpdateTxType(ctx, tx.TxHash, tx.ContractAddress, TxTransfer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) UpdateApproves(ctx context.Context, txs ERC20Approves) error {
	for _, tx := range txs {
		err := db.UpdateTxType(ctx, tx.TxHash, tx.ContractAddress, TxApprove)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) GetTxsToExchange(ctx context.Context, txs Transactions) (Transactions, error) {
	hashes := make([]string, len(txs))

	for i, tx := range txs {
		hashes[i] = tx.Hash
	}

	if len(hashes) == 0 {
		return nil, nil
	}

	query := `SELECT hash, nonce, blocknumber, fromaddress, toaddress, value, gas, gasprice, input
					FROM transactions
    				LEFT JOIN exchanges
    				ON transactions.toAddress = exchanges.address
					WHERE hash IN ( ? )
					  AND exchanges.address IS NOT NULL`

	queryIn, args, err := sqlx.In(query, hashes)
	if err != nil {
		return nil, fmt.Errorf("can't create IN QUERY: %w", err)
	}

	rows, err := db.connection.QueryContext(ctx, queryIn, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("can't get exchange transfers: %w", err)
	}

	return scanTransactions(rows)
}

func (db *Database) GetExchangeTransfer(ctx context.Context, txsToExchange Transactions, diffBlock uint64, diffGasKoef float64) ([]*ExchangeTransfer, error) {
	transfers := make([]*ExchangeTransfer, 0, len(txsToExchange))

	for _, txToExchange := range txsToExchange {
		query := `SELECT hash, nonce, blocknumber, fromaddress, toaddress, 
       					value, gas, gasprice, input
					FROM transactions
    				LEFT JOIN exchangeTransfers
    				ON transactions.hash = exchangeTransfers.txDeposit
					WHERE transactions.toAddress = ?
					  AND transactions.blockNumber BETWEEN ? AND ?
					  AND transactions.value BETWEEN ? AND ?
					  AND exchangeTransfers.txDeposit IS NULL
					  LIMIT 1`

		var (
			minBlock = uint64(1)
			maxBlock = uint64(1)
			minValue = float64(1)
		)

		if txToExchange.BlockNumber > diffBlock {
			minBlock = txToExchange.BlockNumber - diffBlock
		}

		if txToExchange.BlockNumber > 1 {
			maxBlock = txToExchange.BlockNumber - 1
		}

		if txToExchange.Value > float64(txToExchange.Gas*txToExchange.GasPrice)*diffGasKoef {
			minValue = txToExchange.Value - float64(txToExchange.Gas*txToExchange.GasPrice)*diffGasKoef
		}
		rows, err := db.connection.QueryContext(ctx, query, txToExchange.FromAddress,
			minBlock, maxBlock,
			minValue, txToExchange.Value)
		if err != nil {
			return nil, fmt.Errorf("can't get tx to deposit: %w", err)
		}

		txs, err := scanTransactions(rows)
		if err != nil {
			return nil, err
		}

		if len(txs) == 0 {
			logging.Debugf(ctx, "tx to deposit not found (tx to exchange %s, range blocks %v-%v, range value %v-%v, to address %v)", txToExchange.Hash, minBlock, maxBlock, minValue, txToExchange.Value, txToExchange.ToAddress)

			continue
		}

		_, err = db.connection.ExecContext(ctx,
			`INSERT INTO exchangeTransfers(txDeposit, txExchange) VALUES(?,?)`,
			txs[0].Hash, txToExchange.Hash)
		if err != nil {
			return nil, fmt.Errorf("can't add exchange transfer: %w", err)
		}

		transfers = append(transfers, &ExchangeTransfer{
			TxExchange: txToExchange.Hash,
			TxDeposit:  txs[0].Hash,
		})
	}

	return transfers, nil
}

func scanTransactions(rows *sql.Rows) (txs Transactions, err error) {
	for rows.Next() {
		var tx Transaction

		err = rows.Scan(
			&tx.Hash,
			&tx.Nonce,
			&tx.BlockNumber,
			&tx.FromAddress,
			&tx.ToAddress,
			&tx.Value,
			&tx.Gas,
			&tx.GasPrice,
			&tx.Input,
		)

		if err != nil {
			return nil, err
		}

		txs = append(txs, &tx)
	}

	return txs, nil
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
