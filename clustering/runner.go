package clustering

import (
	"context"
	"nir/clustering/depositreuse"
	"nir/database"
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
	return depositreuse.Run(ctx, newBlocks.Transactions)
}
