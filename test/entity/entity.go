package entity

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"nir/test/entity/account"
	"nir/test/writer"
)

type Entity interface {
	GetAccounts() []*common.Address
}

func AddEtherToEntity(ctx context.Context, entity Entity, amount int64) error {
	wr, err := writer.FromContext(ctx)
	if err != nil {
		return err
	}

	for _, acc := range entity.GetAccounts() {
		err := account.AddEtherToAccount(ctx, wr, acc, amount)
		if err != nil {
			return err
		}
	}

	return nil
}
