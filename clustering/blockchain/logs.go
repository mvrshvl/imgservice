package blockchain

const ApproveEventTopic = 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925

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

func (l Logs) ToApproves() ERC20Approves {
	return nil
}
