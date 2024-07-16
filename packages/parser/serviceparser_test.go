package parser

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/brofu/simple_ethereum_parser/packages/ethereum"
	"github.com/brofu/simple_ethereum_parser/packages/ethereum/mocks"
	"github.com/brofu/simple_ethereum_parser/packages/logging"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_serviceParser_getBlockNum(t *testing.T) {
	type fields struct {
		processing           bool
		processedBlock       int
		interval             time.Duration
		MaxConcurrentThreads int
		MaxTransactionNumber int
		newAddrLock          sync.Mutex
		addrLock             sync.RWMutex
		addresses            *addressTransactionLRU
		transactionTasks     chan transactionTask
		newTaskNoti          chan int
		finishedTasks        chan struct{}
		newAddresses         []string
		chanAccesser         ethereum.EthereumChainAccesser
	}
	type args struct {
		context   context.Context
		RequestId string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      int
		wantError error
	}{
		{
			name: "normal case 1",
			args: args{
				context:   context.Background(),
				RequestId: "1",
			},
			want:      1024,
			wantError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			chanAccesser := mocks.NewMockEthereumChainAccesser(ctrl)

			req := &ethereum.EthGetCurrentBlockNumberRequest{RequestId: "1"}
			resp := tt.want
			chanAccesser.EXPECT().EthGetCurrentBlockNumber(tt.args.context, req).Return(resp, tt.wantError)

			parser := serviceParser{
				chainAccesser: chanAccesser,
			}

			num, err := parser.getBlockNum(tt.args.context, req)
			assert.Equal(t, tt.want, num)
			assert.Equal(t, tt.wantError, err)
		})
	}
}

func Test_serviceParser_constructGetTransactionRequest(t *testing.T) {
	type args struct {
		addr        string
		blockNum    int
		oldBlockNum int
	}
	tests := []struct {
		name            string
		context         context.Context
		config          ServiceParserConfiguration
		requestId       string
		initialBlockNum int
		args            args
		want            *ethereum.EthGetCurrentTransactionsByAddressRequest
	}{
		{
			name:    "normal case 1",
			context: context.Background(),
			config: ServiceParserConfiguration{
				MaxAddressNumber:     10,
				MaxConcurrentThreads: 10,
				MaxTransactionNumber: 10,
				Interval:             time.Minute * 100, // for test purpose
			},
			requestId:       generateRequestId(),
			initialBlockNum: 0,
			args: args{
				addr:        "0xffff",
				blockNum:    200,
				oldBlockNum: 100,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logger := logging.NewDefaultLogger(logging.LevelDebug)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			chainAccesser := mocks.NewMockEthereumChainAccesser(ctrl)
			req := &ethereum.EthGetCurrentBlockNumberRequest{RequestId: tt.requestId}
			resp := tt.initialBlockNum
			chainAccesser.EXPECT().EthGetCurrentBlockNumber(tt.context, req).Return(resp, nil)

			// Init Parser
			parser := &serviceParser{
				newAddrLock:          sync.Mutex{},
				addrLock:             sync.RWMutex{},
				transactionTasks:     make(chan transactionTask, tt.config.MaxConcurrentThreads),
				interval:             tt.config.Interval,
				maxConcurrentThreads: tt.config.MaxConcurrentThreads,
				maxTransactionNumber: tt.config.MaxTransactionNumber,
				chainAccesser:        chainAccesser,
				logger:               logger,
				addresses:            newAddressTransactionLRU(tt.config.MaxAddressNumber),
			}

			blockNum, err := parser.getBlockNum(tt.context, req)
			assert.Equal(t, tt.initialBlockNum, blockNum)
			assert.Equal(t, nil, err)

			parser.processedBlock = blockNum

			// start Parser
			go parser.start(tt.context)

			oldData := addressTransaction{
				address:  tt.args.addr,
				blockNum: tt.args.oldBlockNum,
			}

			parser.addresses.putAddress(oldData)
			got := parser.constructGetTransactionRequest(tt.args.addr, tt.args.blockNum)
			assert.Equal(t, tt.args.addr, got.FromAddress)
			assert.Equal(t, tt.args.addr, got.ToAddress)
			assert.Equal(t, convertDecimalToHex(tt.args.blockNum), got.ToBlock)
			assert.Equal(t, convertDecimalToHex(tt.args.oldBlockNum), got.FromBlock)
		})
	}
}

func Test_serviceParser_doUpdateTransactions(t *testing.T) {
	type args struct {
		req     *ethereum.EthGetCurrentTransactionsByAddressRequest
		resp    []ethereum.Transaction
		respErr error
	}
	tests := []struct {
		name            string
		context         context.Context
		config          ServiceParserConfiguration
		requestId       string
		initialBlockNum int
		args            args
		addr            string
		oldBlockNum     int
		oldTrxNum       int
	}{
		{
			name:    "normal case 1",
			context: context.Background(),
			config: ServiceParserConfiguration{
				MaxAddressNumber:     10,
				MaxConcurrentThreads: 10,
				MaxTransactionNumber: 10,
				Interval:             time.Minute * 100, // for test purpose
			},
			requestId:       generateRequestId(),
			initialBlockNum: 0,
			args: args{
				req: &ethereum.EthGetCurrentTransactionsByAddressRequest{
					FromBlock:   "0x64",
					ToBlock:     "0xc8",
					FromAddress: "0xffff",
					ToAddress:   "0xffff",
				},
				resp: []ethereum.Transaction{
					{
						BlockHash: "0x6464",
					},
					{
						BlockHash: "0xcccc",
					},
				},
				respErr: nil,
			},
			addr:        "0xffff",
			oldBlockNum: 100,
			oldTrxNum:   0,
		},
		{
			name:    "normal case 2 - txn number > maxTransactionNumber",
			context: context.Background(),
			config: ServiceParserConfiguration{
				MaxAddressNumber:     10,
				MaxConcurrentThreads: 10,
				MaxTransactionNumber: 2,
				Interval:             time.Minute * 100, // for test purpose
			},
			requestId:       generateRequestId(),
			initialBlockNum: 0,
			args: args{
				req: &ethereum.EthGetCurrentTransactionsByAddressRequest{
					FromBlock:   "0x64",
					ToBlock:     "0xc8",
					FromAddress: "0xffff",
					ToAddress:   "0xffff",
				},
				resp: []ethereum.Transaction{
					{
						BlockHash: "0x6464",
					},
					{
						BlockHash: "0xcccc",
					},
					{
						BlockHash: "0x1111",
					},
				},
				respErr: nil,
			},
			addr:        "0xffff",
			oldBlockNum: 100,
			oldTrxNum:   0,
		},
		{
			name:    "normal case 3 - existing old trx",
			context: context.Background(),
			config: ServiceParserConfiguration{
				MaxAddressNumber:     10,
				MaxConcurrentThreads: 10,
				MaxTransactionNumber: 3,
				Interval:             time.Minute * 100, // for test purpose
			},
			requestId:       generateRequestId(),
			initialBlockNum: 0,
			args: args{
				req: &ethereum.EthGetCurrentTransactionsByAddressRequest{
					FromBlock:   "0x64",
					ToBlock:     "0xc8",
					FromAddress: "0xffff",
					ToAddress:   "0xffff",
				},
				resp: []ethereum.Transaction{
					{
						BlockHash: "0x6464",
					},
					{
						BlockHash: "0xcccc",
					},
				},
				respErr: nil,
			},
			addr:        "0xffff",
			oldBlockNum: 100,
			oldTrxNum:   2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logger := logging.NewDefaultLogger(logging.LevelDebug)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			chainAccesser := mocks.NewMockEthereumChainAccesser(ctrl)

			// mock getblocknum
			req := &ethereum.EthGetCurrentBlockNumberRequest{RequestId: tt.requestId}
			resp := tt.initialBlockNum
			chainAccesser.EXPECT().EthGetCurrentBlockNumber(tt.context, req).Return(resp, nil)

			// mock get transactions
			chainAccesser.EXPECT().EthGetCurrentTransactionsByAddress(tt.context, tt.args.req).Return(tt.args.resp, tt.args.respErr)

			// Init Parser
			parser := &serviceParser{
				newAddrLock:          sync.Mutex{},
				addrLock:             sync.RWMutex{},
				transactionTasks:     make(chan transactionTask, tt.config.MaxConcurrentThreads),
				interval:             tt.config.Interval,
				maxConcurrentThreads: tt.config.MaxConcurrentThreads,
				maxTransactionNumber: tt.config.MaxTransactionNumber,
				chainAccesser:        chainAccesser,
				logger:               logger,
				addresses:            newAddressTransactionLRU(tt.config.MaxAddressNumber),
			}

			blockNum, err := parser.getBlockNum(tt.context, req)
			assert.Equal(t, tt.initialBlockNum, blockNum)
			assert.Equal(t, nil, err)

			parser.processedBlock = blockNum

			// start Parser
			go parser.start(tt.context)

			oldtrx := make([]ethereum.Transaction, tt.oldTrxNum, parser.maxTransactionNumber)
			for i := 0; i < tt.oldTrxNum; i++ {
				oldtrx[i] = ethereum.Transaction{}
			}

			oldData := addressTransaction{
				address:      tt.addr,
				blockNum:     tt.oldBlockNum,
				transactions: oldtrx,
			}
			parser.addresses.putAddress(oldData)

			parser.doUpdateTransactions(tt.context, tt.args.req)

			got, ok := parser.addresses.dataMap[tt.addr]
			assert.Equal(t, true, ok)
			assert.Equal(t, tt.addr, got.addressTransaction.address)
			assert.Equal(t, tt.args.req.FromBlock, convertDecimalToHex(got.addressTransaction.blockNum))
			assert.Equal(t, minInt(parser.maxTransactionNumber, len(tt.args.resp)+tt.oldTrxNum), len(got.addressTransaction.transactions))
		})
	}
}

func Test_serviceParser_updateAddress(t *testing.T) {
	type args struct {
		context context.Context
	}
	tests := []struct {
		name            string
		context         context.Context
		config          ServiceParserConfiguration
		requestId       string
		initialBlockNum int
		args            args
		newAddress      []string
		oldAddress      []addressTransaction
		oldAddressLogic bool
	}{
		{
			name:    "normal case 1 - less than max",
			context: context.Background(),
			config: ServiceParserConfiguration{
				MaxAddressNumber:     3,
				MaxConcurrentThreads: 10,
				MaxTransactionNumber: 2,
				Interval:             time.Minute * 100, // for test purpose
			},
			requestId:       generateRequestId(),
			initialBlockNum: 0,
			args:            args{},
			newAddress:      []string{"0x1111", "0x1110", "0x1112"},
		},
		{
			name:    "normal case 2 - more than max",
			context: context.Background(),
			config: ServiceParserConfiguration{
				MaxAddressNumber:     3,
				MaxConcurrentThreads: 10,
				MaxTransactionNumber: 2,
				Interval:             time.Minute * 100, // for test purpose
			},
			requestId:       generateRequestId(),
			initialBlockNum: 0,
			args:            args{},
			newAddress:      []string{"0x1111", "0x1110", "0x1112", "0x1113"},
		},
		{
			name:    "normal case 3 - check data",
			context: context.Background(),
			config: ServiceParserConfiguration{
				MaxAddressNumber:     3,
				MaxConcurrentThreads: 10,
				MaxTransactionNumber: 2,
				Interval:             time.Minute * 100, // for test purpose
			},
			requestId:       generateRequestId(),
			initialBlockNum: 0,
			args:            args{},
			newAddress:      []string{"0x1111", "0x1110"},
			oldAddress: []addressTransaction{
				{
					address:      "0x0001",
					transactions: []ethereum.Transaction{},
				},
				{
					address:      "0x0002",
					transactions: []ethereum.Transaction{},
				},
				{
					address:      "0x0003",
					transactions: []ethereum.Transaction{},
				},
			},
			oldAddressLogic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logger := logging.NewDefaultLogger(logging.LevelDebug)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			chainAccesser := mocks.NewMockEthereumChainAccesser(ctrl)

			// mock getblocknum
			req := &ethereum.EthGetCurrentBlockNumberRequest{RequestId: tt.requestId}
			resp := tt.initialBlockNum
			chainAccesser.EXPECT().EthGetCurrentBlockNumber(tt.context, req).Return(resp, nil)

			// Init Parser
			parser := &serviceParser{
				newAddrLock:          sync.Mutex{},
				addrLock:             sync.RWMutex{},
				transactionTasks:     make(chan transactionTask, tt.config.MaxConcurrentThreads),
				interval:             tt.config.Interval,
				maxConcurrentThreads: tt.config.MaxConcurrentThreads,
				maxTransactionNumber: tt.config.MaxTransactionNumber,
				maxAddressNumber:     tt.config.MaxAddressNumber,
				chainAccesser:        chainAccesser,
				logger:               logger,
				addresses:            newAddressTransactionLRU(tt.config.MaxAddressNumber),
			}

			blockNum, _ := parser.getBlockNum(tt.context, req)
			parser.processedBlock = blockNum

			// start Parser
			go parser.start(tt.context)

			for _, d := range tt.oldAddress {
				parser.addresses.putAddress(d)
			}
			parser.newAddresses = tt.newAddress
			parser.updateAddress(tt.args.context)

			assert.Equal(t, minInt(len(tt.newAddress)+parser.maxAddressNumber, parser.maxAddressNumber), parser.addresses.size())

			for i := len(tt.newAddress) - 1; i >= 0 && i < parser.maxAddressNumber; i-- {
				got := parser.addresses.getAddress(tt.newAddress[i])
				assert.Equal(t, tt.newAddress[i], got.address)
			}

			if tt.oldAddressLogic { // specific for case 3
				got := parser.addresses.getAddress(tt.oldAddress[0].address)
				assert.Equal(t, (*addressTransaction)(nil), got)
				got = parser.addresses.getAddress(tt.oldAddress[1].address)
				assert.Equal(t, (*addressTransaction)(nil), got)
				got = parser.addresses.getAddress(tt.oldAddress[2].address)
				assert.Equal(t, tt.oldAddress[2].address, got.address)
				assert.Equal(t, tt.oldAddress[2].address, got.address)
			}
		})
	}
}

func Test_serviceParser_GetTransactions(t *testing.T) {
	type args struct {
		address string
	}

	tests := []struct {
		name            string
		context         context.Context
		config          ServiceParserConfiguration
		requestId       string
		initialBlockNum int
		args            args
		oldData         []addressTransaction
		want            []ethereum.Transaction
	}{
		{
			name:    "normal case 1",
			context: context.Background(),
			config: ServiceParserConfiguration{
				MaxAddressNumber:     3,
				MaxConcurrentThreads: 10,
				MaxTransactionNumber: 2,
				Interval:             time.Minute * 100, // for test purpose
			},
			requestId:       generateRequestId(),
			initialBlockNum: 0,
			args:            args{},
			oldData: []addressTransaction{
				{
					address:      "0x0001",
					transactions: []ethereum.Transaction{},
				},
				{
					address:      "0x0002",
					transactions: []ethereum.Transaction{},
				},
				{
					address:      "0x0003",
					transactions: []ethereum.Transaction{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logger := logging.NewDefaultLogger(logging.LevelDebug)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			chainAccesser := mocks.NewMockEthereumChainAccesser(ctrl)

			// mock getblocknum
			req := &ethereum.EthGetCurrentBlockNumberRequest{RequestId: tt.requestId}
			resp := tt.initialBlockNum
			chainAccesser.EXPECT().EthGetCurrentBlockNumber(tt.context, req).Return(resp, nil)

			// Init Parser
			parser := &serviceParser{
				newAddrLock:          sync.Mutex{},
				addrLock:             sync.RWMutex{},
				transactionTasks:     make(chan transactionTask, tt.config.MaxConcurrentThreads),
				interval:             tt.config.Interval,
				maxConcurrentThreads: tt.config.MaxConcurrentThreads,
				maxTransactionNumber: tt.config.MaxTransactionNumber,
				maxAddressNumber:     tt.config.MaxAddressNumber,
				chainAccesser:        chainAccesser,
				logger:               logger,
				addresses:            newAddressTransactionLRU(tt.config.MaxAddressNumber),
			}

			blockNum, _ := parser.getBlockNum(tt.context, req)
			parser.processedBlock = blockNum

			// start Parser
			go parser.start(tt.context)

			for _, d := range tt.oldData {
				parser.addresses.putAddress(d)
			}

			for _, d := range tt.oldData {
				got := parser.GetTransactions(d.address)
				assert.Equal(t, d.transactions, got)
			}

			got := parser.GetTransactions("0xfffnotexisti")
			assert.Equal(t, []ethereum.Transaction{}, got)
		})
	}
}
