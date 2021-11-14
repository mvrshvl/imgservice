package main

const (
	commonAmount = 10000000000

	countExchanges            = 1
	countCluster              = 1
	countAccounts             = 10
	countTransactions         = 50
	maxCountAccountsInCluster = 4
	transactonsPerSecond      = 30
	mixTokensDepth            = 4
)

func getWriterNodes() []string {
	return []string{
		"ws://localhost:59048",
		"ws://localhost:59050",
		"ws://localhost:59052",
		"ws://localhost:59054",
		"ws://localhost:59056",
	}
}

func getAccountsWithBalance() []string {
	return []string{
		"29f4455cd82770e305096f33c2a53f13efed2974873c883a6d7ca7d1bdcdf0c7",
		"4ff8a64d3f91b3ad9bb319d9118553ff1eed0696179fbf29957e67b5fb799449",
		"87d81138d9020bc465e412107572cffdad49f380b33ebe7be1bc4c365cddf5e1",
		"07d7884a48248cfd5e05687791604510c8574fd651ebd2bae3c3c916da2f3b4a",
		"cbabfda8a81814406e1d7fb4c2d4b902a07cb98f1fddccbaec25a3137a762a1c",
	}
}
