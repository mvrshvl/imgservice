package clustering

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"nir/clustering/transfer"
)

type Cluster struct {
	Accounts                  map[string]struct{}
	AccountsExchangeTransfers map[string][]*transfer.ExchangeTransfer
	AccountsTokenTransfers    map[string][]*transfer.TokenTransfer
}

type Clusters []*Cluster

func NewCluster() *Cluster {
	return &Cluster{Accounts: make(map[string]struct{}), AccountsExchangeTransfers: make(map[string][]*transfer.ExchangeTransfer), AccountsTokenTransfers: make(map[string][]*transfer.TokenTransfer)}
}

func (cl *Cluster) Merge(cluster *Cluster) bool {
	for acc := range cluster.Accounts {
		if _, ok := cl.Accounts[acc]; ok {
			cl.merge(cluster)

			return true
		}
	}

	return false
}

func (cl *Cluster) merge(cluster *Cluster) {
	for acc := range cluster.Accounts {
		cl.Accounts[acc] = struct{}{}

		if exchangeTransfers, ok := cluster.AccountsExchangeTransfers[acc]; ok {
			cl.AccountsExchangeTransfers[acc] = append(cl.AccountsExchangeTransfers[acc], exchangeTransfers...)
		}

		if tokenTransfers, ok := cluster.AccountsTokenTransfers[acc]; ok {
			cl.AccountsTokenTransfers[acc] = append(cl.AccountsTokenTransfers[acc], tokenTransfers...)
		}
	}
}

const (
	accountsIndex = iota
	depositsIndex
	clustersIndex
	ownersIndex
	exchangesIndex
)

func getNodeTypes() map[int]string {
	return map[int]string{
		accountsIndex:  "Accounts",
		depositsIndex:  "Deposits",
		clustersIndex:  "Clusters",
		ownersIndex:    "Owners",
		exchangesIndex: "Exchanges",
	}
}

func getCategories() (categories []*opts.GraphCategory) {
	nodeTypes := getNodeTypes()

	for i := 0; i < len(nodeTypes); i++ {
		categories = append(categories, &opts.GraphCategory{
			Name: nodeTypes[i],
			Label: &opts.Label{
				Show:      true,
				Color:     "#000000",
				Position:  "right",
				Formatter: nodeTypes[i],
			},
		})
	}

	return
}

func newAccountNode(name string) opts.GraphNode {
	return opts.GraphNode{
		Name:      name,
		Category:  accountsIndex,
		ItemStyle: &opts.ItemStyle{Color: "#fc8452"},
	}
}

func newDepositNode(name string) opts.GraphNode {
	return opts.GraphNode{
		Name:      name,
		Category:  depositsIndex,
		ItemStyle: &opts.ItemStyle{Color: "#f9e215"},
	}
}

func newClusterNode(name string) opts.GraphNode {
	return opts.GraphNode{
		Name:      name,
		Category:  clustersIndex,
		ItemStyle: &opts.ItemStyle{Color: "#f92a13"},
	}
}

func newExchangeNode(name string) opts.GraphNode {
	return opts.GraphNode{
		Name:      name,
		Category:  exchangesIndex,
		ItemStyle: &opts.ItemStyle{Color: "#3ba272"},
	}
}

func newOwnerNode(name string) opts.GraphNode {
	return opts.GraphNode{
		Name:      name,
		Category:  ownersIndex,
		ItemStyle: &opts.ItemStyle{Color: "#44bcba"},
	}
}

// выделить тип чартс который будет выводить все графики, считать количество аккаунтов и т.п. и делать пирог со всеми подсчетами

func (cls Clusters) GenerateGraph(exchanges map[string]opts.GraphNode, tokenOwners map[string]opts.GraphNode, showSingleAccounts bool) *charts.Graph {
	nodes := make([]opts.GraphNode, 0)
	links := make([]opts.GraphLink, 0)

	isAdded := make(map[string]struct{})
	for _, node := range exchanges {
		node.ItemStyle = newExchangeNode("").ItemStyle
		nodes = append(nodes, node)
	}

	for _, node := range tokenOwners {
		node.ItemStyle = newOwnerNode("").ItemStyle
		nodes = append(nodes, node)
	}

	for numCluster, cluster := range cls {
		if len(cluster.AccountsExchangeTransfers) == 1 && showSingleAccounts {
			continue
		}

		clusterNode := opts.GraphNode{Name: fmt.Sprintf("cluster%d", numCluster), ItemStyle: newClusterNode("").ItemStyle}
		nodes = append(nodes, clusterNode)

		for account := range cluster.Accounts {
			nodes = append(nodes, opts.GraphNode{Name: account, ItemStyle: newAccountNode("").ItemStyle})

			links = append(links, opts.GraphLink{Source: account, Target: clusterNode.Name})
		}

		for _, transfers := range cluster.AccountsExchangeTransfers {
			for _, ts := range transfers {
				exchange, ok := exchanges[ts.TxToExchange.ToAddress]
				if !ok {
					continue
				}

				account := ts.TxToDeposit.FromAddress
				deposit := ts.TxToDeposit.ToAddress

				if _, ok := isAdded[deposit]; !ok {
					isAdded[deposit] = struct{}{}

					nodes = append(nodes, opts.GraphNode{Name: deposit, ItemStyle: newDepositNode("").ItemStyle})
				}

				links = append(links, opts.GraphLink{Source: clusterNode.Name, Target: account})
				links = append(links, opts.GraphLink{Source: account, Target: deposit})
				links = append(links, opts.GraphLink{Source: deposit, Target: exchange.Name})
			}
		}

		for toAccount, transfers := range cluster.AccountsTokenTransfers {
			for _, ts := range transfers {
				links = append(links, opts.GraphLink{Source: ts.FromAddress, Target: toAccount})
			}
		}
	}

	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Clusters"}),
	)

	graph.AddSeries("graph", nodes, links,
		charts.WithGraphChartOpts(
			opts.GraphChart{Force: &opts.GraphForce{Repulsion: 1000}},
		),
	)

	return graph
}

func (cls Clusters) GenerateLegend() *charts.Graph {
	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Legend"}),
	)

	graph.AddSeries("graph", []opts.GraphNode{newClusterNode("c"), newDepositNode("d"), newExchangeNode("e"), newAccountNode("a"), newOwnerNode("o")}, nil,
		charts.WithGraphChartOpts(
			opts.GraphChart{Force: &opts.GraphForce{Repulsion: 100},
				Categories: getCategories()},
		),
		charts.WithLabelOpts(*getCategories()[accountsIndex].Label),
		charts.WithLabelOpts(*getCategories()[depositsIndex].Label),
		charts.WithLabelOpts(*getCategories()[ownersIndex].Label),
		charts.WithLabelOpts(*getCategories()[exchangesIndex].Label),
		charts.WithLabelOpts(*getCategories()[clustersIndex].Label),
	)

	return graph
}

func (cls Clusters) Merge(clusters Clusters) (newClusters Clusters) {
	for _, iCluster := range cls {
		for _, jCluster := range clusters {
			if ok := iCluster.Merge(jCluster); ok {
				newClusters = append(newClusters, iCluster)
			}
		}
	}

	return
}
