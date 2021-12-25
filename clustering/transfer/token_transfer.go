package transfer

import "math/big"

type TokenTransfer struct {
	TokenAddress string
	FromAddress  string
	ToAddress    string
	Value        big.Int
}
