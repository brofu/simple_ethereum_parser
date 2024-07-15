package ethereum

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/brofu/simple_ethereum_parser/packages/logging"
)

var (
	entryPoint     string = "https://cloudflare-eth.com/"
	testEntryPoint string = "http://localhost:8080/rpc"
	contentType    string = "application/json"

	JsonRpcVersion              = "2.0"
	MethodTraceFilter           = "trace_filter"
	MethodGetCurrentBlockNumber = "eth_blockNumber"

	logger = logging.NewDefaultLogger(logging.LevelInfo)
)

type JsonRpcTraceFilterParams struct {
	FromBlock   string
	ToBlock     string
	FromAddress []string
	ToAddress   []string
}

type RPCRequest struct {
	Jsonrpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

type RPCResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// EthJsonRpcClient implements interface `EthereumChainAccesser`
// It is based on HTTP and JsonRPC 2.0
type EthJsonRpcClient struct {
	entryPoint string
}

func NewEthJsonRpcClient(entryPoint string) EthereumChainAccesser {
	return &EthJsonRpcClient{
		entryPoint: entryPoint,
	}
}

// EthGetCurrentBlockNumber get the block number
// TODO: Refactor these 2 functions
func (this *EthJsonRpcClient) EthGetCurrentBlockNumber(context context.Context, req *EthGetCurrentBlockNumberRequest) (int, error) {

	r := RPCRequest{
		Jsonrpc: JsonRpcVersion,
		Method:  MethodGetCurrentBlockNumber,
		ID:      req.RequestId,
	}

	rawReq, err := json.Marshal(r)
	if err != nil {
		logger.Errorf("marshal data fail | method: %s, err: %s", MethodGetCurrentBlockNumber, err.Error())
		return 0, err
	}

	resp, err := http.Post(this.entryPoint, contentType, bytes.NewBuffer(rawReq))
	if err != nil {
		logger.Errorf("chain call fail | method: %s, err: %s", MethodGetCurrentBlockNumber, err.Error())
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("chain call fail | method: %s, StatusCode: %s", MethodGetCurrentBlockNumber, resp.StatusCode)
		return 0, errors.New("status code not equal 200 | status: " + resp.Status)
	}

	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("read response data fail| method: %s, err: %s", MethodGetCurrentBlockNumber, err.Error())
		return 0, err
	}

	data := RPCResponse{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		logger.Errorf("unmarshal response data fail | method: %s, err: %s", MethodGetCurrentBlockNumber, err)
		return 0, err
	}
	bnString := data.Result.(string)
	bnInt, err := strconv.ParseInt(bnString, 0, 64)
	if err != nil {
		logger.Errorf("convert block number fail | bnString: %s, err: %s", bnString, err)
		return 0, err
	}
	return int(bnInt), nil
}

func (this *EthJsonRpcClient) EthGetCurrentTransactionsByAddress(context context.Context, req *EthGetCurrentTransactionsByAddressRequest) ([]Transaction, error) {

	params := []JsonRpcTraceFilterParams{
		{
			FromBlock:   req.FromBlock,
			ToBlock:     req.ToBlock,
			FromAddress: []string{req.FromAddress},
			ToAddress:   []string{req.ToAddress},
		},
	}

	rawParams, err := json.Marshal(params)

	if err != nil {
		logger.Errorf("marshal params fail | method: %s, err: %s", MethodTraceFilter, err.Error())
		return nil, err
	}

	r := RPCRequest{
		Jsonrpc: JsonRpcVersion,
		Method:  MethodTraceFilter,
		Params:  rawParams,
		ID:      req.RequestId,
	}
	rawReq, err := json.Marshal(r)
	if err != nil {
		logger.Errorf("marshal data fail | method: %s, err: %s", MethodTraceFilter, err.Error())
		return nil, err
	}

	resp, err := http.Post(this.entryPoint, contentType, bytes.NewBuffer(rawReq))
	if err != nil {
		logger.Errorf("chain call fail | method: %s, err: %s", MethodTraceFilter, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("chain call fail | method: %s, StatusCode: %s", MethodTraceFilter, resp.StatusCode)
		return nil, errors.New("status code not equal 200 | status: " + resp.Status)
	}

	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("read response data fail| method: %s, err: %s", MethodTraceFilter, err.Error())
		return nil, err
	}

	data := &RPCResponse{}
	err = json.Unmarshal(rawData, data)
	if err != nil {
		logger.Errorf("unmarshal response data fail | method: %s, err: %s", MethodTraceFilter, err)
		return nil, err
	}
	if data.Error != nil {
		logger.Errorf("get error from chain | method: %s, err code: %d, err msg: %s", MethodTraceFilter, data.Error.Code, data.Error.Message)
		return nil, err
	}

	// if get correct response from chain,
	// usually, there should be NO error for the following steps.
	ifs, _ := data.Result.([]interface{})
	res := make([]Transaction, len(ifs))
	for i, it := range ifs {
		trx, _ := it.(Transaction)
		res[i] = trx
	}
	return res, nil
}
