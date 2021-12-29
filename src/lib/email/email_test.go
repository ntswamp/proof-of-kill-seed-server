package lib_email

import (
	"app/src/constant"
	"fmt"
	"testing"
	"time"
)

// Test sending plain text email
func TestSendEmail(t *testing.T) {
	// Get email util for smtp settings "test"
	emailUtil, err := NewEmailUtil("test")
	if err != nil {
		t.Fatal(err)
		return
	}
	// Prepare a new email
	link := `<a href="http://google.com">Link</a>`
	msg := fmt.Sprintf("Test %s", link)
	req, err := emailUtil.NewEmailRequest([]string{"tauriktester@gmail.com"}, "Test Subject", msg)
	if err != nil {
		t.Fatal(err)
		return
	}
	// send the email
	err = req.SendMail()
	if err != nil {
		t.Fatal(err)
		return
	}
}

// Test sending an email using a parse html template
func TestTemplateEmail(t *testing.T) {
	// Get email util for smtp settings "test"
	emailUtil, err := NewEmailUtil("test")
	if err != nil {
		t.Fatal(err)
		return
	}
	// Variable data for the template
	templateData := OfferAcceptedTemplate{
		ItemThumbnailUrl: "https://s3-ap-northeast-1.amazonaws.com/crosslink-static-dev/thumbnails/Weapon/weapon_r1sword_001_w_w.png",
		UserHistoryUrl:   "http://127.0.0.1:5500/static/account-history.html",
		UserName:         "Chris",
		ItemName:         "TestItem",
		ItemUrl:          "www.google.com",
		SentDate:         time.Now().Format("2006-01-02 15:04:05"),
	}
	// Prepare new email
	req, err := emailUtil.NewEmailRequest([]string{"tauriko@gmail.com"}, "TemplateTest", "")
	if err != nil {
		t.Fatal(err)
		return
	}
	// Set our email body to template
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/offer-accepted-mail-ja.html", templateData)
	if err != nil {
		t.Fatal(err)
		return
	}
	// Send email
	err = req.SendMail()
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestParseTemplates(t *testing.T) {
	thumbnail := "www.thumbnail.com"
	history := "www.historyurl.com"
	username := "TestUser"
	itemName := "Rare Item"
	itemUrl := "www.itemurl.com"
	tokenId := "cl_test"
	sentDate := time.Now().Format("2006-01-02 15:04:05")
	price := "100"
	currency := "IOST"

	// Get email util for smtp settings "test"
	emailUtil, err := NewEmailUtil("test")
	if err != nil {
		t.Fatal(err)
		return
	}

	// Prepare new email
	req, err := emailUtil.NewEmailRequest([]string{""}, "", "")
	if err != nil {
		t.Fatal(err)
		return
	}

	//Offer Accepted
	offerAccepted := OfferAcceptedTemplate{
		ItemThumbnailUrl: thumbnail,
		UserHistoryUrl:   history,
		UserName:         username,
		ItemName:         itemName,
		ItemUrl:          itemUrl,
		SentDate:         sentDate,
		TokenId:          tokenId,
	}
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/offer-accepted-mail-ja.html", offerAccepted)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/offer-accepted-mail-en.html", offerAccepted)
	if err != nil {
		t.Fatal(err)
		return
	}
	//Offer Expired
	offerExpired := OfferExpiredTemplate{
		ItemThumbnailUrl: thumbnail,
		UserHistoryUrl:   history,
		UserName:         username,
		ItemName:         itemName,
		ItemUrl:          itemUrl,
		SentDate:         sentDate,
		TokenId:          tokenId,
	}
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/offer-expired-mail-ja.html", offerExpired)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/offer-expired-mail-en.html", offerExpired)
	if err != nil {
		t.Fatal(err)
		return
	}
	//Offer Received
	offerReceived := OfferReceivedTemplate{
		ItemThumbnailUrl: thumbnail,
		UserHistoryUrl:   history,
		UserName:         username,
		ItemName:         itemName,
		ItemUrl:          itemUrl,
		SentDate:         sentDate,
		TokenId:          tokenId,
		Price:            price,
		Currency:         currency,
	}
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/offer-receive-mail-ja.html", offerReceived)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/offer-receive-mail-en.html", offerReceived)
	if err != nil {
		t.Fatal(err)
		return
	}
	//Bid Received
	bidTemplate := AuctionBidReceivedTemplate{
		ItemThumbnailUrl: thumbnail,
		UserHistoryUrl:   history,
		UserName:         username,
		ItemName:         itemName,
		ItemUrl:          itemUrl,
		SentDate:         sentDate,
		TokenId:          tokenId,
		Price:            price,
		Currency:         currency,
	}
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/auction-bid-received-mail-ja.html", bidTemplate)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/auction-bid-received-mail-en.html", bidTemplate)
	if err != nil {
		t.Fatal(err)
		return
	}
	//Auction Expired
	auctionExpired := AuctionExpiredTemplate{
		ItemThumbnailUrl: thumbnail,
		UserHistoryUrl:   history,
		UserName:         username,
		ItemName:         itemName,
		ItemUrl:          itemUrl,
		SentDate:         sentDate,
		TokenId:          tokenId,
	}
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/auction-expired-mail-ja.html", auctionExpired)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/auction-expired-mail-en.html", auctionExpired)
	if err != nil {
		t.Fatal(err)
		return
	}
	//Auction Sold
	auctionSold := AuctionSoldTemplate{
		ItemThumbnailUrl: thumbnail,
		UserHistoryUrl:   history,
		UserName:         username,
		ItemName:         itemName,
		ItemUrl:          itemUrl,
		SentDate:         sentDate,
		TokenId:          tokenId,
		Price:            price,
		Currency:         currency,
	}
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/auction-sold-mail-ja.html", auctionSold)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/auction-sold-mail-en.html", auctionSold)
	if err != nil {
		t.Fatal(err)
		return
	}
	//Auction Won
	auctionWon := AuctionWonTemplate{
		ItemThumbnailUrl: thumbnail,
		UserHistoryUrl:   history,
		UserName:         username,
		ItemName:         itemName,
		ItemUrl:          itemUrl,
		SentDate:         sentDate,
		TokenId:          tokenId,
		Price:            price,
		Currency:         currency,
	}
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/auction-won-mail-ja.html", auctionWon)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = req.ParseTemplate(constant.TEMPLATE_DIR+"email/auction-won-mail-en.html", auctionWon)
	if err != nil {
		t.Fatal(err)
		return
	}

}

func TestIsValidEmail(t *testing.T) {
	var testCases = map[string]bool{
		"dskasdkao....&&&@": false,
		"a@a.com":           true,
		"a+1@a.com":         false,
	}
	for adr, isValid := range testCases {
		validationResult := IsValidEmail(adr)
		if validationResult != isValid {
			t.Fail()
		}

	}
}
