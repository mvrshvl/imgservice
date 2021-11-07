package exchange

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"nir/test/entity/account"
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
