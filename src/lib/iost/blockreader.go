package lib_iost

import (
	"context"
	rpcpb "github.com/iost-official/go-iost/rpc/pb"
	"google.golang.org/grpc"
	"time"
)

type BlockReader struct {
	blocks map[int64]*rpcpb.BlockResponse

	HeadBlock        int64
	currentBlock     int64
	currentSendBlock int64
	endBlock         int64
	timeout          time.Duration
	conn             *grpc.ClientConn

	headBlockTick *time.Ticker
	newBlockTick  *time.Ticker

	quit              chan interface{}
	blockChannel      chan *rpcpb.BlockResponse
	blockRetryRequest chan int64
	blockOut          chan *rpcpb.BlockResponse

	stopped  bool
	errs     chan error
	errCount int
}

const MAX_ERRORS int = 5

// Takes node address, connects to the node and returns BlockReader
func NewBlockReader(addr string) (*BlockReader, error) {
	br := new(BlockReader)
	err := br.connectToNode(addr)
	return br, err
}

// terminate blockreader, cleaning up tickers,channels & connection to node
func (self *BlockReader) terminate() {
	if !self.stopped {
		if self.quit != nil {
			close(self.quit)
		}
		if self.headBlockTick != nil {
			self.headBlockTick.Stop()
		}
		if self.newBlockTick != nil {
			self.newBlockTick.Stop()
		}
		if self.conn != nil {
			self.conn.Close()
		}
		self.stopped = true
	}
}

// Calls terminate method, shutting down BlockReader
func (self *BlockReader) End() {
	self.terminate()
}

func (self *BlockReader) Start(done chan []error) {
	errs := []error{}
L:
	for {
		select {
		case <-self.quit:
			// Break loop if quit was closed
			break L
		case err := <-self.errs:
			// Receive errors
			errs = append(errs, err)
			self.errCount++
			if self.errCount >= MAX_ERRORS {
				self.terminate()
			}
		case <-self.newBlockTick.C:
			if self.currentBlock <= self.HeadBlock {
				// Get block from node
				go self.getBlockFromNode(self.currentBlock)
				self.currentBlock++
			}
			// Send available block
			if block, ok := self.blocks[self.currentSendBlock]; ok {
				self.blockOut <- block
				delete(self.blocks, self.currentSendBlock)
				self.currentSendBlock++
			}
		case <-self.headBlockTick.C:
			// Periodically update head block unless we have a endblock set
			if self.endBlock == 0 {
				behind := self.HeadBlock - self.currentBlock
				// No point in updating head block if we are really far behind
				if behind < 100 {
					lib, err := self.GetLatestIrreversibleBlock()
					if err != nil {
						errs = append(errs, err)
						self.terminate()
					}
					self.HeadBlock = lib
				}
			}
		case block := <-self.blockChannel:
			// Add incoming block to block map
			self.blocks[block.Block.Number] = block
		case blockNum := <-self.blockRetryRequest:
			// Get block from node
			go self.getBlockFromNode(blockNum)
		}
	}
	// Send errors back
	done <- errs
}

func (self *BlockReader) RequestBlockByNum(blockNum int64) {
	go self.getBlockFromNode(blockNum)
}

// Retrieves a block from the iost node
func (self *BlockReader) getBlockFromNode(blockNum int64) {
	resp := new(rpcpb.BlockResponse)
	req := &rpcpb.GetBlockByNumberRequest{Number: blockNum, Complete: true}
	ctx := context.Background()
	err := self.conn.Invoke(ctx, "/rpcpb.ApiService/GetBlockByNumber", req, resp)
	if err != nil {
		self.errs <- err
		self.blockRetryRequest <- blockNum
		return
	}
	self.blockChannel <- resp
}

// If startBlock is 0 it will start reading from the latest irreversible block.
// If endBlock is 0 it will continously read from the blockchain.
func (self *BlockReader) Setup(startBlock, endBlock int64) (chan *rpcpb.BlockResponse, error) {
	// Blocks map
	self.blocks = make(map[int64]*rpcpb.BlockResponse)
	// Init channels
	self.blockChannel = make(chan *rpcpb.BlockResponse)
	self.blockOut = make(chan *rpcpb.BlockResponse)
	self.blockRetryRequest = make(chan int64)
	self.quit = make(chan interface{})
	self.errs = make(chan error)

	// Settings
	self.endBlock = endBlock

	// Init tickers
	self.headBlockTick = time.NewTicker(time.Second)
	self.newBlockTick = time.NewTicker(10 * time.Millisecond)

	// Get current LIB and set block number related vars
	headBlock, err := self.GetLatestIrreversibleBlock()
	if err != nil {
		return nil, err
	}
	if self.endBlock != 0 {
		self.HeadBlock = self.endBlock
	} else {
		self.HeadBlock = headBlock
	}

	if startBlock != 0 {
		self.currentBlock = startBlock
	} else {
		self.currentBlock = headBlock
	}
	self.currentSendBlock = self.currentBlock

	return self.blockOut, nil
}

func (self *BlockReader) connectToNode(addr string) error {
	var err error
	self.conn, err = grpc.Dial(addr, grpc.WithInsecure())
	return err
}

func (self *BlockReader) GetLatestIrreversibleBlock() (int64, error) {
	resp := new(rpcpb.ChainInfoResponse)
	err := self.conn.Invoke(context.Background(), "/rpcpb.ApiService/GetChainInfo", &rpcpb.EmptyRequest{}, resp)
	if err != nil {
		return 0, err
	}
	return resp.LibBlock, nil
}
