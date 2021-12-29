package handler

type NameGetHandler struct {
	ApiBaseHandler
}

type NameGetRequest struct {
	ApiRequest
	Data NameGetRequestData
}

type NameGetRequestData struct {
	RegistrationType int

	MailAddress     string
	Password        string
	PasswordConfirm string

	IostAccountName string
	IostMsg         string
	IostSign        string
	IostPubKey      string
	IostTxHash      string
	IostTxData      string
	IostTxSecret    string
}

/*
func (self *RegisterHandler) Setup(context *gin.Context) error {
	err := self.ApiBaseHandler.Setup(context)
	if err != nil {
		return lib_error.WrapError(err)
	}
	self.Public = true
	return nil
}

func (self *RegisterHandler) Process() error {
	data := self.Request.(*UserRegisterRequest).Data

	currentUser := self.GetUser()
	if currentUser != nil {
		return lib_error.NewAppError(http.StatusNotAcceptable, "already logged in")
	}

	switch data.RegistrationType {
	case RegistrationType.MailAddress:
		err := self.tryRegisterEmail(data.MailAddress, data.Password, data.PasswordConfirm)
		if err != nil {
			return lib_error.WrapError(err)
		}
	case RegistrationType.IostSign:
		err := self.tryRegisterIostSign(data.IostAccountName, data.IostMsg, data.IostSign, data.IostPubKey)
		if err != nil {
			return lib_error.WrapError(err)
		}
	case RegistrationType.IostTx:
		err := self.tryRegisterIostTx(data.IostAccountName, data.IostTxHash, data.IostTxData, data.IostTxSecret)
		if err != nil {
			return lib_error.WrapError(err)
		}
	default:
		return lib_error.NewAppError(http.StatusBadRequest, "illegal registration type")
	}

	// Possibly add auto-login here?

	return self.WriteSuccessJson()
}

func (self *RegisterHandler) tryRegisterEmail(emailAddress, password, passwordConfirm string) error {
	now := time.Now()
	// Password Check
	if len(password) < 8 || len(password) > 255 {
		return lib_error.NewAppError(http.StatusBadRequest, "illegal password")
	}

	if password != passwordConfirm {
		return lib_error.NewAppError(http.StatusBadRequest, "password and confirm do not match")
	}

	// Check if email is a valid email
	if !lib_email.IsValidEmail(emailAddress) {
		return lib_error.NewAppError(http.StatusBadRequest, "illegal email format")
	}

	// Hash and Encrypt email
	emailHash := lib_email.HashEmail(emailAddress)
	emailEnc, err := lib_email.EncryptEmailToDbFormat(emailAddress)
	if err != nil {
		return lib_error.WrapError(err)
	}

	// Make Auth Code for email verification
	authCode, err := lib_email.GenerateAuthCode()
	if err != nil {
		return lib_error.WrapError(err)
	}

	// Get DbClient/Manager
	dbClient, err := self.GetDbClient(constant.DbIdolverse)
	if err != nil {
		return lib_error.WrapError(err)
	}
	modelMgr := dbClient.GetModelManager()

	// check if email address already taken
	record := &model.UserAccount{}
	err = dbClient.GetDB().Where("`email` = ?", emailAddress).Find(&record).Error
	if err != nil {
		return lib_error.WrapError(err)
	}
	if record != nil {
		return lib_error.NewAppError(http.StatusConflict, "email address is already taken")
	}

	// Hash PW
	pwHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return lib_error.WrapError(err)
	}

	// Set up and save models for user account
	newUserAccount := &model.UserAccount{
		Email:        emailAddress,
		Password:     pwHash,
		RegisteredAt: now,
		EmailTmp:     *emailConfirmTmp,
	}
	err = self.saveModel(newUserAccount, dbClient)
	if err != nil {
		return lib_error.WrapError(err)
	}
	//
	// Send activation email here
	//
	emailUtil, emailErr := lib_email.NewEmailUtil("sale")
	if emailErr == nil {
		lang := self.Request.GetLanguageCode()
		return SendAccountConfirmationEmail(authCode, emailAddress, lang, emailUtil)
	}
	return nil
}

func (self *RegisterHandler) tryRegisterIostSign(iostAccountName, iostMsg, iostSign, iostPubKey string) error {
	if iostAccountName == "" {
		return lib_error.NewAppError(app_defines_tokenlink.StatusCode.IostInvalidAccount, "IostAccountName is empty")
	}
	// Check if Msg is stored
	redisClient, err := self.GetRedisClient(constant.RedisCache)
	if err != nil {
		return lib_error.WrapError(err)
	}
	msgUser, err := app_util_user.GetIostMsgFromStore(redisClient, iostMsg)
	if err != nil {
		return lib_error.WrapError(err)
	}
	if iostAccountName != msgUser {
		return lib_error.NewAppError(http.StatusBadRequest, "IostAccountName does not match user for msg")
	}

	// Get DbClient/Manager
	dbClient, err := self.GetDbClient(constant.DbIdolverse)
	if err != nil {
		return lib_error.WrapError(err)
	}
	modelMgr := dbClient.GetModelManager()

	// Check if this wallet is already registered or not
	userWallet := &app_models_tokenlink.UserWalletAddress{}
	ins, err := modelMgr.GetModel(userWallet, iostAccountName)
	if err != nil {
		return lib_error.WrapError(err)
	} else if ins != nil || userWallet.AccountId != 0 {
		return lib_error.NewAppError(app_defines_tokenlink.StatusCode.IostInvalidAccount, "User already registered")
	}
	// Check if the pubkey matches the account name
	iostSdk, err := app_util_iost.NewIOSTDevSDK()
	if err != nil {
		return lib_error.WrapError(err)
	}

	decPubKey, err := app_util_iwallet.DecodeIWalletKey(iostPubKey)
	if err != nil {
		return lib_error.WrapError(err)
	}
	pubKeyStr := app_util_iwallet.IwalletKeyToString(decPubKey)
	isPubKeyOwner, err := app_util_iost.IsPubKeyOwner(iostSdk, iostAccountName, pubKeyStr)
	if err != nil {
		return lib_error.WrapError(err)
	}
	if !isPubKeyOwner {
		return lib_error.NewAppError(app_defines_tokenlink.StatusCode.IostInvalidPubKey, "Not PubKey Owner")
	}

	// Check if sig is real
	realSign, err := app_util_iwallet.VerifySign(iostMsg, iostPubKey, iostSign)
	if err != nil {
		return lib_error.WrapError(err)
	}
	if !realSign {
		return lib_error.NewAppError(app_defines_tokenlink.StatusCode.IostSignatureNoMatch, "Signature does not match")
	}

	// Everything seems OK, create user
	newUserWallet := &app_models_tokenlink.UserWalletAddress{
		WalletAddress: iostAccountName,
	}
	newUser := &app_models_tokenlink.UserAccount{
		AccountType:      app_defines_tokenlink.UserAccountType.Iost,
		ActivationStatus: app_defines_tokenlink.UserAccountActivationStatus.Complete,
		CreatedTime:      time.Now(),

		Wallet: *newUserWallet,
	}
	err = self.saveModel(newUser, dbClient)
	if err != nil {
		return lib_error.WrapError(err)
	}
	return nil
}

func (self *RegisterHandler) tryRegisterIostTx(iostAccountName, iostTxHash, iostTxData, iostTxSecret string) error {
	if len(iostAccountName) == 0 {
		return lib_error.NewAppError(app_defines_tokenlink.StatusCode.IostInvalidAccount, "IostAccountName is empty")
	}
	if len(iostTxSecret) == 0 || len(iostTxData) == 0 {
		return lib_error.NewAppError(http.StatusBadRequest, "txData is empty")
	}
	// Check if Msg is stored
	redisClient, err := self.GetRedisClient(constant.RedisCache)
	if err != nil {
		return lib_error.WrapError(err)
	}

	storedTxData, err := app_util_user.GetIostMsgFromStore(redisClient, iostTxSecret)
	if err != nil {
		return lib_error.WrapError(err)
	}

	if len(storedTxData) == 0 {
		return lib_error.NewAppError(http.StatusBadRequest, "storedTxData is empty")
	}

	// Delete msg
	err = app_util_user.DeleteIostMsgFromStore(redisClient, iostTxSecret)
	if err != nil {
		return lib_error.WrapError(err)
	}

	// Get DbClient/Manager
	dbClient, err := self.GetDbClient(constant.DbIdolverse)
	if err != nil {
		return lib_error.WrapError(err)
	}
	modelMgr := dbClient.GetModelManager()
	// Check if this wallet is already registered or not
	userWallet := &app_models_tokenlink.UserWalletAddress{}
	ins, err := modelMgr.GetModel(userWallet, iostAccountName)
	if err != nil {
		return lib_error.WrapError(err)
	} else if ins != nil || userWallet.AccountId != 0 {
		return lib_error.NewAppError(app_defines_tokenlink.StatusCode.IostInvalidAccount, "User already registered")
	}

	iostSdk, err := app_util_iost.NewIOSTDevSDK()
	if err != nil {
		return lib_error.WrapError(err)
	}

	tx, err := app_util_iost.GetTxByHash(iostSdk, iostTxHash)
	if err != nil {
		return lib_error.WrapError(err)
	}

	publisher := tx.GetTransaction().GetPublisher()
	if publisher != iostAccountName {
		return lib_error.NewAppError(app_defines_tokenlink.StatusCode.IostInvalidAccount, "IostAccountName does not match tx publisher")
	}

	receipts := tx.GetTransaction().GetTxReceipt().GetReceipts()
	if len(receipts) != 1 {
		return lib_error.NewAppError(http.StatusBadRequest, "Unexpected receipt length for tx")
	}
	receipt := receipts[0]
	funcName := strings.Split(receipt.GetFuncName(), "/")[1]
	if funcName != "sign" {
		return lib_error.NewAppError(http.StatusBadRequest, "Unexpected funcName for tx")
	}
	content := receipt.GetContent()
	txData := []string{}
	err = json.Unmarshal([]byte(content), &txData)
	if err != nil {
		return lib_error.WrapError(err)
	}
	if len(txData) != 1 {
		return lib_error.NewAppError(http.StatusBadRequest, "Unexpected txData length for tx")
	}
	if txData[0] != storedTxData || txData[0] != iostTxData {
		return lib_error.NewAppError(http.StatusBadRequest, "Unexpected txData for tx")
	}

	newUserWallet := &app_models_tokenlink.UserWalletAddress{
		WalletAddress: iostAccountName,
	}
	// TODO: Everything seems OK, create user
	newUser := &lib_db.UserAccount{
		//TODO
	}
	err = self.saveModel(newUser, dbClient)
	if err != nil {
		return lib_error.WrapError(err)
	}
	return nil
}

func SendAccountConfirmationEmail(authCode, to, lang string, emailUtil *EmailUtil) error {
	siteUrl := constant.GetSaleSiteURL()

	loc, _ := time.LoadLocation("Asia/Tokyo")

	templateData := &AccountConfirmationTemplate{
		Url:      fmt.Sprintf("%s/account-create-mail-confirm.html?key=%s", siteUrl, authCode),
		SentDate: time.Now().In(loc).Format("2006-01-02 15:04:05 MST"),
	}

	var subject string
	templateName := constant.TEMPLATE_DIR
	switch lang {
	case app_defines_tokenlink.LanguageCode.En:
		templateName += "email/account-confirmation-mail-en.html"
		subject = "TOKEN LINK Information for completing your registration"
	case app_defines_tokenlink.LanguageCode.Ja:
		templateName += "email/account-confirmation-mail-ja.html"
		subject = "TOKEN LINKアカウント登録用メール送付のお知らせ"
	default:
		return lib_error.NewAppError(http.StatusBadRequest, "Invalid Language Code")
	}

	// Prepare new email
	emailReq, err := emailUtil.NewEmailRequest([]string{to}, subject, "")
	if err != nil {
		return lib_error.WrapError(err)
	}
	// Set our email body to template
	err = emailReq.ParseTemplate(templateName, templateData)
	if err != nil {
		return lib_error.WrapError(err)
	}
	// Send email
	err = emailReq.SendMail()
	if err != nil {
		return lib_error.WrapError(err)
	}
	return nil
}

func (self *RegisterHandler) saveModel(model interface{}, dbClient *lib_db.Client) error {
	dbClient.StartTransaction()
	defer dbClient.RollbackTransaction()

	modelMgr := dbClient.GetModelManager()

	err := modelMgr.CachedSave(model, nil)
	if err != nil {
		return lib_error.WrapError(err)
	}
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


*/

func NameGet() HandlerInterface {
	handler := &NameGetHandler{}
	handler.Request = &NameGetRequest{}
	return handler
}
