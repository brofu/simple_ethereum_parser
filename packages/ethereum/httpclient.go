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
	logger     logging.Logger
}

func NewEthJsonRpcClient(entryPoint string, logger logging.Logger) EthereumChainAccesser {
	return &EthJsonRpcClient{
		entryPoint: entryPoint,
		logger:     logger,
	}
}

// EthGetCurrentBlockNumber get the block number
// TODO: Refactor these 2 functions
func (this *EthJsonRpcClient) EthGetCurrentBlockNumber(ctx context.Context, req *EthGetCurrentBlockNumberRequest) (int, error) {

	r := RPCRequest{
		Jsonrpc: JsonRpcVersion,
		Method:  MethodGetCurrentBlockNumber,
		ID:      req.RequestId,
	}

	rawReq, err := json.Marshal(r)
	if err != nil {
		this.logger.Errorf("marshal data fail | method: %s, err: %s", MethodGetCurrentBlockNumber, err.Error())
		return 0, err
	}

	httpReq, err := constructHttpRequest(ctx, http.MethodPost, this.entryPoint, contentType, bytes.NewBuffer(rawReq))
	if err != nil {
		this.logger.Errorf("construct request fail | method: %s, err: %s", MethodGetCurrentBlockNumber, err.Error())
		return 0, err
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		this.logger.Errorf("chain call fail | method: %s, err: %s", MethodGetCurrentBlockNumber, err.Error())
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		this.logger.Errorf("chain call fail | method: %s, StatusCode: %d", MethodGetCurrentBlockNumber, resp.StatusCode)
		return 0, errors.New("status code not equal 200 | status: " + resp.Status)
	}

	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		this.logger.Errorf("read response data fail| method: %s, err: %s", MethodGetCurrentBlockNumber, err.Error())
		return 0, err
	}

	data := RPCResponse{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		this.logger.Errorf("unmarshal response data fail | method: %s, err: %s", MethodGetCurrentBlockNumber, err)
		return 0, err
	}
	bnString := data.Result.(string)
	bnInt, err := strconv.ParseInt(bnString, 0, 64)
	if err != nil {
		this.logger.Errorf("convert block number fail | bnString: %s, err: %s", bnString, err)
		return 0, err
	}
	return int(bnInt), nil
}

func (this *EthJsonRpcClient) EthGetCurrentTransactionsByAddress(ctx context.Context, req *EthGetCurrentTransactionsByAddressRequest) ([]Transaction, error) {

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
		this.logger.Errorf("marshal params fail | method: %s, err: %s", MethodTraceFilter, err.Error())
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
		this.logger.Errorf("marshal data fail | method: %s, err: %s", MethodTraceFilter, err.Error())
		return nil, err
	}

	httpReq, err := constructHttpRequest(ctx, http.MethodPost, this.entryPoint, contentType, bytes.NewBuffer(rawReq))
	if err != nil {
		this.logger.Errorf("construct request fail | method: %s, err: %s", MethodTraceFilter, err.Error())
		return nil, err
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		this.logger.Errorf("chain call fail | method: %s, err: %s", MethodTraceFilter, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		this.logger.Errorf("chain call fail | method: %s, StatusCode: %s", MethodTraceFilter, resp.StatusCode)
		return nil, errors.New("status code not equal 200 | status: " + resp.Status)
	}

	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		this.logger.Errorf("read response data fail| method: %s, err: %s", MethodTraceFilter, err.Error())
		return nil, err
	}

	data := &RPCResponse{}
	err = json.Unmarshal(rawData, data)
	if err != nil {
		this.logger.Errorf("unmarshal response data fail | method: %s, err: %s", MethodTraceFilter, err)
		return nil, err
	}
	if data.Error != nil {
		this.logger.Errorf("get error from chain | method: %s, err code: %d, err msg: %s", MethodTraceFilter, data.Error.Code, data.Error.Message)
		return nil, err
	}

	// if get correct response from chain,
	// usually, there should be NO error for the following steps.
	rawTrx, err := json.Marshal(data.Result)
	if err != nil {
		this.logger.Errorf("converting transaction data fail | err: %s", err.Error())
		return nil, err
	}
	var res []Transaction
	err = json.Unmarshal(rawTrx, &res)
	if err != nil {
		this.logger.Errorf("converting transaction data fail | err: %s", err.Error())
		return nil, err
	}

	return res, nil
}

func constructHttpRequest(ctx context.Context, method, url, contentType string, body io.Reader) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return req, err
}
