package linking

import (
	"context"
	"nir/clustering/common"
	"nir/database"
	"nir/di"
)

const wei = "wei"

type AddressLinks struct {
	AllLinks        *Node
	LinksByCurrency map[string]*Node
}

type Node struct {
	address       string
	parent        *Node
	txsWithParent database.Transactions
	childs        map[string]*Node
}

func New(address string, parent *Node) *Node {
	return &Node{
		address: address,
		parent:  parent,
		childs:  make(map[string]*Node),
	}
}

func Run(ctx context.Context, address string) (*AddressLinks, error) {
	root := New(address, nil)

	err := root.linking(ctx)
	if err != nil {
		return nil, err
	}

	return &AddressLinks{
		AllLinks:        root,
		LinksByCurrency: root.splitByCurrency(),
	}, nil
}

func (n *Node) linking(ctx context.Context) error {
	acc, err := getAccount(ctx, n.address)
	if err != nil {
		return err
	}

	if acc.AccType == database.MinerAccount || acc.AccType == database.ExchangeAccount {
		return nil
	}

	txs, err := getTransactions(ctx, n.address)
	if err != nil {
		return err
	}

	for _, tx := range txs {
		address := tx.FromAddress

		if address == n.address {
			address = tx.ToAddress
		}

		if n.parent != nil && address == n.parent.address {
			continue
		}

		if ch, ok := n.childs[address]; ok {
			ch.txsWithParent = append(ch.txsWithParent, tx)
		}

		node := New(address, n)

		node.childs[node.address] = node

		err := node.linking(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func getAccount(ctx context.Context, address string) (acc *database.Account, err error) {
	err = di.FromContext(ctx).Invoke(func(db *database.Database) (innerErr error) {
		acc, innerErr = db.GetAccount(ctx, address)

		return innerErr
	})

	return
}

func getTransactions(ctx context.Context, address string) (txs database.Transactions, err error) {
	err = di.FromContext(ctx).Invoke(func(db *database.Database) (innerErr error) {
		txs, innerErr = db.GetTransactionsByAddress(ctx, address)
		if innerErr != nil {
			return innerErr
		}

		return common.Clustering(ctx, txs)
	})

	return
}

func (node *Node) splitByCurrency() map[string]*Node {
	linksByCurrency := make(map[string]*Node)

	for _, tx := range node.txsWithParent {
		currency := wei

		if tx.ContractAddress != nil {
			currency = *tx.ContractAddress
		}

		if _, ok := linksByCurrency[currency]; !ok {
			linksByCurrency[currency] = New(node.address, node.parent)
		}

		linksByCurrency[currency].txsWithParent = append(linksByCurrency[currency].txsWithParent, tx)
	}

	for _, node := range node.childs {
		childLinks := node.splitByCurrency()
		for c, child := range childLinks {
			if links, ok := linksByCurrency[c]; ok {
				links.childs[child.address] = child
			}
		}
	}

	return linksByCurrency
}
