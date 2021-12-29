package lib_email

type TestTemplate struct {
	TestLink1 string
	TestLink2 string
}

type OfferAcceptedTemplate struct {
	ItemThumbnailUrl string
	UserHistoryUrl   string
	UserName         string
	ItemName         string
	ItemUrl          string
	SentDate         string
	TokenId          string
}

type OfferExpiredTemplate struct {
	ItemThumbnailUrl string
	UserHistoryUrl   string
	UserName         string
	ItemName         string
	ItemUrl          string
	SentDate         string
	TokenId          string
}

type OfferReceivedTemplate struct {
	ItemThumbnailUrl string
	UserHistoryUrl   string
	UserName         string
	ItemName         string
	ItemUrl          string
	SentDate         string
	TokenId          string
	Price            string
	Currency         string
}

type AuctionBidReceivedTemplate struct {
	ItemThumbnailUrl string
	UserHistoryUrl   string
	UserName         string
	ItemName         string
	ItemUrl          string
	SentDate         string
	TokenId          string
	Price            string
	Currency         string
}

type AuctionExpiredTemplate struct {
	ItemThumbnailUrl string
	UserHistoryUrl   string
	UserName         string
	ItemName         string
	ItemUrl          string
	SentDate         string
	TokenId          string
}

type AuctionSoldTemplate struct {
	ItemThumbnailUrl string
	UserHistoryUrl   string
	UserName         string
	ItemName         string
	ItemUrl          string
	SentDate         string
	TokenId          string
	Price            string
	Currency         string
}

type AuctionWonTemplate struct {
	ItemThumbnailUrl string
	UserHistoryUrl   string
	UserName         string
	ItemName         string
	ItemUrl          string
	SentDate         string
	TokenId          string
	Price            string
	Currency         string
}

type AccountConfirmationTemplate struct {
	Url      string
	SentDate string
}

type SaleCommentAddedTemplate struct {
	ItemThumbnailUrl string
	UserHistoryUrl   string
	Sender           string
	Receiver         string
	ItemName         string
	ItemUrl          string
	SentDate         string
}

type AuctionBidUpdateTemplate struct {
	ItemThumbnailUrl string
	UserHistoryUrl   string
	UserName         string
	NewBidAmount     string
	Currency         string
	ItemName         string
	ItemUrl          string
	SentDate         string
}

type PasswordRecoveryTemplate struct {
	Url      string
	SentDate string
}
