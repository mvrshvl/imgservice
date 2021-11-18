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
	depositsIndex  = 1
	accountsIndex  = 2
	clustersIndex  = 3
	ownersIndex    = 4
	exchangesIndex = 5
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

	for i := 1; i <= len(nodeTypes)+1; i++ {
		categories = append(categories, &opts.GraphCategory{
			Name: nodeTypes[i-1],
			//Label: &opts.Label{
			//	Show:      false,
			//	Position:  "right",
			//	Formatter: nodeTypes[i-1],
			//},
		})
	}

	return
}

func newNode(name string, category int) opts.GraphNode {
	return opts.GraphNode{
		Name:       name,
		Category:   category,
		SymbolSize: 20,
		ItemStyle: &opts.ItemStyle{
			Opacity: 0,
		},
	}
}

// выделить тип чартс который будет выводить все графики, считать количество аккаунтов и т.п. и делать пирог со всеми подсчетами

func (cls Clusters) GenerateGraph(exchanges map[string]opts.GraphNode, tokenOwners map[string]opts.GraphNode, showSingleAccounts bool) *charts.Graph {
	nodes := make([]opts.GraphNode, 0)
	links := make([]opts.GraphLink, 0)

	isAdded := make(map[string]struct{})
	for _, node := range exchanges {
		//node.ItemStyle = newExchangeNode("").ItemStyle
		nodes = append(nodes, newNode(node.Name, exchangesIndex))
	}

	for _, node := range tokenOwners {
		//node.ItemStyle = newOwnerNode("").ItemStyle
		nodes = append(nodes, newNode(node.Name, ownersIndex))
	}

	for numCluster, cluster := range cls {
		if len(cluster.AccountsExchangeTransfers) == 1 && showSingleAccounts {
			continue
		}

		//clusterNode := opts.GraphNode{Name: fmt.Sprintf("cluster%d", numCluster), ItemStyle: newClusterNode("").ItemStyle}
		clusterNode := newNode(fmt.Sprintf("cluster%d", numCluster), clustersIndex)
		nodes = append(nodes, clusterNode)

		for account := range cluster.Accounts {
			//nodes = append(nodes, opts.GraphNode{Name: account, ItemStyle: newAccountNode("").ItemStyle})
			nodes = append(nodes, newNode(account, accountsIndex))

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

					//nodes = append(nodes, opts.GraphNode{Name: deposit, ItemStyle: newDepositNode("").ItemStyle})
					nodes = append(nodes, newNode(deposit, depositsIndex))
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
		charts.WithTitleOpts(opts.Title{Title: "Clustering"}),
		charts.WithLegendOpts(opts.Legend{Show: true, Right: "30%"}),
	)

	graph.AddSeries("graph", nodes, links,
		charts.WithGraphChartOpts(
			opts.GraphChart{Force: &opts.GraphForce{Repulsion: 500},
				Categories: getCategories()},
		),
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
