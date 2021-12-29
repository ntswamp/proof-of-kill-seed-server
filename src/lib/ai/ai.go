package lib_ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
)

const BASE_URL = "http://34.146.144.108:8000"

func ProduceResponse(userName, userInput string, lang string) string {
	//producing cookie value
	jar, err := cookiejar.New(nil)
	if err != nil {
		//handle errors
		log.Fatalln(err)

	}
	httpClient := &http.Client{
		Jar: jar,
	}

	response := "I'm busy right now, try me later."

	//make session
	err = makeSession(userName, httpClient)
	if err != nil {
		return response
	}

	response, err = talk(httpClient, userInput, lang)
	if err != nil {
		return response
	}
	//remove quotes
	response = removeQuotes(response)

	return response
}

func ProduceBinaryResponse(userName, userInput string, lang string) []byte {
	//producing cookie value
	jar, err := cookiejar.New(nil)
	if err != nil {
		//handle errors
		log.Fatalln(err)

	}
	httpClient := &http.Client{
		Jar: jar,
	}

	response := []byte("I'm busy right now, try me later.")

	//make session
	err = makeSession(userName, httpClient)
	if err != nil {
		return response
	}

	response, err = binaryTalk(httpClient, userInput, lang)
	if err != nil {
		return response
	}

	return response
}

func makeSession(userName string, client *http.Client) error {
	reqByte, err := json.Marshal(map[string]string{"name": userName})
	if err != nil {
		return err
	}
	reqBody := bytes.NewReader(reqByte)

	req, err := http.NewRequest(http.MethodPut, BASE_URL+"/create_session/", reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		err = errors.New(BASE_URL + "/whoami/" + "\nresp.StatusCode: " + strconv.Itoa(resp.StatusCode))
		return err
	}

	defer resp.Body.Close()
	return nil
}

type sessionDetail struct {
	UserId            string        `json:"username"`
	ConversationTimes int           `json:"idx"`
	Chatctx           []interface{} `json:"chatctx"`
	Created_at        string        `json:"created_at"`
}

func getSessionDetail(client *http.Client) (*sessionDetail, error) {
	resp, err := client.Get(BASE_URL + "/whoami/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New(BASE_URL + "/whoami/" + "\nresp.StatusCode: " + strconv.Itoa(resp.StatusCode))
		return nil, err
	}

	result := &sessionDetail{}
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func isFirstTimeTalk(client *http.Client) bool {
	detail, _ := getSessionDetail(client)
	if detail != nil {
		return detail.ConversationTimes < 2
	}
	return true
}

func talk(client *http.Client, message string, language string) (string, error) {

	postByte, err := json.Marshal(map[string]string{
		"message": message,
		"lang":    language,
	})
	if err != nil {
		return "", err
	}
	postBody := bytes.NewReader(postByte)

	//is it the first time to talk to AI ?
	endpoint := BASE_URL + "/v1/delta/"
	if isFirstTimeTalk(client) {
		endpoint = BASE_URL + "/v1/alpha/"
	}

	//send request
	resp, err := client.Post(
		endpoint,
		"",
		postBody,
	)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		err = errors.New(BASE_URL + "/v1/alpha(delta)/" + "\nresp.StatusCode: " + strconv.Itoa(resp.StatusCode))
		return "", err
	}

	defer resp.Body.Close()

	//parse
	var result map[string]string

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return "", err
	}

	response := result[language]

	return response, nil
}

func binaryTalk(client *http.Client, message string, language string) ([]byte, error) {

	postByte, err := json.Marshal(map[string]string{
		"message": message,
		"lang":    language,
	})
	if err != nil {
		return nil, err
	}
	postBody := bytes.NewReader(postByte)

	//is it the first time to talk to AI ?
	endpoint := BASE_URL + "/v1/delta/"
	if isFirstTimeTalk(client) {
		endpoint = BASE_URL + "/v1/alpha/"
	}

	//send request
	resp, err := client.Post(
		endpoint,
		"",
		postBody,
	)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		err = errors.New(BASE_URL + "/v1/alpha(delta)/" + "\nresp.StatusCode: " + strconv.Itoa(resp.StatusCode))
		return nil, err
	}

	defer resp.Body.Close()

	//parse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func removeQuotes(s string) string {
	s = strings.TrimPrefix(s, "\"")
	s = strings.TrimSuffix(s, "\"")
	return s
}
