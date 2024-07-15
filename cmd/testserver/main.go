package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"time"

	"github.com/brofu/simple_ethereum_parser/packages/ethereum"
)

func main() {
	http.HandleFunc("/rpc", rpcHandler)
	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", nil)
}

func rpcHandler(w http.ResponseWriter, r *http.Request) {
	var req ethereum.RPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, -32700, "Parse error", nil)
		return
	}

	var result interface{}
	var err *ethereum.RPCError

	switch req.Method {
	case "eth_blockNumber":
		result, err = getBlockNumber(req.Params)
	case "trace_filter":
		result, err = traceFilter(req.Params)
	default:
		err = &ethereum.RPCError{Code: -32601, Message: "Method not found"}
	}

	response := ethereum.RPCResponse{
		Jsonrpc: "2.0",
		ID:      req.ID,
	}

	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}

	json.NewEncoder(w).Encode(response)
}
func getBlockNumber(params json.RawMessage) (interface{}, *ethereum.RPCError) {
	result := fmt.Sprintf("%d", time.Now().Unix())
	return result, nil
}

func convertBN(n string) int {
	bn, err := strconv.Atoi(n)
	if err != nil {
		return time.Now().Minute()
	}
	return bn
}

func traceFilter(params json.RawMessage) (interface{}, *ethereum.RPCError) {
	req := []ethereum.JsonRpcTraceFilterParams{}

	if err := json.Unmarshal(params, &req); err != nil {
		fmt.Println(err)
		return nil, &ethereum.RPCError{Code: -32602, Message: "Invalid params"}
	}

	bh := fmt.Sprintf("%d", time.Now().UnixNano())

	result := []ethereum.Transaction{
		{
			BlockNumber:     convertBN(req[0].FromBlock),
			TransactionHash: bh,
		},
		{
			BlockNumber:     convertBN(req[0].ToBlock),
			TransactionHash: bh,
		},
	}

	return result, nil
}

func respondWithError(w http.ResponseWriter, code int, message string, id interface{}) {
	response := ethereum.RPCResponse{
		Jsonrpc: "2.0",
		Error:   &ethereum.RPCError{Code: code, Message: message},
		ID:      id,
	}
	json.NewEncoder(w).Encode(response)
}
