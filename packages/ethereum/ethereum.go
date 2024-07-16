package ethereum

import "context"

//go:generate mockgen -destination=../ethereum/mocks/mock_ethereum.go -package=mocks github.com/brofu/simple_ethereum_parser/packages/ethereum EthereumChainAccesser

type Action struct {
	From     string `json:"from"`
	CallType string `json:"callType"`
	Gas      string `json:"gas"`
	Input    string `json:"input"`
	To       string `json:"to"`
	Value    string `json:"value"`
}

type Result struct {
	GasUsed string `json:"gasUsed"`
	Output  string `json:"output"`
}

type Transaction struct {
	Action              Action   `json:"action"`
	BlockHash           string   `json:"blockHash"`
	BlockNumber         int      `json:"blockNumber"`
	Result              Result   `json:"result"`
	Subtraces           int      `json:"subtraces"`
	TraceAddress        []string `json:"traceAddress"`
	TransactionHash     string   `json:"transactionHash"`
	TransactionPosition int      `json:"transactionPosition"`
	Type                string   `json:"type"`
}

type EthGetCurrentBlockNumberRequest struct {
	RequestId string `json:"request_id"`
}

type EthGetCurrentTransactionsByAddressRequest struct {
	FromBlock   string `json:"from_block"`
	ToBlock     string `json:"to_block"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	RequestId   string `json:"request_id"`
}

type EthereumChainAccesser interface {
	EthGetCurrentTransactionsByAddress(context.Context, *EthGetCurrentTransactionsByAddressRequest) ([]Transaction, error)
	EthGetCurrentBlockNumber(context.Context, *EthGetCurrentBlockNumberRequest) (int, error)
}
