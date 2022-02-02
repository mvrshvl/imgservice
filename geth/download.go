package geth

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gocarina/gocsv"
	"nir/config"
	"nir/database"
	"nir/di"
	"os"
	"os/exec"
	"path"
	"strconv"
)

const (
	DataDirectory   = "./geth/data"
	StaticDirectory = "./geth/data/static"
)

func DownloadData(ctx context.Context, fromBlock, toBlock uint64) (*database.NewBlocks, error) {
	err := di.FromContext(ctx).Invoke(func(cfg *config.Config) error {
		cmd := exec.Command("bash", "./geth/download.sh", strconv.FormatUint(fromBlock, 10), strconv.FormatUint(toBlock, 10), cfg.Ethereum.Address)

		return cmd.Run()
	})
	if err != nil {
		return nil, err
	}

	return parseNewBlocks()
}

func parseNewBlocks() (*database.NewBlocks, error) {
	var (
		blocks         []*database.Block
		txs            []*database.Transaction
		exchanges      []*database.Exchange
		tokenTransfers database.TokenTransfers
		logs           database.Logs
	)

	err := ParseCSV(path.Join(DataDirectory, "blocks.csv"), &blocks)
	if err != nil {
		return nil, err
	}

	err = ParseCSV(path.Join(DataDirectory, "transactions.csv"), &txs)
	if err != nil {
		return nil, err
	}

	err = ParseCSV(path.Join(DataDirectory, "logs.csv"), &logs)
	if err != nil {
		return nil, err
	}

	err = ParseCSV(path.Join(DataDirectory, "token_transfers.csv"), &tokenTransfers)
	if err != nil {
		return nil, err
	}

	return database.GetNewBlocks(txs, blocks, exchanges, tokenTransfers, logs)
}

func ParseCSV(filename string, out interface{}) error {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}

	err = gocsv.UnmarshalCSV(gocsv.DefaultCSVReader(f), out)
	if err != nil && !errors.Is(err, gocsv.ErrEmptyCSVFile) {
		return err
	}

	return nil
}

func ParseJSON(filename string, out interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, out)
}
