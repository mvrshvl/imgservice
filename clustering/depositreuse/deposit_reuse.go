package depositreuse

import "nir/clustering/transfer"

type Cluster struct {
	Accounts map[string] []*transfer.ExchangeTransfer
}

func NewCluster()*Cluster{
	return &Cluster{Accounts: make(map[string][]*transfer.ExchangeTransfer)}
}

func (cl *Cluster) Add(transfer *transfer.ExchangeTransfer) {
	cl.Accounts[transfer.TxToDeposit.FromAddress] = append(cl.Accounts[transfer.TxToDeposit.FromAddress], transfer)
}

func (cl *Cluster) HasAnAccounts(accs map[string][]*transfer.ExchangeTransfer) bool{
	for acc := range accs{
		if _, ok := cl.Accounts[acc]; ok {
			return true
		}
	}

	return false
}

func (cl *Cluster) Merge(cluster *Cluster){
	for acc, transfers := range cluster.Accounts {
		for _, transfer := range transfers {
			if
		}
		cl.Accounts[acc] = append(cl.Accounts[acc], transfers...)
	}
}

func Find(transfers []*transfer.ExchangeTransfer) {
	depositsWithSenders := make(map[string] *Cluster)

	for _, t := range transfers {
		if depositsWithSenders[t.TxToDeposit.ToAddress] == nil {
			depositsWithSenders[t.TxToDeposit.ToAddress] = NewCluster()
		}

		depositsWithSenders[t.TxToDeposit.ToAddress].Add(t)
	}

	users := make(map[int])

	for _, sendersA := range depositsWithSenders {
		for _, sendersB := range depositsWithSenders {
			if isOneUser(sendersA, sendersB) {
				users[]
			}
		}
	}
}

func isOneUser(sendersA, sendersB map[string] *transfer.ExchangeTransfer) bool {
	for sender := range sendersA{
		if _, ok := sendersB[sender]; ok {
			return ok
		}
	}

	return false
}

func merge(ds map[string] *Cluster) {
	for {


		for _, clusterA := range ds {
			for _, clusterB := range ds {
				if clusterA.HasAnAccounts(clusterB.Accounts) {

				}
			}
		}
	}
}