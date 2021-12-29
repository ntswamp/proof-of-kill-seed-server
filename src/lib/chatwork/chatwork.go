package lib_chatwork

import (
	"app/src/constant"
	lib_log "app/src/lib/log"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const APIUrl string = "https://api.chatwork.com/v2/"
const TokenHeader string = "X-ChatWorkToken"

func postRequest(endpoint, token string, data []byte) ([]byte, error) {
	if token == "" {
		return nil, fmt.Errorf("Token must not be empty")
	}
	url := fmt.Sprintf("%s%s", APIUrl, endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	// Add headers
	req.Header.Add(TokenHeader, token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Http client
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respData, nil
}

func PostNewRoomMessage(roomId, token, msg string) ([]byte, error) {
	if roomId == "" {
		return nil, fmt.Errorf("empty room id")
	}
	data := url.Values{}
	data.Set("body", msg)
	endpoint := fmt.Sprintf("rooms/%s/messages", roomId)
	return postRequest(endpoint, token, []byte(data.Encode()))
}

func ReturnErrorAndAlert(err error, msg string, args ...interface{}) error {
	msg = fmt.Sprintf(msg, args...)
	PostNewRoomMessage(constant.CHATWORK_ALERT_ROOMID, constant.CHATWORK_APIKEY, msg)
	return err
}

func LogErrorAndAlert(logger *lib_log.Logger, msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	lib_log.Error("%s %s", logger.Prefix, msg)
	PostNewRoomMessage(constant.CHATWORK_ALERT_ROOMID, constant.CHATWORK_APIKEY, msg)
}
