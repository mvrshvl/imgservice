package main

import (
	"context"
	"encoding/json"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"io"
	"log"
	"nir/clustering"
	"nir/clustering/airdrop"
	"nir/clustering/blockchain"
	"nir/clustering/depositreuse"
	"nir/clustering/selfauth"
	"nir/clustering/transfer"
	"nir/config"
	"nir/di"
	logging "nir/log"
	"os"
	"sync"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	container, err := di.BuildContainer(
		config.New,
		logging.New,
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx = di.WithContext(ctx, container)

	logging.Info(ctx, "Prepare data...")

	chain, err := prepareBlockchain(ctx)
	if err != nil {
		log.Fatal(err)
	}

	var (
		ts  []*transfer.ExchangeTransfer
		cfg *config.Config
	)

	err = di.FromContext(ctx).Invoke(func(c *config.Config) {
		cfg = c
	})
	if err != nil {
		log.Fatal(err)
	}

	var (
		wg                                                 sync.WaitGroup
		depositClusters, airdropClusters, selfauthClusters clustering.Clusters
	)

	logging.Info(ctx, "Start clustering.")

	wg.Add(3)
	go func() {
		defer wg.Done()

		ts = transfer.GetExchangeTransfers(chain, cfg.Clustering.MaxBlockDiff)
		depositClusters = depositreuse.Find(ts)

		err = SaveClusters("deposit.json", depositClusters)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()

		airdropClusters, err = airdrop.Find(chain.TokenTransfers)
		if err != nil {
			log.Fatal(err)
		}

		err = SaveClusters("airdrop.json", airdropClusters)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()

		selfauthClusters = selfauth.Find(chain.Approves)

		err = SaveClusters("self-auth.json", selfauthClusters)
		if err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()

	logging.Info(ctx, "Start merging.")

	// todo optimize this
	m := airdropClusters.Merge(depositClusters)

	err = SaveClusters("airdrop_deposit.json", m)
	if err != nil {
		log.Fatal(err)
	}

	merged := m.Merge(selfauthClusters)

	err = SaveClusters("all.json", merged)
	if err != nil {
		log.Fatal(err)
	}

	logging.Info(ctx, "Start rendering.")

	err = RenderGraph(airdrop.GetAirdropDistributors(chain.TokenTransfers), chain.Exchanges, cfg.Output.GraphDepositsReuse, merged, cfg.ShowSingleAccount)
	if err != nil {
		log.Fatal(err)
	}
}

func RenderGraph(owners []string, exchanges blockchain.Exchanges, filepath string, clusters clustering.Clusters, showSingleAccounts bool) error {
	exchangesNodes := make(map[string]opts.GraphNode)
	for _, exch := range exchanges {
		exchangesNodes[exch.Address] = opts.GraphNode{Name: exch.Name}
	}

	ownersNodes := make(map[string]opts.GraphNode)
	for _, owner := range owners {
		ownersNodes[owner] = opts.GraphNode{Name: owner}
	}

	page := components.NewPage()
	page.AddCharts(
		clusters.GenerateGraph(exchangesNodes, ownersNodes, showSingleAccounts),
		//clusters.GenerateLegend(),
	)

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	return page.Render(io.MultiWriter(f))
}

func SaveClusters(name string, clusters clustering.Clusters) error {
	bytes, err := json.Marshal(clusters)
	if err != nil {
		return err
	}

	return os.WriteFile(name, bytes, os.ModePerm)
}
