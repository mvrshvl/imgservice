package linking

import (
	"context"
	"fmt"
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
	cache         map[string]*Node
}

func New(address string, parent *Node, cache map[string]*Node) *Node {
	return &Node{
		address: address,
		parent:  parent,
		childs:  make(map[string]*Node),
		cache:   cache,
	}
}

func Run(ctx context.Context, address string) (*[]AddressLinks, error) {
	entity, err := getEntity(ctx, address)
	if err != nil {
		return nil, err
	}

	nodesCh := make(chan *Node, len(entity.Accounts))
	// надо переписать на кластеризацию внутри Link
	for _, addr := range entity.Accounts {
		addr := addr
		go func() {
			_, err := LinkNode(ctx, addr.Address)
			if err != nil {
				fmt.Println("can't link node", addr.Address, err)
			}
		}()
	}
	var (
		nodes []*Node
	)
}

func LinkNode(ctx context.Context, address string) (*AddressLinks, error) {
	cache := make(map[string]*Node)

	root := New(address, nil, cache)

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

	if acc.AccType == database.MinerAccount || acc.AccType == database.ExchangeAccount && n.parent != nil {
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

		if n.parent != nil && n.isCycle(address) {
			continue
		}

		if ch, ok := n.childs[address]; ok {
			ch.txsWithParent = append(ch.txsWithParent, tx)

			continue
		}

		cachedNode := n.getCachedNode(address)
		if cachedNode != nil {
			n.childs[cachedNode.address] = cachedNode
			cachedNode.txsWithParent = append(cachedNode.txsWithParent, tx)

			continue
		}

		err := n.addChild(address, tx).linking(ctx)
		if err != nil {
			return err
		}
	}

	n.cache[n.address] = n

	return nil
}

func (n *Node) addChild(address string, tx *database.Transaction) *Node {
	child := New(address, n, n.cache)

	n.childs[child.address] = child
	child.txsWithParent = append(child.txsWithParent, tx)

	return child
}

func (n *Node) isCycle(address string) bool {
	if n.address == address {
		return true
	}

	if n.parent == nil {
		return false
	}

	return n.parent.isCycle(address)
}

func (n *Node) getCachedNode(address string) *Node {
	cached, ok := n.cache[address]
	if !ok {
		return nil
	}

	copyCached := New(cached.address, n.parent, n.cache)
	copyCached.childs = cached.childs

	return copyCached
}

func getAccount(ctx context.Context, address string) (acc *database.Account, err error) {
	err = di.FromContext(ctx).Invoke(func(db *database.Database) (innerErr error) {
		acc, innerErr = db.GetAccount(ctx, address)

		return innerErr
	})

	return
}

func getEntity(ctx context.Context, address string) (entity *database.Entity, err error) {
	err = di.FromContext(ctx).Invoke(func(db *database.Database) (innerErr error) {
		entity, innerErr = db.GetEntity(ctx, address)

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

		return nil
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
			linksByCurrency[currency] = New(node.address, node.parent, node.cache)
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
