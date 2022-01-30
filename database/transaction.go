package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"math/big"
	logging "nir/log"
)

type TxType string

const (
	address0          = "0x0000000000000000000000000000000000000000"
	TxApprove  TxType = "approve"
	TxTransfer TxType = "transfer"
)

type Transaction struct {
	Hash             string   `csv:"hash" db:"hash"`
	Nonce            uint64   `csv:"nonce" db:"nonce"`
	BlockNumber      uint64   `csv:"block_number" db:"blockNumber"`
	TransactionIndex *uint64  `csv:"transaction_index" db:"transactionIndex"`
	FromAddress      string   `csv:"from_address" db:"fromAddress"`
	ToAddress        string   `csv:"to_address" db:"toAddress"`
	Value            *big.Int `csv:"value" db:"value"`
	Gas              *big.Int `csv:"gas" db:"gas"`
	GasPrice         *big.Int `csv:"gas_price" db:"gasPrice"`
	Input            string   `csv:"input" db:"input"`
	ContractAddress  *string  `db:"contractAddress"`
	Type             *TxType  `db:"type"`
}

type Transactions []*Transaction

func (tx *Transaction) AddTransaction(ctx context.Context, dbTx *sql.Tx) error {
	_, err := dbTx.ExecContext(ctx,
		`INSERT INTO transactions(hash, nonce, blockNumber, transactionIndex, FromAddress, toAddress, value, gas, gasPrice, input)
    			VALUES(?,?,?,?,?,?,?,?,?,?)`,
		tx.Hash, tx.Nonce, tx.BlockNumber, tx.TransactionIndex, tx.FromAddress, tx.ToAddress, tx.Value.Int64(), tx.Gas.Int64(), tx.GasPrice.Int64(), tx.Input)

	if err != nil {
		return fmt.Errorf("can't add tx: %w", err)
	}

	err = Account{
		Address: tx.FromAddress,
		AccType: eoa,
	}.AddAccount(ctx, dbTx)
	if err != nil {
		return err
	}

	return Account{
		Address: tx.FromAddress,
		AccType: eoa,
	}.AddAccount(ctx, dbTx)
}

func (txs Transactions) AddTransactions(ctx context.Context, dbTx *sql.Tx) error {
	for _, tx := range txs {
		err := tx.AddTransaction(ctx, dbTx)
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateTxType(ctx context.Context, dbTx *sql.Tx, hash, toAddress, contractAddr string, value int64, t TxType) error {
	_, err := dbTx.ExecContext(ctx,
		`UPDATE transactions SET toAddress = ?, contractAddress = ?, type = ?, value = ? WHERE hash = ?`,
		toAddress, contractAddr, t, value, hash)

	if err != nil {
		return fmt.Errorf("can't update tx: %w", err)
	}

	return nil
}

func updateTxDeploy(ctx context.Context, tx *sql.Tx, hash, contractAddr string, value int64, t TxType) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE transactions SET contractAddress = ?, type = ?, value = ? WHERE hash = ?`,
		contractAddr, t, value, hash)

	if err != nil {
		return fmt.Errorf("can't update tx: %w", err)
	}

	return nil
}

func (txs TokenTransfers) UpdateTokenTransfers(ctx context.Context, dbTx *sql.Tx) error {
	for _, tx := range txs {
		if tx.SourceAddress == address0 {
			err := updateTxDeploy(ctx, dbTx, tx.TxHash, tx.ContractAddress, tx.Value, TxTransfer)
			if err != nil {
				return err
			}

			continue
		}

		err := UpdateTxType(ctx, dbTx, tx.TxHash, tx.TargetAddress, tx.ContractAddress, tx.Value, TxTransfer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (txs ERC20Approves) UpdateApproves(ctx context.Context, dbTx *sql.Tx) error {
	for _, tx := range txs {
		err := UpdateTxType(ctx, dbTx, tx.TxHash, tx.ToAddress, tx.ContractAddress, 0, TxApprove)
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

	query := `SELECT hash, nonce, blockNumber, transactionIndex, fromAddress, toAddress, value, gas, gasPrice, input, contractAddress, type FROM transactions
    				LEFT JOIN exchanges
    				ON transactions.toAddress = exchanges.Address
					WHERE hash IN ( ? )
					  AND exchanges.Address IS NOT NULL`

	queryIn, args, err := sqlx.In(query, hashes)
	if err != nil {
		return nil, fmt.Errorf("can't create IN QUERY: %w", err)
	}

	rows, err := db.connection.QueryContext(ctx, queryIn, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("can't get ExchangeAccount transfers: %w", err)
	}

	txsToExchange, err := scanTransactions(rows)
	if err != nil {
		return nil, err
	}

	// update accounts set deposit and ExchangeAccount type
	for _, txToExchange := range txsToExchange {
		err = db.UpdateAccountType(ctx, txToExchange.FromAddress, deposit)
		if err != nil {
			return nil, err
		}

		err = db.UpdateAccountType(ctx, txToExchange.ToAddress, ExchangeAccount)
		if err != nil {
			return nil, err
		}
	}

	return txsToExchange, nil
}

func (db *Database) GetExchangeTransfer(ctx context.Context, txsToExchange Transactions, diffBlock uint64, diffGasKoef float64) ([]*ExchangeTransfer, error) {
	transfers := make([]*ExchangeTransfer, 0, len(txsToExchange))

	for _, txToExchange := range txsToExchange {
		query := `SELECT hash, nonce, blockNumber, fromAddress, transactionIndex, toAddress, value, gas, gasPrice, input, contractAddress, type FROM transactions
    				LEFT JOIN exchangeTransfers
    				ON transactions.hash = exchangeTransfers.txDeposit
					LEFT JOIN accounts
					ON transactions.FromAddress = accounts.address
					WHERE transactions.toAddress = ?
					  AND transactions.blockNumber BETWEEN ? AND ?
					  AND transactions.value BETWEEN ? AND ?
					  AND exchangeTransfers.txDeposit IS NULL
					  AND NOT accounts.accountType = 'ExchangeAccount'
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

		if float64(txToExchange.Value.Int64()) > float64(txToExchange.Gas.Int64()*txToExchange.GasPrice.Int64())*diffGasKoef {
			minValue = float64(txToExchange.Value.Int64()) - float64(txToExchange.Gas.Int64()*txToExchange.GasPrice.Int64())*diffGasKoef
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
			logging.Debugf(ctx, "tx to deposit not found (tx to ExchangeAccount %s, range blocks %v-%v, range value %v-%v, to Address %v)", txToExchange.Hash, minBlock, maxBlock, minValue, txToExchange.Value, txToExchange.ToAddress)

			continue
		}

		_, err = db.connection.ExecContext(ctx,
			`INSERT INTO exchangeTransfers(txDeposit, txExchange) VALUES(?,?)`,
			txs[0].Hash, txToExchange.Hash)
		if err != nil {
			return nil, fmt.Errorf("can't add ExchangeAccount transfer: %w", err)
		}

		transfers = append(transfers, &ExchangeTransfer{
			TxExchange: txToExchange.Hash,
			TxDeposit:  txs[0].Hash,
		})
	}

	return transfers, nil
}

func (db *Database) GetTransactionsByAddress(ctx context.Context, address string) (Transactions, error) {
	query := `SELECT hash, nonce, blockNumber, fromAddress, transactionIndex, toAddress,
				value, gas, gasPrice, input, contractAddress, type FROM transactions
					WHERE ( toAddress = ?
					OR fromAddress = ? )
					AND value > 0`

	rows, err := db.connection.QueryContext(ctx, query, address, address)
	if err != nil {
		return nil, fmt.Errorf("can't get txs by address: %w", err)
	}

	return scanTransactions(rows)
}

func scanTransactions(rows *sql.Rows) (txs Transactions, err error) {
	for rows.Next() {
		var tx Transaction

		err = rows.Scan(
			&tx.Hash,
			&tx.Nonce,
			&tx.BlockNumber,
			&tx.TransactionIndex,
			&tx.FromAddress,
			&tx.ToAddress,
			&tx.Value,
			&tx.Gas,
			&tx.GasPrice,
			&tx.Input,
			&tx.ContractAddress,
			&tx.Type,
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
