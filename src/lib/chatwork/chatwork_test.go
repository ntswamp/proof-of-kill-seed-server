package lib_chatwork

import (
	"app/src/constant"
	lib_log "app/src/lib/log"
	"fmt"
	"log"
	"testing"
)

func TestReturnErrorAndAlert(t *testing.T) {
	e := fmt.Errorf("hellyeah")
	ReturnErrorAndAlert(e, "TestReturnErrorAndAlert")
}

func TestLogErrorAndAlert(t *testing.T) {

	Logger := lib_log.NewLogger("cmd/eth/event_watcher.go")

	LogErrorAndAlert(Logger, `start watching from block:9467791
	args:{BlockHash:<nil> FromBlock:+9467791 ToBlock:<nil> Addresses:[0x526Db928aa631CD4E92aD11338370e98B07c6de7] Topics:[[0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef]]}
	address:0x526db928aa631cd4e92ad11338370e98b07c6de7
	processing a sale from opensea
	
	(/home/nts/go/web-sudoku-server/src/lib/db/callback_update.go:83) 
	[2021-10-15 14:39:41]  pq: syntax error at or near "=" %d`, 2009)
}

func TestPostNewRoomMessage(t *testing.T) {
	_, err := PostNewRoomMessage("test", constant.CHATWORK_ALERT_ROOMID, constant.CHATWORK_APIKEY)
	if err != nil {
		log.Fatal(err)
	}
}
