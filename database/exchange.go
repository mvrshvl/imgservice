package database

type Exchange struct {
	Address string `csv:"address"`
	Name    string `csv:"name"`
}

type Exchanges []*Exchange

func (exchanges Exchanges) MapAddresses() map[string]struct{} {
	exchs := make(map[string]struct{})
	for _, exch := range exchanges {
		exchs[exch.Address] = struct{}{}
	}

	return exchs
}
