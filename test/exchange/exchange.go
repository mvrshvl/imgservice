package exchange

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"nir/test/entity/account"
	"nir/test/writer"
)

type Exchange struct {
	name    string
	account *account.Account
	clients map[common.Address]*exchangeClient
}

type exchangeClient struct {
	account *common.Address
	deposit *account.Account
}

func CreateExchange() (*Exchange, error) {
	acc, err := account.CreateAccount()
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	return &Exchange{
		name:    id.String(),
		account: acc,
		clients: make(map[common.Address]*exchangeClient),
	}, nil
}

func (exch *Exchange) CreateAccountIfNotExist(address *common.Address) (*common.Address, error) {
	acc, err := account.CreateAccount()
	if err != nil {
		return nil, err
	}

	exchAcc, ok := exch.clients[*address]
	if ok {
		return exchAcc.deposit.GetAddress(), nil
	}

	exch.clients[*address] = &exchangeClient{
		account: address,
		deposit: acc,
	}

	return acc.GetAddress(), nil
}

func (exch *Exchange) GetName() string {
	return exch.name
}

func (exch *Exchange) GetEthFromDeposit(ctx context.Context, acc *common.Address, amount int64) error {
	wr, err := writer.FromContext(ctx)
	if err != nil {
		return err
	}

	for _, client := range exch.clients {
		address := client.deposit.GetAddress()
		if address == acc {

			err := account.AddEtherToAccount(ctx, wr, address, writer.GasLimit)
			if err != nil {
				return err
			}
			_, err = client.deposit.SendTransaction(ctx, exch.account.GetAddress(), amount)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (exch *Exchange) GetAccounts() (addresses []*common.Address) {
	for _, client := range exch.clients {
		addresses = append(addresses, client.deposit.GetAddress())
	}

	return addresses
}
