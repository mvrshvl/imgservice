package selfauth

import (
	"nir/clustering"
	"nir/clustering/blockchain"
)

type AddressApproves struct {
	address     string
	fromApprove map[string][]*blockchain.ERC20Approve
	toApprove   map[string][]*blockchain.ERC20Approve
}

func Find(approves blockchain.ERC20Approves) clustering.Clusters {
	addresses := getAddresses(approves)

	// Получить адреса для которых количество входящих аппрувов или исходящих аппрувов не больше 10
	/*
		SELECT * FROM transactions
		WHERE type = approve
		AND blockNumber BETWEEN ? AND >
		GROUP BY contractAddress, fromAddress

		SELECT contractAddress, toAddress, COUNT(*) FROM transactions
		WHERE type = approve
		AND blockNumber BETWEEN ? AND >
		GROUP BY contractAddress, toAddress

		SELECT * FROM transactions
		LEFT JOIN tab
		ON contractAddress = tab.contractAddress AND
		   toAddress = tab.toAddress AND
		WHERE transactions.blockNumber BETWEEN ? AND ?
		AND   transactions.type = approve.
	*/
	// аккаунты не должны быть биржами
	// Аккаунты по каждому контракту собрать в кластер

	for address, addressApproves := range addresses {
		if len(addressApproves.fromApprove) > 10 || len(addressApproves.toApprove) > 10 || len(addressApproves.fromApprove) > 5 || len(addressApproves.toApprove) > 5 {
			delete(addresses, address)
		}
	}

	tokens := toTokens(addresses)

	var cls clustering.Clusters

	for token, tokenAprroves := range tokens {
		cl := clustering.NewCluster()

		cl.TokensAuth[token] = tokenAprroves

		for _, tokenApprove := range tokenAprroves {
			cl.Accounts[tokenApprove.FromAddress] = struct{}{}
			cl.Accounts[tokenApprove.ToAddress] = struct{}{}
		}

		cls = append(cls, cl)
	}

	cls.Merge(cls)

	return cls
}

func getAddresses(approves blockchain.ERC20Approves) map[string]*AddressApproves {
	addresses := make(map[string]*AddressApproves)

	for _, approve := range approves {
		if _, ok := addresses[approve.ToAddress]; !ok {
			addresses[approve.ToAddress] = &AddressApproves{
				address:     approve.ToAddress,
				fromApprove: make(map[string][]*blockchain.ERC20Approve),
				toApprove:   make(map[string][]*blockchain.ERC20Approve),
			}
		}

		addresses[approve.ToAddress].fromApprove[approve.FromAddress] = append(addresses[approve.ToAddress].fromApprove[approve.FromAddress], approve)

		if _, ok := addresses[approve.FromAddress]; !ok {
			addresses[approve.FromAddress] = &AddressApproves{
				address:     approve.FromAddress,
				fromApprove: make(map[string][]*blockchain.ERC20Approve),
				toApprove:   make(map[string][]*blockchain.ERC20Approve),
			}
		}

		addresses[approve.FromAddress].toApprove[approve.ToAddress] = append(addresses[approve.FromAddress].toApprove[approve.ToAddress], approve)
	}

	return addresses
}

func toTokens(addresses map[string]*AddressApproves) map[string]map[string]*blockchain.ERC20Approve {
	tokens := make(map[string]map[string]*blockchain.ERC20Approve)
	for _, approves := range addresses {
		for _, from := range approves.fromApprove {
			for _, approve := range from {
				if _, ok := tokens[approve.ContractAddress]; !ok {
					tokens[approve.ContractAddress] = make(map[string]*blockchain.ERC20Approve)
				}

				tokens[approve.ContractAddress][approve.TxHash] = approve
			}
		}

		for _, to := range approves.toApprove {
			for _, approve := range to {
				if _, ok := tokens[approve.ContractAddress]; !ok {
					tokens[approve.ContractAddress] = make(map[string]*blockchain.ERC20Approve)
				}

				tokens[approve.ContractAddress][approve.TxHash] = approve
			}
		}
	}

	return tokens
}
