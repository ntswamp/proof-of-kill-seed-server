package main

import (
	"app/src/constant"
	lib_chatwork "app/src/lib/chatwork"
	lib_db "app/src/lib/db"
	lib_log "app/src/lib/log"
	lib_redis "app/src/lib/redis"
	"app/src/model"
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	Logger           = lib_log.NewLogger("cmd/eth/event_watcher.go")
	FailureTolerance = 0 //currently zero fault tols
)

func main() {

	Logger.Debug("started watching on eth blocks at %v", time.Now())

	/*========================*/
	//db & redis
	redisCache, err := lib_redis.NewClient(constant.RedisCache)
	if err != nil {
		lib_chatwork.LogErrorAndAlert(Logger, "failed connecting redis: %v", err)
		log.Fatal(err)
	}
	defer redisCache.Terminate()
	dbClient, err := lib_db.Connect(constant.DbIdolverse, redisCache)
	if err != nil {
		lib_chatwork.LogErrorAndAlert(Logger, "failed connecting db: %v", err)
		log.Fatal(err)
	}
	defer dbClient.Close()

	//set up eth client
	ethClient, err := ethclient.Dial(constant.INFURA_TESTNET)
	if err != nil {
		lib_chatwork.LogErrorAndAlert(Logger, "eth client err: %v", err)
		log.Fatal(err)
	}

	/*========================*/
	//catch up on blocks and start watching
	lastCheckedBlockNumber, err := getLastCheckedBlockNumber(dbClient)
	if err != nil {
		lib_chatwork.LogErrorAndAlert(Logger, "failed getting the last checked block number: %v", err)
		log.Fatal(err)
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(constant.OPENSEA_SALE_CONTRACT_TESTNET)},
		FromBlock: new(big.Int).SetUint64(lastCheckedBlockNumber + 1),
		ToBlock:   nil,                                                                    //latest block at the time
		Topics:    [][]common.Hash{{common.HexToHash(constant.ETH_TOKEN_TRANSFER_TOPIC)}}, //only listen to transfer topics
	}

	fmt.Printf("start watching from block:%d\n", lastCheckedBlockNumber+1)
	fmt.Printf("args:%+v\n", query)

	incomingLogs := make(chan types.Log)
	sub, err := ethClient.SubscribeFilterLogs(context.Background(), query, incomingLogs)
	if err != nil {
		lib_chatwork.LogErrorAndAlert(Logger, "failed to subscribe logs with args: %+v", query)
		log.Fatal(err)
	}

	logs, err := ethClient.FilterLogs(context.Background(), query)
	if err != nil {
		lib_chatwork.LogErrorAndAlert(Logger, "eth client err: %v", err)
		log.Fatal(err)
	}

	//catch up
	for _, Log := range logs {
		err := processLog(dbClient, &Log)
		if err != nil {
			lib_chatwork.LogErrorAndAlert(Logger, "failed processing a event log.\ntx hash:%s\nblock hash:%s\ncontract:%s\nerror: %v", Log.TxHash.String(), Log.BlockHash.String(), Log.Address.String(), err)
			FailureTolerance--
			if FailureTolerance < 0 {
				lib_chatwork.LogErrorAndAlert(Logger, "too many failures upon log processing. check out your system and restart the service.")
				log.Fatal(err)
			}
		}
	}

	/*========================*/

	//start watching
	for {
		select {
		case err := <-sub.Err():
			lib_chatwork.LogErrorAndAlert(Logger, "failed fetching event logs:%v", err)
			log.Fatal(err)
		case incomingLog := <-incomingLogs:
			err := processLog(dbClient, &incomingLog)
			if err != nil {
				lib_chatwork.LogErrorAndAlert(Logger, "failed processing a event log.\ntx hash:%s\nblock hash:%s\ncontract:%s\nerror: %v", incomingLog.TxHash.String(), incomingLog.BlockHash.String(), incomingLog.Address.String(), err)
				FailureTolerance--
				if FailureTolerance < 0 {
					lib_chatwork.LogErrorAndAlert(Logger, "too many failures upon log processing. check on your system and restart the service.")
					log.Fatal(err)
				}
			}
		}
	}
}

func getLastCheckedBlockNumber(dbClient *lib_db.Client) (uint64, error) {
	record := &model.BlockchainWatcher{
		CurrencyType: constant.ETH,
	}
	db := dbClient.GetDB().Find(record)
	if db.Error != nil {
		return 0, db.Error
	}
	return uint64(record.LastCheckedBlockNumber), nil
}

func processLog(dbClient *lib_db.Client, log *types.Log) error {
	//prepare db tx
	dbClient.StartTransaction()
	defer dbClient.RollbackTransaction()
	modelMgr := dbClient.GetModelManager()

	contractAddress := strings.ToLower(log.Address.String())
	fmt.Printf("address:%s\n", contractAddress)
	switch contractAddress {
	case constant.OPENSEA_SALE_CONTRACT_TESTNET:
		fmt.Println("processing a sale from opensea")
		//save sale tx to db
		saleRecord := &model.PuzzleSaleTransaction{
			BuyerEthAccount: log.Topics[2].String(),
			PuzzleTokenId:   log.Topics[3].String(),
			TxHash:          log.TxHash.String(),
			CreatedAt:       time.Now(),
		}

		err := modelMgr.CachedSave(saleRecord, nil)
		if err != nil {
			return err
		}

	default:
		//do nothing
		fmt.Printf("contract unintelligible: %s\n", contractAddress)
		return nil
	}

	//update LastCheckedBlockNumber after the processing
	if log.BlockNumber > math.MaxInt64 {
		return fmt.Errorf("block number overflow:%d", log.BlockNumber)
	}

	bwRecord := &model.BlockchainWatcher{}
	_, err := modelMgr.GetModel(bwRecord, constant.ETH)
	if err != nil {
		return err
	}
	bwRecord.CurrencyType = constant.ETH
	bwRecord.LastCheckedBlockNumber = int64(log.BlockNumber)
	err = modelMgr.CachedSave(bwRecord, nil)
	if err != nil {
		return err
	}

	// Write and commit all transactions in the modelmanager
	err = modelMgr.WriteAll()
	if err != nil {
		return err
	}
	err = dbClient.CommitTransaction()
	if err != nil {
		return err
	}

	return nil
}

func getMostRecentBlockNumber(client ethclient.Client) error {
	//get the most recent block number

	latestBlockNumberOnChain, err := client.BlockNumber(context.Background())
	if latestBlockNumberOnChain > math.MaxInt64 {
		err = fmt.Errorf("block number overflow: %v", latestBlockNumberOnChain)
	}
	return err
}
