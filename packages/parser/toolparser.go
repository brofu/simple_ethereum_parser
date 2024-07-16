package parser

import (
	"github.com/brofu/simple_ethereum_parser/packages/ethereum"
	"github.com/brofu/simple_ethereum_parser/packages/logging"
)

type toolParser struct {
	logger        logging.Logger
	chainAccesser ethereum.EthereumChainAccesser
}

func NewToolParser(logger logging.Logger, chainAccesser ethereum.EthereumChainAccesser) Parser {
	return &toolParser{
		logger:        logger,
		chainAccesser: chainAccesser,
	}
}

func (this *toolParser) GetTransactions(address string) []ethereum.Transaction {

	// get current block number
	bn := this.GetCurrentBlock()
	if bn == 0 {
		this.logger.Errorf("get latest block number fail")
		return nil
	}

	req := &ethereum.EthGetCurrentTransactionsByAddressRequest{
		FromBlock:   "0x0",
		ToBlock:     convertDecimalToHex(bn),
		FromAddress: address,
		ToAddress:   address,
		RequestId:   generateRequestId(),
	}
	transactions, err := this.chainAccesser.EthGetCurrentTransactionsByAddress(nil, req)
	if err != nil {
		this.logger.Errorf("get error: %s", err.Error())
		return nil
	}
	return transactions
}

// Subscribe is not necessary for cmd tool scenarios
func (this *toolParser) Subscribe(address string) bool {
	return true
}

func (this *toolParser) GetCurrentBlock() int {
	req := &ethereum.EthGetCurrentBlockNumberRequest{
		RequestId: generateRequestId(),
	}

	bn, err := this.chainAccesser.EthGetCurrentBlockNumber(nil, req)
	if err != nil {
		this.logger.Errorf("get error: %s", err.Error())
		return 0
	}
	return bn
}
