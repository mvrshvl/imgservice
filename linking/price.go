package linking

type Prices struct {
	currency map[string]Price
}

type Price struct {
	price  float64
	amount uint64
}
