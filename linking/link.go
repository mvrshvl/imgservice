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
	account       *database.Account
	parent        *Node
	txsWithParent database.Transactions
	childs        map[string]*Node
	cache         map[string]*Node
}

func New(ctx context.Context, address string, parent *Node, cache map[string]*Node) (*Node, error) {
	acc, err := getAccount(ctx, address)
	if err != nil {
		return nil, err
	}

	return &Node{
		account: acc,
		parent:  parent,
		childs:  make(map[string]*Node),
		cache:   cache,
	}, nil
}

func Run(ctx context.Context, address string) (*[]AddressLinks, error) {
	entity, err := getEntity(ctx, address)
	if err != nil {
		return nil, err
	}

	nodesCh := make(chan *AddressLinks, len(entity.Accounts))

	for _, addr := range entity.Accounts {

		addr := addr
		go func() {
			links, err := LinkNode(ctx, addr.Address)
			if err != nil {
				fmt.Println("can't link node", addr.Address, err)
			}

			nodesCh <- links
		}()
	}

	nodes := make([]*AddressLinks, len(entity.Accounts))

	for i := range nodes {
		nodes[i] = <-nodesCh
	}

	return nil, err
}

func LinkNode(ctx context.Context, address string) (*AddressLinks, error) {
	cache := make(map[string]*Node)

	root, err := New(ctx, address, nil, cache)
	if err != nil {
		return nil, err
	}

	err = root.linking(ctx)
	if err != nil {
		return nil, err
	}

	linksByCurrency, err := root.splitByCurrency(ctx)
	if err != nil {
		return nil, err
	}

	return &AddressLinks{
		AllLinks:        root,
		LinksByCurrency: linksByCurrency,
	}, nil
}

func (n *Node) linking(ctx context.Context) error {
	if n.account.AccType == database.MinerAccount || n.account.AccType == database.ExchangeAccount && n.parent != nil {
		return nil
	}

	txs, err := getTransactions(ctx, n.account.Address)
	if err != nil {
		return err
	}

	for _, tx := range txs {
		address := tx.FromAddress

		if address == n.account.Address {
			address = tx.ToAddress
		}

		if n.parent != nil && address == n.parent.account.Address {
			continue
		}

		if n.parent != nil && n.isCycle(address) {
			continue
		}

		if ch, ok := n.childs[address]; ok {
			ch.txsWithParent = append(ch.txsWithParent, tx)

			continue
		}

		cachedNode, err := n.getCachedNode(ctx, address)
		if err != nil {
			return err
		}

		if cachedNode != nil {
			n.childs[cachedNode.account.Address] = cachedNode
			cachedNode.txsWithParent = append(cachedNode.txsWithParent, tx)

			continue
		}

		child, err := n.addChild(ctx, address, tx)
		if err != nil {
			return err
		}

		err = child.linking(ctx)
		if err != nil {
			return err
		}
	}

	n.cache[n.account.Address] = n

	return nil
}

func (n *Node) addChild(ctx context.Context, address string, tx *database.Transaction) (*Node, error) {
	child, err := New(ctx, address, n, n.cache)
	if err != nil {
		return nil, err
	}

	n.childs[child.account.Address] = child
	child.txsWithParent = append(child.txsWithParent, tx)

	return child, nil
}

func (n *Node) isCycle(address string) bool {
	if n.account.Address == address {
		return true
	}

	if n.parent == nil {
		return false
	}

	return n.parent.isCycle(address)
}

func (n *Node) getCachedNode(ctx context.Context, address string) (*Node, error) {
	cached, ok := n.cache[address]
	if !ok {
		return nil, nil
	}

	copyCached, err := New(ctx, cached.account.Address, n.parent, n.cache)
	if err != nil {
		return nil, err
	}

	copyCached.childs = cached.childs

	return copyCached, nil
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

func (node *Node) splitByCurrency(ctx context.Context) (linksByCurrency map[string]*Node, err error) {
	linksByCurrency = make(map[string]*Node)

	for _, tx := range node.txsWithParent {
		currency := wei

		if tx.ContractAddress != nil {
			currency = *tx.ContractAddress
		}

		if _, ok := linksByCurrency[currency]; !ok {
			linksByCurrency[currency], err = New(ctx, node.account.Address, node.parent, node.cache)
			if err != nil {
				return nil, err
			}
		}

		linksByCurrency[currency].txsWithParent = append(linksByCurrency[currency].txsWithParent, tx)
	}

	for _, node := range node.childs {
		childLinks, err := node.splitByCurrency(ctx)
		if err != nil {
			return nil, err
		}

		for c, child := range childLinks {
			if links, ok := linksByCurrency[c]; ok {
				links.childs[child.account.Address] = child
			}
		}
	}

	return linksByCurrency, nil
}
