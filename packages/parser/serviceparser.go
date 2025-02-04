package parser

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/brofu/simple_ethereum_parser/packages/ethereum"
	"github.com/brofu/simple_ethereum_parser/packages/logging"
)

type addressTransaction struct {
	address      string
	blockNum     int
	transactions []ethereum.Transaction
}

type transactionTask struct {
	blockNum int
	address  string
}

type ServiceParserConfiguration struct {
	MaxAddressNumber            int
	MaxTransactionNumber        int
	MaxConcurrentThreads        int
	Interval                    time.Duration
	GetBlockNumberQueryTimeout  time.Duration
	GetTransactionsQueryTimeout time.Duration
}

// serviceParser implements the `Parser` interface
type serviceParser struct {
	// mark if there is going on transactions task
	processing bool

	// processed block, when the instance is new started, this mean the started block number
	processedBlock int

	// Interval to check if there is new block
	interval time.Duration
	// max number of concurrent worker to get transactions
	maxConcurrentThreads int
	// max number of transactions of an address would be stored in storage
	maxTransactionNumber int
	// max number of addresses
	maxAddressNumber int

	// Lock for new subscribed address list update
	newAddrLock sync.Mutex
	// Lock used for address list update
	addrLock sync.RWMutex

	addresses *addressTransactionLRU

	transactionTasks chan transactionTask
	// Used to notify there is new task of `get of transactions`. Sent from `task distributor` to `task executor`
	newTaskNoti chan int
	// Used to notify there is ONE task finished. Sent from `task executor workers` to `task executor`
	finishedTasks chan struct{}

	// new subscribed addresses to handle
	newAddresses []string

	//ethereum chain accesser
	chainAccesser ethereum.EthereumChainAccesser

	// logger
	// may consider to allow the caller to set this.
	logger logging.Logger

	// timeout when can chain
	getBlockNumTimeOut          time.Duration
	getTransactionsQueryTimeout time.Duration
}

// NewServiceParser construct an instance of `serviceParser`
func NewServiceParser(ctx context.Context, logger logging.Logger, chainAccesser ethereum.EthereumChainAccesser, config ServiceParserConfiguration) Parser {

	parser := &serviceParser{
		newAddrLock:                 sync.Mutex{},
		addrLock:                    sync.RWMutex{},
		transactionTasks:            make(chan transactionTask),
		newTaskNoti:                 make(chan int),
		finishedTasks:               make(chan struct{}),
		interval:                    config.Interval,
		maxConcurrentThreads:        config.MaxConcurrentThreads,
		maxTransactionNumber:        config.MaxTransactionNumber,
		maxAddressNumber:            config.MaxAddressNumber,
		chainAccesser:               chainAccesser,
		logger:                      logger,
		addresses:                   newAddressTransactionLRU(config.MaxAddressNumber),
		getBlockNumTimeOut:          config.GetBlockNumberQueryTimeout,
		getTransactionsQueryTimeout: config.GetTransactionsQueryTimeout,
	}

	req := &ethereum.EthGetCurrentBlockNumberRequest{
		RequestId: generateRequestId(),
	}

	blockNum, err := parser.getBlockNum(ctx, req)
	if err != nil { // good practice to fast fail.
		parser.logger.Errorf("get init block number fail | error: %s", err.Error())
		panic(err)
	}
	parser.processedBlock = blockNum
	go parser.start(ctx)
	return parser
}

func (this *serviceParser) GetCurrentBlock() int {
	return this.processedBlock
}

func (this *serviceParser) Subscribe(address string) bool {
	this.newAddrLock.Lock()
	defer this.newAddrLock.Unlock()
	this.newAddresses = append(this.newAddresses, address)
	return true
}

func (this *serviceParser) GetTransactions(address string) []ethereum.Transaction {

	this.addrLock.RLock()
	data := this.addresses.getAddress(address)
	this.addrLock.RUnlock()

	if data == nil {
		return []ethereum.Transaction{}
	} else {
		return data.transactions
	}
}

func (this *serviceParser) start(ctx context.Context) {
	go this.startTaskDistribution(ctx) // start task distribution
	go this.startTaskExecution(ctx)    // start task execution
}

//startTaskDistribution is the controller of distributing tasks (to get transaction from new block)
func (this *serviceParser) startTaskDistribution(ctx context.Context) {

	this.logger.Infof("task distributor started")
	ticker := time.Tick(this.interval)

	for {
		select {
		case <-ctx.Done():
			this.logger.Infof("task distributor existing")
			return
		case <-ticker:
			req := &ethereum.EthGetCurrentBlockNumberRequest{
				RequestId: generateRequestId(),
			}

			blockNum, err := this.getBlockNum(ctx, req)
			if err != nil { //if there is err, just skip this round.
				this.logger.Errorf("get init block number with timer fail | error: %s", err.Error())
				continue
			}

			if blockNum == this.processedBlock { // NO new block
				this.logger.Infof("no new block | processedBlock: %d, new blockNum: %d", this.processedBlock, blockNum)
				continue
			}

			if this.processing { // there is ongoing tasks, wait it finished, and do nothing for now.
				this.logger.Infof("processing, skip this round | processedBlock: %d, new blockNum: %d", this.processedBlock, blockNum)
				continue
			}

			// No ongoing tasks, kick off the work.
			this.logger.Infof("kick up a new round of task | processedBlock: %d, new blockNum: %d", this.processedBlock, blockNum)
			this.processedBlock = blockNum
			this.updateAddress(ctx)
			if this.addresses.size() == 0 { // edged case: the timer is trigger before there is any address
				continue
			}
			this.newTaskNoti <- this.addresses.size()
			this.distributeTasks(blockNum)
		}
	}
}

// startTaskExecution is the controller of task execution
func (this *serviceParser) startTaskExecution(ctx context.Context) {

	this.logger.Infof("task executor controller starting")

	// spawn MaxConcurrentThreads of workers
	for i := 0; i < this.maxConcurrentThreads; i++ {
		go this.executeTasks(ctx, i)
	}

	for {
		select {
		case <-ctx.Done():
			this.logger.Infof("task executor controller existing")
			return
		case taskNum := <-this.newTaskNoti:
			this.logger.Infof("getting new tasks | number: %d", taskNum)
			this.processing = true
			// monitor the finished tasks
			finished := 0
			for range this.finishedTasks {
				finished += 1
				if finished == taskNum { // record the finished workers
					break
				}
				this.logger.Debugf("controller. finished tasks number: %d, total: %d", finished, taskNum)
			}
			this.logger.Infof("controller. finished tasks number: %d, total: %d", finished, taskNum)
			this.processing = false
		}
	}
}

// updateAddress pick up the new subscribed addressed and insert them into the storage (`this.addresses`)
func (this *serviceParser) updateAddress(ctx context.Context) {

	// get new addresses
	this.newAddrLock.Lock()
	newAddresses := this.newAddresses
	this.newAddresses = []string{}
	this.newAddrLock.Unlock()

	if len(newAddresses) == 0 {
		return
	}

	this.logger.Infof("%d addresses new added: %s", len(newAddresses), newAddresses)

	this.addrLock.Lock()
	defer this.addrLock.Unlock()
	// add the new addresses, this would stop all the API queries
	for _, addr := range newAddresses {
		this.addresses.putAddress(addressTransaction{
			address:      addr,
			blockNum:     this.processedBlock,
			transactions: make([]ethereum.Transaction, 0, this.maxTransactionNumber),
		})
	}
}

// executeTasks is the real worker to execute tasks.
// Need to make sure to notify the `controller` no matter the task is successful or not
func (this *serviceParser) executeTasks(ctx context.Context, workerNum int) {

	this.logger.Infof("worker started | worker number: %d", workerNum)
	for {
		select {
		case task := <-this.transactionTasks:
			this.updateTransactions(ctx, task.address, task.blockNum)
			this.logger.Infof("finished task | worker: %d, address: %s", workerNum, task.address)
			this.finishedTasks <- struct{}{}
		case <-ctx.Done():
			this.logger.Infof("worker existing | worker number: %d", workerNum)
			return
		}
	}
}

// distributeTasks distribute the task (to get transactions of new block) to the queue.
func (this *serviceParser) distributeTasks(newBlockNum int) {
	addresses := this.addresses.allAddresses()
	this.logger.Debugf("existing addresses %v", addresses)
	for _, addr := range addresses {
		this.logger.Infof("distribute task | address: %s", addr)
		this.transactionTasks <- transactionTask{newBlockNum, addr}
	}
}

// updateTransactions update the transaction of an address
func (this *serviceParser) updateTransactions(ctx context.Context, addr string, blockNum int) {

	req := this.constructGetTransactionRequest(addr, blockNum)
	if req == nil {
		this.logger.Errorf("construct get transaction request fail | address: %s", addr)
	}

	this.doUpdateTransactions(ctx, req)
}

// constructGetTransactionRequest construct the request of get transaction
func (this *serviceParser) constructGetTransactionRequest(addr string, blockNum int) *ethereum.EthGetCurrentTransactionsByAddressRequest {
	addrData := this.addresses.getAddressIn(addr)
	if addrData == nil { // this should not happen
		return nil
	}

	// update
	startBn := convertDecimalToHex(addrData.blockNum)
	endBn := convertDecimalToHex(blockNum)
	req := &ethereum.EthGetCurrentTransactionsByAddressRequest{
		FromBlock:   startBn,
		ToBlock:     endBn,
		FromAddress: addr,
		ToAddress:   addr,
		RequestId:   fmt.Sprintf("%d", time.Now().UnixNano()),
	}
	return req
}

func (this *serviceParser) doUpdateTransactions(ctx context.Context, req *ethereum.EthGetCurrentTransactionsByAddressRequest) {

	addrData := this.addresses.getAddressIn(req.FromAddress)
	if addrData == nil { // this should not happen
		this.logger.Errorf("get address from storage fail | address: %s", req.FromAddress)
		return
	}

	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(this.getTransactionsQueryTimeout))
	defer cancelFunc()
	resp, err := this.chainAccesser.EthGetCurrentTransactionsByAddress(ctx, req)
	if err != nil {
		this.logger.Errorf("call ethereum chain to get Transactions fail | req: %v, error: %s", req, err.Error())
		return
	}

	var newTrx []ethereum.Transaction
	newTrxNum := len(resp)

	if newTrxNum >= this.maxTransactionNumber {
		newTrx = resp[:this.maxTransactionNumber]
	} else {
		newTrx = resp
		space := this.maxTransactionNumber - len(newTrx)
		if len(addrData.transactions) <= space {
			newTrx = append(newTrx, addrData.transactions...)
		} else {
			newTrx = append(newTrx, addrData.transactions[:space]...)
		}
	}
	addrData.transactions = newTrx

	this.logger.Infof("update transactions success | address: %s, new trx number: %d, total: %d",
		req.FromAddress, minInt(newTrxNum, this.maxTransactionNumber), len(addrData.transactions))
}

func (this *serviceParser) getBlockNum(ctx context.Context, req *ethereum.EthGetCurrentBlockNumberRequest) (int, error) {
	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(this.getBlockNumTimeOut))
	defer cancelFunc()
	bn, err := this.chainAccesser.EthGetCurrentBlockNumber(ctx, req)
	if err != nil {
		return 0, err
	}
	return bn, nil
}

func convertDecimalToHex(num int) string {
	return fmt.Sprintf("0x%s", fmt.Sprintf("%x", num))
}

//generateRequestId generate a new request ID
// TODO: there are better solutions for this. but need more efforts
func generateRequestId() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
