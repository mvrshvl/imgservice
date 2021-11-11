package depositreuse

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"nir/clustering/transfer"
)

type Cluster struct {
	Accounts map[string][]*transfer.ExchangeTransfer
}

type Clusters []*Cluster

func NewCluster() *Cluster {
	return &Cluster{Accounts: make(map[string][]*transfer.ExchangeTransfer)}
}

func (cl *Cluster) AddTransfer(transfer *transfer.ExchangeTransfer) {
	for _, ts := range cl.Accounts[transfer.TxToDeposit.FromAddress] {
		if ts.TxToDeposit.Hash == transfer.TxToDeposit.Hash {
			return
		}
	}

	cl.Accounts[transfer.TxToDeposit.FromAddress] = append(cl.Accounts[transfer.TxToDeposit.FromAddress], transfer)
}

func (cl *Cluster) AddTransfers(transfers []*transfer.ExchangeTransfer) {
	for _, t := range transfers {
		cl.AddTransfer(t)
	}
}

func (cl *Cluster) HasAnAccounts(accs map[string][]*transfer.ExchangeTransfer) bool {
	for acc := range accs {
		if _, ok := cl.Accounts[acc]; ok {
			return true
		}
	}

	return false
}

func (cl *Cluster) MergeMatches(currentDeposit string, depositsWithSenders map[string]*Cluster) {
	for deposit, cluster := range depositsWithSenders {
		if !cl.HasAnAccounts(cluster.Accounts) || currentDeposit == deposit {
			continue
		}

		for _, transfers := range cluster.Accounts {
			cl.AddTransfers(transfers)
		}

		delete(depositsWithSenders, deposit)
	}
}

func Find(transfers []*transfer.ExchangeTransfer) Clusters {
	depositsWithSenders := make(map[string]*Cluster)

	for _, t := range transfers {
		if depositsWithSenders[t.TxToDeposit.ToAddress] == nil {
			depositsWithSenders[t.TxToDeposit.ToAddress] = NewCluster()
		}

		depositsWithSenders[t.TxToDeposit.ToAddress].AddTransfer(t)
	}

	noMatches := make(map[string]*Cluster)
	for deposit, cluster := range depositsWithSenders {
		noMatches[deposit] = cluster
	}

	for deposit, cluster := range noMatches {
		cluster.MergeMatches(deposit, noMatches)
	}

	var clusters []*Cluster
	for _, cluster := range noMatches {
		clusters = append(clusters, cluster)
	}

	return clusters
}

func (cls Clusters) GenerateGraph(exchanges map[string]opts.GraphNode) *charts.Graph {
	nodes := make([]opts.GraphNode, 0)
	links := make([]opts.GraphLink, 0)

	isAdded := make(map[string]struct{})
	for _, node := range exchanges {
		nodes = append(nodes, node)
	}

	for numCluster, cluster := range cls {
		clusterNode := opts.GraphNode{Name: fmt.Sprintf("cluster%d", numCluster)}
		nodes = append(nodes, clusterNode)

		for _, transfers := range cluster.Accounts {
			for _, ts := range transfers {
				exchange, ok := exchanges[ts.TxToExchange.ToAddress]
				if !ok {
					continue
				}

				account := fmt.Sprintf("account-%s", ts.TxToDeposit.FromAddress)
				deposit := fmt.Sprintf("deposit-%s", ts.TxToDeposit.ToAddress)

				if _, ok := isAdded[account]; !ok {
					isAdded[account] = struct{}{}

					nodes = append(nodes, opts.GraphNode{Name: account})
				}

				if _, ok := isAdded[deposit]; !ok {
					isAdded[deposit] = struct{}{}

					nodes = append(nodes, opts.GraphNode{Name: deposit})
				}

				links = append(links, opts.GraphLink{Source: clusterNode.Name, Target: account})
				links = append(links, opts.GraphLink{Source: account, Target: deposit})
				links = append(links, opts.GraphLink{Source: deposit, Target: exchange.Name})
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
