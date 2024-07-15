package protocol

import (
	"encoding/json"

	"github.com/brofu/simple_ethereum_parser/packages/ethereum"
)

var (
	//
	ErrCodeUnmarl = -100
	ErrMsgUnmarl  = "Unmarshal request error"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type JsonRequest struct {
	RequestId string `json:"request_id"`
	Params    json.RawMessage
}

type JsonResponse struct {
	RequestId string      `json:"request_id"`
	Result    interface{} `json:"result,omitempty"`
	Error     Error       `json:"error,omitempty"`
}

type GetBlockNumberResult struct {
	BlockNumber int `json:"block_number"`
}

type GetTransactionsParams struct {
	Address string `json:"address"`
}

type SubscribeParams struct {
	Address string `json:"address"`
}

type GetTransactionsResult struct {
	Transactions []ethereum.Transaction `json:"transactions"`
}
