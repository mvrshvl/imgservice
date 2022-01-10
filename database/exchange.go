package database

import (
	"context"
	"fmt"
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

func (db *Database) AddExchange(ctx context.Context, ex *Exchange) error {
	_, err := db.connection.ExecContext(ctx,
		`INSERT INTO exchanges(address, name)
    			VALUES(?,?)`,
		ex.Address, ex.Name)

	if err != nil {
		return fmt.Errorf("can't add exchange: %w", err)
	}

	return nil
}

func (db *Database) AddExchanges(ctx context.Context, exs Exchanges) error {
	for _, ex := range exs {
		err := db.AddExchange(ctx, ex)
		if err != nil {
			return err
		}
	}

	return nil
}
