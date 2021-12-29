package lib_session

import (
	"app/src/constant"
	lib_db "app/src/lib/db"
	"app/src/model"
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
)

const DefaultCookieTTL = 60 * 60 * 24 * 30 // 30 Days

type TokenUtil struct {
	dbClient *lib_db.Client
}

func NewTokenUtil(dbClient *lib_db.Client) *TokenUtil {
	return &TokenUtil{
		dbClient: dbClient,
	}
}

func (self *TokenUtil) MakeRememberMeToken(context *gin.Context, userId uint64) error {
	expiresAt := time.Now().Add(time.Second * DefaultCookieTTL)
	pid := strconv.FormatUint(userId, 10)
	// Generate a token
	hash, token, err := self.generateToken(pid)
	if err != nil {
		return err
	}

	// Save Hash to DB
	self.dbClient.StartTransaction()
	defer self.dbClient.RollbackTransaction()
	modelMgr := self.dbClient.GetModelManager()

	userRm := &model.RememberMeToken{
		Hash:      hash,
		UserId:    userId,
		ExpiresAt: expiresAt,
	}

	err = modelMgr.CachedSave(userRm, nil)
	if err != nil {
		return err
	}
	err = modelMgr.WriteAll()
	if err != nil {
		return err
	}
	err = self.dbClient.CommitTransaction()
	if err != nil {
		return err
	}

	// Save token to cookie
	store := securecookie.New([]byte(constant.COOKIE_AUTHKEY), []byte(constant.COOKIE_CIPHERKEY))
	encoded, err := store.Encode(constant.COOKIE_REMEMBER, token)
	if err != nil {
		return err
	}

	if constant.IsProduction() {
		context.SetCookie(constant.COOKIE_NAME, encoded, DefaultCookieTTL, "/", "", true, true)
	} else {
		context.SetCookie(constant.COOKIE_NAME, encoded, DefaultCookieTTL, "/", "", false, true)
	}

	return nil
}

func (self *TokenUtil) ExpireToken(context *gin.Context) error {
	// Get cookie
	ck, errNotFound := context.Cookie(constant.COOKIE_NAME)
	if errNotFound != nil {
		return nil
	}
	_, hash, err := self.decodeToken(ck)
	if err != nil {
		return err
	}
	// Expire cookie
	expireToken(context)

	self.dbClient.StartTransaction()
	defer self.dbClient.RollbackTransaction()
	modelMgr := self.dbClient.GetModelManager()

	err = modelMgr.SetDelete(&model.RememberMeToken{Hash: hash})
	if err != nil {
		return err
	}

	err = modelMgr.WriteAll()
	if err != nil {
		return err
	}
	err = self.dbClient.CommitTransaction()
	if err != nil {
		return err
	}
	return nil
}

func (self *TokenUtil) CheckRememberMeToken(context *gin.Context) (uint64, error) {
	// Check if cookie exists
	ck, errNotFound := context.Cookie(constant.COOKIE_NAME)
	if errNotFound != nil {
		return 0, nil
	}

	pid, hash, err := self.decodeToken(ck)
	if err != nil {
		expireToken(context)
		return 0, nil
	}

	// Get token info stored in DB
	userId, _ := strconv.ParseUint(pid, 10, 64)
	userRm := &model.RememberMeToken{}
	ins, err := self.dbClient.GetModelManager().GetModel(userRm, hash)
	if err != nil || ins == nil {
		expireToken(context)
		return 0, err
	}

	if userId != userRm.UserId {
		expireToken(context)
		return 0, nil
	}

	if userRm.ExpiresAt.Before(time.Now()) {
		// RM Token has expired
		self.dbClient.StartTransaction()
		defer self.dbClient.RollbackTransaction()

		modelMgr := self.dbClient.GetModelManager()

		err = modelMgr.SetDelete(userRm)
		if err != nil {
			return 0, err
		}
		err = modelMgr.WriteAll()
		if err != nil {
			return 0, err
		}
		err = self.dbClient.CommitTransaction()
		if err != nil {
			return 0, err
		}
		return 0, nil
	}

	return userRm.UserId, nil
}

func (self *TokenUtil) ClearRememberMeTokens(userId uint64, whitelist []string) error {
	tokens := []*model.RememberMeToken{}

	// Inital query
	db := self.dbClient.GetDB()
	db = db.Table("user_remember_me_tokens").Where("`account_id` = ?", userId)

	// Apply whitelist
	if len(whitelist) > 0 {
		db = db.Where("`hash` NOT IN(?)", whitelist)
	}

	err := db.Find(&tokens).Error
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	// Delete found tokens
	self.dbClient.StartTransaction()
	defer self.dbClient.RollbackTransaction()

	modelMgr := self.dbClient.GetModelManager()

	for _, token := range tokens {
		err = modelMgr.SetDelete(token)
		if err != nil {
			return err
		}
	}

	err = modelMgr.WriteAll()
	if err != nil {
		return err
	}

	err = self.dbClient.CommitTransaction()
	if err != nil {
		return err
	}
	return nil
}

func (self *TokenUtil) decodeToken(ck string) (string, string, error) {
	// decode token from cookie
	store := securecookie.New([]byte(constant.COOKIE_AUTHKEY), []byte(constant.COOKIE_CIPHERKEY))
	var token string
	err := store.Decode(constant.COOKIE_REMEMBER, ck, &token)
	if err != nil {
		return "", "", err
	}

	tokenBytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return "", "", err
	}

	idx := bytes.IndexByte(tokenBytes, ';')

	pid := string(tokenBytes[:idx])
	sum := sha512.Sum512(tokenBytes)
	hash := base64.StdEncoding.EncodeToString(sum[:])
	return pid, hash, nil
}

func (self *TokenUtil) generateToken(pid string) (string, string, error) {
	tokenBytes := make([]byte, 32+len(pid)+1)
	copy(tokenBytes, pid)
	tokenBytes[len(pid)] = ';'
	_, err := io.ReadFull(rand.Reader, tokenBytes[len(pid)+1:])
	if err != nil {
		return "", "", err
	}
	sum := sha512.Sum512(tokenBytes)
	hash := base64.StdEncoding.EncodeToString(sum[:])
	token := base64.URLEncoding.EncodeToString(tokenBytes)
	return hash, token, nil
}

func expireToken(context *gin.Context) {
	if constant.IsProduction() {
		context.SetCookie(constant.COOKIE_NAME, "", -1, "/", "", true, true)
	} else {
		context.SetCookie(constant.COOKIE_NAME, "", -1, "/", "", false, true)
	}

}
