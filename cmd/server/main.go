package main

import (
	"context"
	"encoding/json"
	"net/http"

	"time"

	"github.com/brofu/simple_ethereum_parser/packages/ethereum"
	"github.com/brofu/simple_ethereum_parser/packages/logging"
	"github.com/brofu/simple_ethereum_parser/packages/parser"
	"github.com/brofu/simple_ethereum_parser/protocol"
)

var (
	entryPoint     = "https://cloudflare-eth.com/"
	testEntryPoint = "http://localhost:8080/rpc"
)

func main() {
	context := context.Background()
	logger := logging.NewDefaultLogger(logging.LevelDebug)
	chainAccesser := ethereum.NewEthJsonRpcClient(testEntryPoint, logger)
	config := parser.ServiceParserConfiguration{
		MaxAddressNumber:     100,
		MaxTransactionNumber: 100,
		MaxConcurrentThreads: 10,
		Interval:             time.Millisecond * 5000,
	}
	parser := parser.NewServiceParser(context, logger, chainAccesser, config)

	handler := &Handler{
		parser: parser,
		logger: logger,
	}

	http.HandleFunc("/get-block-number", handler.GetBlockNumber)
	http.HandleFunc("/get-transactions", handler.GetTransactions)
	http.HandleFunc("/subscribe", handler.Subscribe)

	logger.Infof("Starting server on :8081...")
	http.ListenAndServe(":8081", nil)
}

type Handler struct {
	parser parser.Parser
	logger logging.Logger
}

func (this *Handler) GetBlockNumber(w http.ResponseWriter, r *http.Request) {

	var req protocol.JsonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		this.logger.Errorf("decode request fail | err: %s", err.Error())
		respondWithError(w, protocol.ErrCodeUnmarl, protocol.ErrMsgUnmarl, "")
		return
	}

	bn := this.parser.GetCurrentBlock()

	resp := protocol.JsonResponse{
		RequestId: req.RequestId,
		Result:    bn,
	}

	json.NewEncoder(w).Encode(resp)
}

func (this *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {

	var req protocol.JsonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		this.logger.Errorf("decode request fail | err: %s", err.Error())
		respondWithError(w, protocol.ErrCodeUnmarl, protocol.ErrMsgUnmarl, "")
		return
	}

	var params protocol.GetTransactionsParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		this.logger.Errorf("unmarl params fail | err: %s", err.Error())
		respondWithError(w, protocol.ErrCodeUnmarl, protocol.ErrMsgUnmarl, "")
		return
	}

	transactions := this.parser.GetTransactions(params.Address)

	resp := protocol.JsonResponse{
		RequestId: req.RequestId,
		Result:    transactions,
	}

	json.NewEncoder(w).Encode(resp)
}

func (this *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {

	var req protocol.JsonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		this.logger.Errorf("decode request fail | err: %s", err.Error())
		respondWithError(w, protocol.ErrCodeUnmarl, protocol.ErrMsgUnmarl, "")
		return
	}

	var params protocol.SubscribeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		this.logger.Errorf("unmarl params fail | err: %s", err.Error())
		respondWithError(w, protocol.ErrCodeUnmarl, protocol.ErrMsgUnmarl, "")
		return
	}

	success := this.parser.Subscribe(params.Address)

	resp := protocol.JsonResponse{
		RequestId: req.RequestId,
		Result:    success,
	}

	json.NewEncoder(w).Encode(resp)
}

func respondWithError(w http.ResponseWriter, code int, message string, id string) {
	response := protocol.JsonResponse{
		Error:     protocol.Error{Code: code, Message: message},
		RequestId: id,
	}
	json.NewEncoder(w).Encode(response)
}
