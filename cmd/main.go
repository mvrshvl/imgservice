package main

import (
	"context"
	"fmt"
	"log"
	"nir/clustering"
	"nir/config"
	"nir/database"
	"nir/di"
	"nir/geth"
	logging "nir/log"
	"nir/server"
	"path"
	"time"
)

const (
	batchBlocksSize = 500
	testRun         = true
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	container, err := di.BuildContainer(
		config.New,
		logging.New,
		database.New,
		geth.New,
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx = di.WithContext(ctx, container)

	srv := server.New("localhost:8080")

	errNotify := make(chan error, 1)

	go func() {
		errNotify <- srv.Run(ctx)
	}()

	subscriber, err := loadData(ctx, errNotify)
	if err != nil {
		log.Fatal(err)
	}

	go clustering.Run(ctx, subscriber, errNotify)
	logging.Info(ctx, <-errNotify)
}

func loadData(ctx context.Context, errNotify chan error) (chan *database.NewBlocks, error) {
	notifyBlock := make(chan *database.NewBlocks, 1000)

	err := di.FromContext(ctx).Invoke(func(db *database.Database) error {
		innerErr := db.Connect(ctx)
		if innerErr != nil {
			return innerErr
		}

		innerErr = addExchanges(ctx, db)
		if innerErr != nil {
			return innerErr
		}

		innerErr = addBlackList(ctx, db)
		if innerErr != nil {
			return innerErr
		}

		dbBlockNum, innerErr := db.GetLastBlock(ctx)
		if innerErr != nil {
			return innerErr
		}

		go collectData(ctx, dbBlockNum, notifyBlock, errNotify)

		return nil
	})

	return notifyBlock, err
}

func addExchanges(ctx context.Context, db *database.Database) error {
	var exchanges database.Exchanges

	err := geth.ParseCSV(path.Join(geth.StaticDirectory, "exchanges.csv"), &exchanges)
	if err != nil {
		return fmt.Errorf("can't parse exchanges: %w", err)
	}

	return exchanges.AddExchanges(ctx, db.GetConnection())
}

type Blacklist struct {
	Address string
	Comment string
	Date    string
}

func addBlackList(ctx context.Context, db *database.Database) error {
	var blacklist []Blacklist

	err := geth.ParseJSON(path.Join(geth.StaticDirectory, "blacklist.json"), &blacklist)
	if err != nil {
		return fmt.Errorf("can't parse blacklist: %w", err)
	}

	for _, blacklistAccount := range blacklist {
		account := &database.Account{
			Address: blacklistAccount.Address,
			AccType: database.ScammerAccount,
		}

		if len(blacklistAccount.Comment) > 0 {
			account.Comment = &blacklistAccount.Comment
		}

		err = account.AddAccount(ctx, db.GetConnection())
		if err != nil {
			return fmt.Errorf("can't add account from blacklist: %w", err)
		}
	}

	return nil
}

func collectData(ctx context.Context, fromBlock uint64, notifyChan chan *database.NewBlocks, errNotify chan error) {
	var (
		ethLastBlock uint64
		blocks       *database.NewBlocks
		err          error
	)

	defer func() {
		if err != nil {
			errNotify <- err
		}
	}()

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		err = di.FromContext(ctx).Invoke(func(w *geth.Worker) (innerErr error) {
			ethLastBlock, innerErr = w.GetLastBlock(ctx)
			return
		})
		if err != nil {
			return
		}

		logging.Infof(ctx, "Current block %d, saved block %d", ethLastBlock, fromBlock)

		if ethLastBlock-fromBlock < batchBlocksSize {
			continue
		}

		blocks, err = geth.DownloadData(ctx, fromBlock+1, fromBlock+batchBlocksSize)
		if err != nil {
			return
		}

		logging.Infof(ctx, "Imported blocks = %d, txs = %d, transfers = %d, approves = %d", len(blocks.Blocks), len(blocks.Transactions), len(blocks.TokenTransfers), len(blocks.Approves))
		err = blocks.Save(ctx)
		if err != nil {
			return
		}

		select {
		case notifyChan <- blocks:
		default:
			logging.Warn(ctx, "can't send newBlocks to clustering")

			continue
		}

		fromBlock += batchBlocksSize
	}
}

//
//func RenderGraph(owners []string, exchanges blockchain.Exchanges, filepath string, clusters clustering.Clusters, showSingleAccounts bool) error {
//	exchangesNodes := make(map[string]opts.GraphNode)
//	for _, exch := range exchanges {
//		exchangesNodes[exch.Address] = opts.GraphNode{Name: exch.Name}
//	}
//
//	ownersNodes := make(map[string]opts.GraphNode)
//	for _, owner := range owners {
//		ownersNodes[owner] = opts.GraphNode{Name: owner}
//	}
//
//	page := components.NewPage()
//	page.AddCharts(
//		clusters.GenerateGraph(exchangesNodes, ownersNodes, showSingleAccounts),
//		//clusters.GenerateLegend(),
//	)
//
//	f, err := os.Create(filepath)
//	if err != nil {
//		return err
//	}
//
//	return page.Render(io.MultiWriter(f))
//}
//
//func SaveClusters(name string, clusters clustering.Clusters) error {
//	bytes, err := json.Marshal(clusters)
//	if err != nil {
//		return err
//	}
//
//	return os.WriteFile(name, bytes, os.ModePerm)
//}
