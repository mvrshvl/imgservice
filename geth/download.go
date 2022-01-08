package geth

import (
	"context"
	"github.com/gocarina/gocsv"
	"nir/config"
	"nir/database"
	"nir/di"
	"os"
	"os/exec"
	"path"
	"strconv"
)

const dataDirectory = "./geth/data"

func DownloadData(ctx context.Context, fromBlock, toBlock uint64) (*database.NewBlocks, error) {
	err := di.FromContext(ctx).Invoke(func(cfg *config.Config) error {
		cmd := exec.Command("bash", "./geth/download.sh", strconv.FormatUint(fromBlock, 10), strconv.FormatUint(toBlock, 10), cfg.Ethereum.Address)
		return cmd.Run()
	})
	if err != nil {
		return nil, err
	}

	return parseNewBlocks(ctx)
}

func parseNewBlocks(ctx context.Context) (*database.NewBlocks, error) {
	var (
		blocks         []*database.Block
		txs            []*database.Transaction
		exchanges      []*database.Exchange
		tokenTransfers database.TokenTransfers
		logs           database.Logs
	)

	err := di.FromContext(ctx).Invoke(func(c *config.Config) error {
		err := parseCSV(path.Join(dataDirectory, "blocks.csv"), &blocks)
		if err != nil {
			return err
		}

		err = parseCSV(path.Join(dataDirectory, "transactions.csv"), &txs)
		if err != nil {
			return err
		}

		err = parseCSV(path.Join(dataDirectory, "exchanges.csv"), &exchanges)
		if err != nil {
			return err
		}

		err = parseCSV(path.Join(dataDirectory, "logs.csv"), &logs)
		if err != nil {
			return err
		}

		return parseCSV(path.Join(dataDirectory, "token_transfers.csv"), &tokenTransfers)
	})
	if err != nil {
		return nil, err
	}

	return database.GetNewBlocks(txs, blocks, exchanges, tokenTransfers, logs)
}

func parseCSV(filename string, out interface{}) error {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}

	return gocsv.UnmarshalCSV(gocsv.DefaultCSVReader(f), out)
}
