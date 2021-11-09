package main

import (
	"context"
	"fmt"
	"log"
	"nir/clustering/transfer"
	"nir/config"
	"nir/di"
	logging "nir/log"
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

	var transfers []*transfer.ExchangeTransfer

	err = di.FromContext(ctx).Invoke(func(c *config.Config) {
		transfers = transfers.GetExchangeTransfers(chain, c.Clustering.MaxBlockDiff)
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(transfers)
}
