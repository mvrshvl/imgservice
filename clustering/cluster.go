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

func (cls Clusters) GenerateGraph(exchanges map[string]opts.GraphNode, tokenOwners map[string]opts.GraphNode, showSingleAccounts bool) *charts.Graph {
	nodes := make([]opts.GraphNode, 0)
	links := make([]opts.GraphLink, 0)

	accountStyle := &opts.ItemStyle{Color: "#fc8452"}  //orange
	depositStyle := &opts.ItemStyle{Color: "#f9e215"}  //yellow
	clusterStyle := &opts.ItemStyle{Color: "#f92a13"}  //red
	exchangeStyle := &opts.ItemStyle{Color: "#3ba272"} //green
	ownersStyle := &opts.ItemStyle{Color: "#44bcba"}   //blue

	isAdded := make(map[string]struct{})
	for _, node := range exchanges {
		node.ItemStyle = exchangeStyle
		nodes = append(nodes, node)
	}

	for _, node := range tokenOwners {
		node.ItemStyle = ownersStyle
		nodes = append(nodes, node)
	}

	for numCluster, cluster := range cls {
		if len(cluster.AccountsExchangeTransfers) == 1 && showSingleAccounts {
			continue
		}

		clusterNode := opts.GraphNode{Name: fmt.Sprintf("cluster%d", numCluster), ItemStyle: clusterStyle}
		nodes = append(nodes, clusterNode)

		for account := range cluster.Accounts {
			nodes = append(nodes, opts.GraphNode{Name: account, ItemStyle: accountStyle})

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

					nodes = append(nodes, opts.GraphNode{Name: deposit, ItemStyle: depositStyle})
				}

				links = append(links, opts.GraphLink{Source: clusterNode.Name, Target: account})
				links = append(links, opts.GraphLink{Source: account, Target: deposit})
				links = append(links, opts.GraphLink{Source: deposit, Target: exchange.Name})
			}
		}

		for toAccount, transfers := range cluster.AccountsTokenTransfers {
			for _, ts := range transfers {
				toAccountNode := opts.GraphNode{Name: toAccount, ItemStyle: accountStyle}
				var fromAccountNode opts.GraphNode

				_, isOwner := tokenOwners[ts.FromAddress]
				if isOwner {
					fromAccountNode = tokenOwners[ts.FromAddress]
				} else {
					fromAccountNode = opts.GraphNode{Name: ts.FromAddress, ItemStyle: accountStyle}
				}

				links = append(links, opts.GraphLink{Source: fromAccountNode.Name, Target: toAccountNode.Name})
			}
		}
	}

	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "graph deposits reuse"}),
	)

	graph.AddSeries("graph", nodes, links,
		charts.WithGraphChartOpts(
			opts.GraphChart{Force: &opts.GraphForce{Repulsion: 100}},
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
