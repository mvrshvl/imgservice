package main

import (
	"context"
	"github.com/gocarina/gocsv"
	"nir/clustering/blockchain"
	"nir/config"
	"nir/di"
	"os"
)

func prepareBlockchain(ctx context.Context) (*blockchain.Blockchain, error) {
	var (
		blocks         []*blockchain.Block
		txs            []*blockchain.Transaction
		exchanges      []*blockchain.Exchange
		tokenTransfers blockchain.TokenTransfers
		logs           blockchain.Logs
	)

	err := di.FromContext(ctx).Invoke(func(c *config.Config) error {
		err := parseCSV(c.BlocksTable, &blocks)
		if err != nil {
			return err
		}

		err = parseCSV(c.TransactionsTable, &txs)
		if err != nil {
			return err
		}

		err = parseCSV(c.ExchangesTable, &exchanges)
		if err != nil {
			return err
		}

		err = parseCSV(c.Logs, &logs)
		if err != nil {
			return err
		}

		return parseCSV(c.TokenTransfersTable, &tokenTransfers)
	})
	if err != nil {
		return nil, err
	}

	return blockchain.New(txs, blocks, exchanges, tokenTransfers, logs)
}

func parseCSV(filename string, out interface{}) error {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}

	return gocsv.UnmarshalCSV(gocsv.DefaultCSVReader(f), out)
}
