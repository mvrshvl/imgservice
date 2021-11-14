package contract

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"nir/test/writer"
	"strings"
)

func BytecodeContract(bin, abiJSON string, args ...interface{}) ([]byte, error) {
	bytecode := common.FromHex(bin)

	ssAbi, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}

	ssInput, err := ssAbi.Pack("", args...)
	if err != nil {
		return nil, err
	}

	return append(bytecode, ssInput...), nil
}

func MethodContract(abiJSON string, name string, args ...interface{}) ([]byte, error) {
	ssAbi, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}

	data, err := ssAbi.Pack(name, args...)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func EstimateTransfer(ctx context.Context, fromAddr common.Address, toAddr *common.Address, amount *big.Int) (gas uint64, err error) {
	err = writer.Execute(ctx, func(w *writer.Writer) (innerErr error) {
		bytecodeTransfer, innerErr := MethodContract(SimpleTokenABI, "transfer", toAddr, big.NewInt(100))
		if err != nil {
			return err
		}

		gas, innerErr = w.EstimateGas(ctx, fromAddr, toAddr, bytecodeTransfer)

		return innerErr
	})

	return
}
