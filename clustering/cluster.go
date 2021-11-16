package clustering

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"nir/clustering/transfer"
)

type Cluster struct {
	AccountsExchangeTransfers map[string][]*transfer.ExchangeTransfer
	AccountsTokenTransfers    map[string][]*transfer.TokenTransfer
}

type Clusters []*Cluster

func NewCluster() *Cluster {
	return &Cluster{AccountsExchangeTransfers: make(map[string][]*transfer.ExchangeTransfer), AccountsTokenTransfers: make(map[string][]*transfer.TokenTransfer)}
}

func (cl *Cluster) AddTransfer(transfer *transfer.ExchangeTransfer) {
	for _, ts := range cl.AccountsExchangeTransfers[transfer.TxToDeposit.FromAddress] {
		if ts.TxToDeposit.Hash == transfer.TxToDeposit.Hash {
			return
		}
	}

	cl.AccountsExchangeTransfers[transfer.TxToDeposit.FromAddress] = append(cl.AccountsExchangeTransfers[transfer.TxToDeposit.FromAddress], transfer)
}

func (cl *Cluster) AddTransfers(transfers []*transfer.ExchangeTransfer) {
	for _, t := range transfers {
		cl.AddTransfer(t)
	}
}

func (cl *Cluster) HasAnAccounts(accs map[string][]*transfer.ExchangeTransfer) bool {
	for acc := range accs {
		if _, ok := cl.AccountsExchangeTransfers[acc]; ok {
			return true
		}
	}

	return false
}

func (cl *Cluster) MergeMatches(currentDeposit string, depositsWithSenders map[string]*Cluster) {
	for deposit, cluster := range depositsWithSenders {
		if !cl.HasAnAccounts(cluster.AccountsExchangeTransfers) || currentDeposit == deposit {
			continue
		}

		for _, transfers := range cluster.AccountsExchangeTransfers {
			cl.AddTransfers(transfers)
		}

		delete(depositsWithSenders, deposit)
	}
}

func (cls Clusters) GenerateGraph(exchanges map[string]opts.GraphNode, showSingleAccounts bool) *charts.Graph {
	nodes := make([]opts.GraphNode, 0)
	links := make([]opts.GraphLink, 0)

	accountStyle := &opts.ItemStyle{Color: "#fc8452"}  //orange
	depositStyle := &opts.ItemStyle{Color: "#f9e215"}  //yellow
	clusterStyle := &opts.ItemStyle{Color: "#f92a13"}  //red
	exchangeStyle := &opts.ItemStyle{Color: "#3ba272"} //green

	isAdded := make(map[string]struct{})
	for _, node := range exchanges {
		node.ItemStyle = exchangeStyle
		nodes = append(nodes, node)
	}

	for numCluster, cluster := range cls {
		if len(cluster.AccountsExchangeTransfers) == 1 && showSingleAccounts {
			continue
		}

		clusterNode := opts.GraphNode{Name: fmt.Sprintf("cluster%d", numCluster), ItemStyle: clusterStyle}
		nodes = append(nodes, clusterNode)

		for _, transfers := range cluster.AccountsExchangeTransfers {
			for _, ts := range transfers {
				exchange, ok := exchanges[ts.TxToExchange.ToAddress]
				if !ok {
					continue
				}

				account := fmt.Sprintf("account-%s", ts.TxToDeposit.FromAddress)
				deposit := fmt.Sprintf("deposit-%s", ts.TxToDeposit.ToAddress)

				if _, ok := isAdded[account]; !ok {
					isAdded[account] = struct{}{}

					nodes = append(nodes, opts.GraphNode{Name: account, ItemStyle: accountStyle})
				}

				if _, ok := isAdded[deposit]; !ok {
					isAdded[deposit] = struct{}{}

					nodes = append(nodes, opts.GraphNode{Name: deposit, ItemStyle: depositStyle})
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
