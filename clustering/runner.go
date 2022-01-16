package clustering

import (
	"context"
	"fmt"
	"nir/clustering/airdrop"
	"nir/clustering/depositreuse"
	"nir/database"
	"sort"
)

func Run(ctx context.Context, subscriber <-chan *database.NewBlocks, errChan chan error) {
	for {
		select {
		case <-ctx.Done():
			return
		case newBlocks := <-subscriber:
			err := clustering(ctx, newBlocks)
			if err != nil {
				errChan <- err
			}
		}
	}
}

func clustering(ctx context.Context, newBlocks *database.NewBlocks) error {
	err := depositreuse.Run(ctx, newBlocks.Transactions)
	if err != nil {
		return fmt.Errorf("can't clustering deposit reuse: %w", err)
	}

	sort.Slice(newBlocks.Blocks, func(i, j int) bool {
		return newBlocks.Blocks[i].Number > newBlocks.Blocks[j].Number
	})

	return airdrop.Run(ctx, newBlocks.Blocks[0].Number)
}
