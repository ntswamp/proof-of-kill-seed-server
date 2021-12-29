package lib_ai

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"testing"
)

func TestMakeSessionANDGetSessionDetail(t *testing.T) {
	//producing cookie value
	jar, err := cookiejar.New(nil)
	if err != nil {
		//handle errors
		t.Error("error:", err)

	}
	httpClient := &http.Client{
		Jar: jar,
	}

	err = makeSession("test123", httpClient)
	if err != nil {
		t.Error("error:", err)
	}
	sd, err := getSessionDetail(httpClient)
	if err != nil {
		t.Error("error:", err)

	}

	if sd.UserId != "test123" {
		t.Error("error:username differs")
	}
}

func TestTalk(t *testing.T) {
	//producing cookie value
	jar, err := cookiejar.New(nil)
	if err != nil {
		//handle errors
		log.Fatalln(err)

	}
	httpClient := &http.Client{
		Jar: jar,
	}

	err = makeSession("hanbyo", httpClient)
	if err != nil {
		t.Error("error:", err)
	}

	detail, err := getSessionDetail(httpClient)
	if err != nil {
		t.Error("error:", err)
	}
	fmt.Println("before talking")
	fmt.Println(detail.ConversationTimes, detail.UserId)

	res, err := talk(httpClient, "何時ですか", "zh")
	if err != nil {
		t.Error("error:", err)
	}
	println(res)

	detail, err = getSessionDetail(httpClient)
	if err != nil {
		t.Error("error:", err)
	}
	fmt.Println("after talking")
	fmt.Println(detail.ConversationTimes, detail.UserId)

}
