package database

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"nir/test/contract"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

const ApproveEventTopic = "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"

type Log struct {
	LogIndex    string `csv:"log_index"`
	Hash        string `csv:"transaction_hash"`
	Index       string `csv:"transaction_index"`
	BlockHash   string `csv:"block_hash"`
	BlockNumber string `csv:"block_number"`
	Address     string `csv:"address"`
	Data        string `csv:"data"`
	Topics      string `csv:"topics"`
}

type Logs []*Log

type ERC20Approve struct {
	TxHash          string
	ContractAddress string
	FromAddress     string
	ToAddress       string
}

type ERC20Approves []*ERC20Approve

func (logs Logs) ToApproves(txs Transactions) (erc20Approves ERC20Approves, err error) {
	approves := make(map[string]struct{})

	for _, l := range logs {
		if topics := strings.Split(l.Topics, ","); len(l.Topics) != 0 && topics[0] == ApproveEventTopic {
			approves[l.Hash] = struct{}{}
		}
	}

	contractAbi, err := abi.JSON(strings.NewReader(contract.ERC20ABI))
	if err != nil {
		return nil, err
	}

	for _, tx := range txs {
		if _, ok := approves[tx.Hash]; ok {
			decodedData, err := hex.DecodeString(tx.Input[10:])
			if err != nil {
				return nil, err
			}

			inputInterface, err := contractAbi.Methods["approve"].Inputs.Unpack(decodedData)
			if err != nil {
				//log this
			}

			approve := &ERC20Approve{
				TxHash:          tx.Hash,
				ContractAddress: tx.ToAddress,
				FromAddress:     tx.FromAddress,
			}

			for _, in := range inputInterface {
				switch input := in.(type) {
				case common.Address:
					approve.ToAddress = strings.ToLower(input.String())
				}
			}

			erc20Approves = append(erc20Approves, approve)
		}
	}

	return
}
