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
	"path"
	"time"
)

const (
	batchBlocksSize = 10
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

	errNotify := make(chan error, 1)
	subscriber, err := loadData(ctx, errNotify)
	if err != nil {
		log.Fatal(err)
	}

	go clustering.Run(ctx, subscriber, errNotify)
	logging.Info(ctx, <-errNotify)
	// остановился на тестировании записи в бд
	//logging.Info(ctx, "Prepare data...")
	//
	//chain, err := prepareBlockchain(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//var (
	//	ts  []*transfer.ExchangeTransfer
	//	cfg *config.Config
	//)
	//
	//err = di.FromContext(ctx).Invoke(func(c *config.Config) {
	//	cfg = c
	//})
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//var (
	//	wg                                                 sync.WaitGroup
	//	depositClusters, airdropClusters, selfauthClusters clustering.Clusters
	//)
	//
	//logging.Info(ctx, "Start clustering.")
	//
	//wg.Add(3)
	//go func() {
	//	defer wg.Done()
	//
	//	ts = transfer.GetTxsToExchange(chain, cfg.Clustering.MaxBlockDiff)
	//	depositClusters = depositreuse.Find(ts)
	//
	//	err = SaveClusters("deposit.json", depositClusters)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//}()
	//
	//go func() {
	//	defer wg.Done()
	//
	//	airdropClusters, err = airdrop.Find(chain.TokenTransfers)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	err = SaveClusters("airdrop.json", airdropClusters)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//}()
	//
	//go func() {
	//	defer wg.Done()
	//
	//	selfauthClusters = selfauth.Find(chain.Approves)
	//
	//	err = SaveClusters("self-auth.json", selfauthClusters)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//}()
	//
	//wg.Wait()
	//
	//logging.Info(ctx, "Start merging.")
	//
	//// todo optimize this
	//m := airdropClusters.Merge(depositClusters)
	//
	//err = SaveClusters("airdrop_deposit.json", m)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//merged := m.Merge(selfauthClusters)
	//
	//err = SaveClusters("all.json", merged)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//logging.Info(ctx, "Start rendering.")
	//
	//err = RenderGraph(airdrop.GetAirdropDistributors(chain.TokenTransfers), chain.Exchanges, cfg.Output.GraphDepositsReuse, merged, cfg.ShowSingleAccount)
	//if err != nil {
	//	log.Fatal(err)
	//}
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
	if !testRun {
		return nil
	}

	var exchanges database.Exchanges

	err := geth.ParseCSV(path.Join(geth.DataDirectory, "exchanges.csv"), &exchanges)
	if err != nil {
		return fmt.Errorf("can't parse exchanges: %w", err)
	}

	return db.AddExchanges(ctx, exchanges)
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
