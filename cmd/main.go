package main

import (
	"context"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"io"
	"log"
	"nir/clustering/blockchain"
	"nir/clustering/depositreuse"
	"nir/clustering/transfer"
	"nir/config"
	"nir/di"
	logging "nir/log"
	"os"
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

	ts = transfer.GetExchangeTransfers(chain, cfg.Clustering.MaxBlockDiff)

	clusters := depositreuse.Find(ts)

	err = RenderGraph(chain.Exchanges, cfg.Output.GraphDepositsReuse, clusters, cfg.ShowSingleAccount)
	if err != nil {
		log.Fatal(err)
	}
}

func RenderGraph(exchanges blockchain.Exchanges, filepath string, clusters depositreuse.Clusters, showSingleAccounts bool) error {
	exchangesNodes := make(map[string]opts.GraphNode)
	for _, exch := range exchanges {
		exchangesNodes[exch.Address] = opts.GraphNode{Name: exch.Name}
	}

	page := components.NewPage()
	page.AddCharts(
		clusters.GenerateGraph(exchangesNodes, showSingleAccounts),
	)

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	return page.Render(io.MultiWriter(f))
}
