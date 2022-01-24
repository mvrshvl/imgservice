package database

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
)

type Exchange struct {
	Address string `csv:"address" db:"address"`
	Name    string `csv:"name" db:"name"`
}

type Exchanges []*Exchange

func (exchanges Exchanges) MapAddresses() map[string]struct{} {
	exchs := make(map[string]struct{})
	for _, exch := range exchanges {
		exchs[exch.Address] = struct{}{}
	}

	return exchs
}

func (ex *Exchange) AddExchange(ctx context.Context, db sqlx.ExecerContext) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO exchanges(Address, name)
    			VALUES(?,?)`,
		ex.Address, ex.Name)

	if err != nil {
		return fmt.Errorf("can't add exchange: %w", err)
	}

	return nil
}

func (exs Exchanges) AddExchanges(ctx context.Context, db sqlx.ExecerContext) error {
	for _, ex := range exs {
		err := ex.AddExchange(ctx, db)
		if err != nil && !strings.Contains(err.Error(), "Duplicate entry") {
			return err
		}
	}

	return nil
}
