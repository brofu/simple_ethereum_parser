package ethereum

import "context"

//go:generate mockgen -destination=../ethereum/mocks/mock_ethereum.go -package=mocks github.com/brofu/simple_ethereum_parser/packages/ethereum EthereumChainAccesser

type Action struct {
	From     string
	CallType string
	Gas      string
	Input    string
	To       string
	Value    string
}

type Result struct {
	GasUsed string
	Output  string
}

type Transaction struct {
	Action              Action
	BlockHash           string
	BlockNumber         int
	Result              Result
	Subtraces           int
	TraceAddress        []string
	TransactionHash     string
	TransactionPosition int
	Type                string
}

type EthGetCurrentBlockNumberRequest struct {
	RequestId string
}

type EthGetCurrentTransactionsByAddressRequest struct {
	FromBlock   string
	ToBlock     string
	FromAddress string
	ToAddress   string
	RequestId   string
}

type EthereumChainAccesser interface {
	EthGetCurrentTransactionsByAddress(context.Context, *EthGetCurrentTransactionsByAddressRequest) ([]Transaction, error)
	EthGetCurrentBlockNumber(context.Context, *EthGetCurrentBlockNumberRequest) (int, error)
}
