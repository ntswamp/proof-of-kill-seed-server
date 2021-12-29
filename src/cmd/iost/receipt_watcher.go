package main

import (
	"app/src/constant"
	lib_db "app/src/lib/db"
	lib_error "app/src/lib/error"
	lib_iost "app/src/lib/iost"
	lib_log "app/src/lib/log"
	lib_redis "app/src/lib/redis"
	lib_sig "app/src/lib/signature"
	"app/src/model"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	rpcpb "github.com/iost-official/go-iost/rpc/pb"
)

var (
	logger = lib_log.NewLogger("cmd/iost/receipt_watcher.go")
)

func main() {
	// 開始時間.
	logger.Debug("started watching iost blocks at %v", time.Now())
	//sendChatWorkServiceMsg(fmt.Sprintf("[%s] Starting TL-IOST BlockReader Service", constant.GetServerEnv()))

	// Get Redis Client
	redisCache, err := lib_redis.NewClient(constant.RedisCache)
	if err != nil {
		logger.Error("failed to talk to redis:%v", err)
		return
	}
	defer redisCache.Terminate()
	// DB接続.
	dbClient, err := lib_db.Connect(constant.DbIdolverse, redisCache)
	if err != nil {
		logger.Error("failed to connect to db:%v", err)
		return
	}
	defer dbClient.Close()

	err = BeginBlockReaderProcess(dbClient)
	if err != nil {
		logger.Error("failed processing blocks: %v", err)
		//sendChatWorkServiceMsg(fmt.Sprintf("[toall]\n[%s] Stopped TL-IOST BlockReader Service with errors:\n%v", config.GetServerEnvName(), err))
		return
	}
	//sendChatWorkServiceMsg(fmt.Sprintf("[toall]\n[%s] Stopped TL-IOST BlockReader Service gracefully", config.GetServerEnvName()))
	logger.Debug("iost block reader stopped at %v", time.Now())
}

/*
func sendChatWorkServiceMsg(msg string) error {
	_, err := lib_chatwork.PostNewRoomMessage(msg, app_defines_tokenlink.ChatworkServicesRoomID, app_defines_tokenlink.ChatworkAPIToken)
	return err
}
*/
func BeginBlockReaderProcess(dbClient *lib_db.Client) error {
	// Init Block Reader
	addr := constant.IOST_TESTNET
	if constant.IsProduction() {
		addr = constant.IOST_MAINNET
	}
	blockReader, err := lib_iost.NewBlockReader(addr)
	if err != nil {
		return lib_error.WrapError(err)
	}

	// Determine start block to read from
	lib, err := blockReader.GetLatestIrreversibleBlock()
	if err != nil {
		return lib_error.WrapError(err)
	}
	//start from lib + 1 or lib if no initial record found
	startBlock, err := GetStartBlock(lib, dbClient)
	if err != nil {
		return lib_error.WrapError(err)
	}

	// Setup block reader
	incomingBlocks, err := blockReader.Setup(startBlock, 0)
	if err != nil {
		return lib_error.WrapError(err)
	}

	// quit channel for exiting loop
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start block reader
	done := make(chan []error)
	go blockReader.Start(done)

	for {
		select {
		case <-quit:
			blockReader.End()
		case errs := <-done:
			if len(errs) > 0 {
				errStr := ""
				for brErrNum, brErr := range errs {
					errStr += fmt.Sprintf("#%d:%v\n", brErrNum, brErr)
					logger.Error("Blockreader Err#%d:%v", brErrNum, brErr)
				}
				return fmt.Errorf("BlockReader exited with %d errors:\n%s", len(errs), errStr)
			}
			return nil
		case blockRes := <-incomingBlocks:
			processBlockErr := ProcessBlock(dbClient, blockRes)
			if processBlockErr != nil {
				logger.Error("ProcessBlock Err:%v", processBlockErr)
				blockReader.End()
			}
		}
	}
}

//if no initial record found, start watching from current LIB
func GetStartBlock(lib int64, dbClient *lib_db.Client) (int64, error) {
	record := &model.BlockchainWatcher{
		CurrencyType: constant.IOST,
	}
	db := dbClient.GetDB().Find(record)
	if db.RecordNotFound() {
		err := func() error {
			dbtx := dbClient.StartTransaction()
			defer dbClient.RollbackTransaction()
			record.LastCheckedBlockNumber = lib
			err := dbtx.Create(record).Error
			if err != nil {
				logger.Error("failed to create record:%v", err)
				return err
			}
			err = dbClient.CommitTransaction()
			if err != nil {
				logger.Error("failed to commit on record:%v", err)
				return err
			}
			return nil
		}()
		if err != nil {
			return 0, err
		}
		return lib, nil
	} else if db.Error != nil {
		return 0, db.Error
	}

	return record.LastCheckedBlockNumber + 1, nil
}

func ProcessBlock(dbClient *lib_db.Client, incomingBlockRes *rpcpb.BlockResponse) error {
	var err error
	// Begin DB transaction
	dbtx := dbClient.StartTransaction()
	defer dbClient.RollbackTransaction()

	// Grab the latest proccessed iost block and lock the record
	record := &model.BlockchainWatcher{
		CurrencyType: constant.IOST,
	}
	scope := dbtx.Set("gorm:query_option", "FOR UPDATE").Find(record)
	if scope.Error != nil {
		logger.Error("failed to get iost_watching_block record:%v", scope.Error)
		return scope.Error
	}
	// Retrieve block for block reader
	expectedBlockNumber := record.LastCheckedBlockNumber + 1
	if expectedBlockNumber != incomingBlockRes.Block.Number {
		return fmt.Errorf("received unexpected block number from block reader")
	}

	if incomingBlockRes.Status != rpcpb.BlockResponse_IRREVERSIBLE {
		logger.Error("block:%v with status:%s", incomingBlockRes.Block.Number, incomingBlockRes.Status.String())
		return fmt.Errorf("block is not irreversible")
	}

	// Loop through all the transactions of the block
	linkToEthContractId := constant.IOST_CONTRACT_LINK_TO_ETH_TESTNET
	otherContractId := "any other contracts we want to look into"

	for _, tx := range incomingBlockRes.Block.Transactions {
		if tx.TxReceipt.StatusCode != rpcpb.TxReceipt_SUCCESS {
			// Ingore unsucessful transactions
			continue
		}
		// Inspect the receipts
		for _, receipt := range tx.TxReceipt.Receipts {
			// Break the receipt down into parts
			arr := strings.Split(receipt.FuncName, "/")
			contractId := arr[0]
			funcName := arr[1]
			// Get transaction publish time
			txTime := tx.GetTime()
			publishedTime := time.Unix(0, txTime)
			// Get transaction publisher
			txPublisher := tx.GetPublisher()
			// Check if the contract id matches one of our expected ids
			switch contractId {
			case linkToEthContractId:
				// Process market receipt
				err = processLinkToEth(dbClient, expectedBlockNumber, funcName, tx.TxReceipt.TxHash, txPublisher, receipt.Content, publishedTime)

			case otherContractId:
				// process other receipt
				err = processOtherContract(dbClient, expectedBlockNumber, funcName, tx.TxReceipt.TxHash, txPublisher, receipt.Content, publishedTime)
			}

			if err != nil {
				// Something went wrong, exit out after logging the error
				if iostError, ok := err.(*lib_error.SaleError); ok {
					logger.Error("contract error(%v) in func:%s hash:%s receipt:%s err:%s", iostError.ErrorType, iostError.FuncName, tx.TxReceipt.TxHash, receipt.Content, iostError.Message)
				} else {
					logger.Error("contract error:%v hash:%v receipt:%s", err.Error(), tx.TxReceipt.TxHash, receipt.Content)
				}
				return err
			}
		}
	}

	// Everything went fine, update the BlockNumber to be read.
	err = dbtx.Model(record).Where("currency_type = ?", record.CurrencyType).Update("last_checked_block_number", expectedBlockNumber).Error
	if err != nil {
		logger.Error("failed updating record:%v", err)
		return err
	}
	// Write and commit all transactions in the modelmanager
	err = dbClient.GetModelManager().WriteAll()
	if err != nil {
		logger.Error("modelMgrの保存でエラー:%v", err)
		return err
	}
	err = dbClient.CommitTransaction()
	if err != nil {
		logger.Error("inhaleOneBlockでコミットエラー:%v", err)
		return err
	}

	return nil
}

// Process LinkToEth Contract
func processLinkToEth(dbClient *lib_db.Client, BlockNumber int64, funcName, txHash, txPublisher, receiptContent string, txTime time.Time) error {

	modelMgr := dbClient.GetModelManager()

	receipt := []string{}
	err := json.Unmarshal([]byte(receiptContent), &receipt)
	if err != nil {
		logger.Error("failed to unmarshal receipt content: %v", err)
		return err
	}
	iostAccount := receipt[0]
	ethAccount := receipt[1]
	signature := receipt[2]

	validSig := lib_sig.IsLegalSig(ethAccount, signature, []byte(iostAccount))
	if !validSig {
		return fmt.Errorf("bad signature")
	}

	switch funcName {
	case "apply_ethereum_address":

		record := &model.Account{}
		err = dbClient.GetDB().Table("user_account").Where("`iost_account` = ?", iostAccount).Find(record).Error
		if err != nil {
			logger.Error("iost account not found: %v", err)
			return err
		}

		//record.EthAccount = ethAccount
		saveOpt := &lib_db.SaveOptions{
			Fields: []string{"EthAccount"},
		}
		err = modelMgr.CachedSave(record, saveOpt)
		if err != nil {
			logger.Error("failed updating user eth account:%v", err)
			return err
		}

		//call register abi
		var AccName string
		var SecKey string
		if constant.IsDebug() {
			AccName = constant.ADMIN_IOST_ACCOUNT_TESTNET
			SecKey = constant.ADMIN_IOST_ACCOUNT_TESTNET_PRIVATE_KEY
		} else {
			AccName = constant.ADMIN_IOST_ACCOUNT
			SecKey = constant.ADMIN_IOST_ACCOUNT_PRIVATE_KEY
		}
		sdkOpt := &lib_iost.IOSTDevSDKOptions{
			AccName: AccName,
			SecKey:  SecKey,
		}
		sdk, err := lib_iost.NewIOSTDevSDKWithOptions(sdkOpt)
		if err != nil {
			logger.Error("failed creating iost sdk:%v", err)
			return err
		}
		txhash, err := lib_iost.CallAbi(sdk, constant.IOST_CONTRACT_RIGISTER_ETH_ADDR_TESTNET, "regist_ethereum_address", iostAccount, ethAccount)
		if err != nil {
			logger.Error("failed calling abi: %v", err)
			return err
		}

		if txhash != "" {
			logger.Debug("eth account %s registered to iost account %s", ethAccount, iostAccount)
		}
		return err

	default:
		// dont process other functions
		return nil
	}
}

func processOtherContract(dbClientSale *lib_db.Client, BlockNumber int64, funcName, txHash, txPublisher, receiptContent string, txTime time.Time) error {

	return nil
}
