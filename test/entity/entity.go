package entity

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"log"
	"nir/test/entity/account"
	"nir/test/startbalance"
	"sync"
)

type Entity interface {
	GetAccounts() []*common.Address
}

func AddEtherToEntities(ctx context.Context, entities []Entity, amount int64) {
	var (
		wg sync.WaitGroup
	)

	for _, ent := range entities {
		ent := ent

		wg.Add(1)

		go func() {
			defer wg.Done()

			err := AddEtherToEntity(ctx, ent, amount)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	wg.Wait()
}

func AddEtherToEntity(ctx context.Context, entity Entity, amount int64) error {
	var wg sync.WaitGroup

	for _, acc := range entity.GetAccounts() {
		acc := acc

		wg.Add(1)

		err := startbalance.AddTask(ctx, func(a *account.Account) {
			defer wg.Done()
			_, err := a.SendTransaction(ctx, acc, amount, true)
			if err != nil {
				log.Fatal(err)
			}
		})
		if err != nil {
			return err
		}
	}

	wg.Wait()

	return nil
}
