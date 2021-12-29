package handler

import (
	"app/src/constant"
	lib_error "app/src/lib/error"
	lib_redis "app/src/lib/redis"
	lib_session "app/src/lib/session"
	"app/src/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type ApiBaseHandler struct {
	BaseHandler
	Request                 RequestInterface
	Restricted              bool
	PermissionLevel         int
	PermissionLevelRequired int
	Public                  bool
	user                    *model.Account
	JsonParamData           map[string]interface{}
}

func (self *ApiBaseHandler) Setup(context *gin.Context) error {
	err := self.BaseHandler.Setup(context)
	if err != nil {
		return lib_error.WrapError(err)
	}
	context.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	context.Writer.Header().Set("Access-Control-Allow-Headers", "*")
	context.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	context.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	// リクエストのロード.
	err = self.loadRequest()
	if err != nil {
		return lib_error.WrapError(err)
	}

	err = self.checkSession()
	if err != nil {
		return lib_error.WrapError(err)
	}

	self.JsonParamData = map[string]interface{}{}

	//default code
	self.JsonParams["code"] = http.StatusContinue
	self.JsonParams["data"] = self.JsonParamData
	return nil
}

// Jsonレスポンスを返す.
func (self *ApiBaseHandler) writeAppJson(status int) error {
	self.JsonParams["code"] = status
	// Set ActivationStatus if user is not nil
	if self.user != nil {
		self.JsonParams["accountPermission"] = constant.PERMISSION_TEXT[self.user.Type]
		if self.user.Type == constant.USER {
			self.PermissionLevel = constant.USER
		}
	}
	return self.WriteJson()
}

// Jsonレスポンスを返す.
func (self *ApiBaseHandler) WriteSuccessJson() error {
	return self.writeAppJson(http.StatusOK)
}

// リクエストのロード.
func (self *ApiBaseHandler) loadRequest() error {
	var BodyLength int = 2097152
	context := self.GetContext()
	if self.Request == nil {
		self.Request = &ApiRequest{}
	}

	contentLength := context.Request.Header.Get("Content-Length")
	if context.Request.Method == http.MethodGet && contentLength == "" {
		return nil
	}
	// bodyの長さ.
	length, err := strconv.Atoi(contentLength)
	if err != nil {
		return lib_error.WrapError(err)
	} else if BodyLength < length {
		// 大きすぎる.
		return lib_error.NewAppErrorWithStackTrace(http.StatusRequestEntityTooLarge, "'Content-Length':too long...")
	} else if length == 0 {
		return nil
	}
	// bodyの長さ分read.
	buf := []byte{}
	tmpBuf := make([]byte, BodyLength)
	for len(buf) < length {
		size, err := context.Request.Body.Read(tmpBuf)
		if err == nil || err == io.EOF {
			buf = append(buf, tmpBuf[:size]...)
		}
		if err != nil {
			break
		}
	}
	if err != nil && err != io.EOF {
		return lib_error.WrapError(err)
	}
	// jsonのデコード.
	err = json.Unmarshal(buf, self.Request)
	if err != nil {
		return lib_error.WrapError(err)
	}
	return nil
}

func (self *ApiBaseHandler) GetUser() *model.Account {
	return self.user
}

func (self *ApiBaseHandler) SetUser(user *model.Account) {
	self.user = user
}

func (self *ApiBaseHandler) Process() error {
	return lib_error.NewAppErrorWithStackTrace(http.StatusContinue, "unimplemented")
}

func (self *ApiBaseHandler) ProcessError(err error) error {
	// ステータス.
	code := http.StatusContinue
	switch err.(type) {
	case *lib_error.AppError:
		code = err.(*lib_error.AppError).Code
	}
	if self.IsDebugRequest {
		self.JsonParams["Error"] = err.Error()
	}
	// データを削除.
	if _, exists := self.JsonParams["data"]; exists {
		delete(self.JsonParams, "data")
	}
	return self.writeAppJson(code)
}

func (self *ApiBaseHandler) getUserAccount(id uint64) (*model.Account, error) {
	if id == 0 {
		return nil, nil
	}
	manager, err := self.GetModelManager(constant.DbIdolverse)
	if err != nil {
		return nil, lib_error.WrapError(err)
	}
	user := &model.Account{}
	ins, err := manager.GetModel(user, id)
	if err != nil {
		return nil, lib_error.WrapError(err)
	} else if ins == nil {
		return nil, nil
	}
	return user, nil
}

// Check if user is authorized
func (self *ApiBaseHandler) CheckUser() error {
	if self.Public || self.GetIsStaff() {
		return nil
	}
	if self.user == nil {
		return lib_error.NewAppError(http.StatusUnauthorized, "access is restricted")
	}
	if self.PermissionLevel < self.PermissionLevelRequired {
		return lib_error.NewAppError(http.StatusUnauthorized, "permission level is too low")
	}
	return nil
}

func (self *ApiBaseHandler) checkToken() error {
	context := self.GetContext()

	dbClient, err := self.GetDbClient(constant.DbIdolverse)
	if err != nil {
		return err
	}

	tokenUtil := lib_session.NewTokenUtil(dbClient)
	userId, err := tokenUtil.CheckRememberMeToken(context)
	if err != nil {
		return err
	}

	self.user, err = self.getUserAccount(userId)
	if err != nil {
		return err
	}

	if self.user != nil {
		redisSession, err := self.GetRedisClient(constant.RedisSession)
		if err != nil {
			return lib_error.WrapError(err)
		}
		sessionUtil, err := lib_session.NewSessionUtil(constant.SESSION_STORE_PREFIX, redisSession)
		if err != nil {
			return lib_error.WrapError(err)
		}
		err = sessionUtil.MakeNewSession(context, constant.SESSION_ID_KEY, lib_session.DefaultSessionTTL, userId, true)
		if err != nil {
			return lib_error.WrapError(err)
		}
	}
	return nil
}

func (self *ApiBaseHandler) checkSession() error {
	context := self.GetContext()

	// Check Session
	session := sessions.Default(context)
	sessId, ok := session.Get(constant.SESSION_ID_KEY).(string)
	if ok && sessId != "" {
		redisSession, err := self.GetRedisClient(constant.RedisSession)
		if err != nil {
			return lib_error.WrapError(err)
		}
		sessUtil, err := lib_session.NewSessionUtil(constant.SESSION_STORE_PREFIX, redisSession)
		if err != nil {
			return lib_error.WrapError(err)
		}
		var userId uint64
		err = sessUtil.GetSessionInfo(sessId, &userId)
		if err != nil {
			return lib_error.WrapError(err)
		}
		if userId != 0 {
			self.user, err = self.getUserAccount(userId)
			if err != nil {
				return lib_error.WrapError(err)
			}
			if self.user != nil {
				sessUtil.RefreshSession(sessId, lib_session.DefaultSessionTTL)
			}
		} else {
			err = sessUtil.ClearSession(context, sessId)
			if err != nil {
				return lib_error.WrapError(err)
			}
		}
	}
	if self.user == nil {
		err := self.checkToken()
		if err != nil {
			return lib_error.WrapError(err)
		}
	}
	return nil
}

func (self *ApiBaseHandler) CheckGlobalRateLimit() error {
	context := self.GetContext()
	clientIp := context.ClientIP()

	key := fmt.Sprintf("GlobalRateLimit:%s", clientIp)

	redisClient, err := self.GetRedisClient(constant.RedisDb)
	if err != nil {
		return lib_error.WrapError(err)
	}
	// Get current attempts
	var current int
	_, err = redisClient.Get(&current, key)
	if err != nil {
		return lib_error.WrapError(err)
	}
	// Check if we hit the cap
	if current >= constant.GLOBAL_RATE_LIMIT {
		self.SetHttpStatus(http.StatusTooManyRequests)
		return lib_error.NewAppError(http.StatusTooManyRequests, "reach rate limitation")
	} else if current == 0 {
		err = redisClient.Multi(func(redisClient *lib_redis.Client, args ...interface{}) error {
			key := args[0].(string)
			ttl := args[1].(int64)
			_, err := redisClient.Incr(key)
			if err != nil {
				return err
			}
			err = redisClient.Expire(key, ttl)
			return err
		}, key, constant.GLOBAL_RATE_TTL)
		if err != nil {
			return lib_error.WrapError(err)
		}
	} else {
		// Increment by 1
		_, err := redisClient.Incr(key)
		if err != nil {
			return lib_error.WrapError(err)
		}
	}
	return nil
}

func (self *ApiBaseHandler) CheckRateLimit() error {
	if !self.RateLimit {
		return nil
	}
	if self.RateLimitTTL < 1 {
		self.RateLimitTTL = 1
	}
	if self.RateLimitCap < 1 {
		self.RateLimitCap = 1
	}
	// Rate Limit
	context := self.GetContext()
	clientIp := context.ClientIP()
	requestPath := context.Request.URL.Path

	key := fmt.Sprintf("RateLimit-%s:%s", requestPath, clientIp)
	// Get Redis Client
	redisClient, err := self.GetRedisClient(constant.RedisDb)
	if err != nil {
		return lib_error.WrapError(err)
	}
	// Get current attempts
	var current int
	_, err = redisClient.Get(&current, key)
	if err != nil {
		return lib_error.WrapError(err)
	}
	// Check if we hit the cap
	if current >= self.RateLimitCap {
		self.SetHttpStatus(http.StatusTooManyRequests)
		return lib_error.NewAppError(http.StatusTooManyRequests, "reach rate limitation")
	} else {
		// Increment by 1
		v, err := redisClient.Incr(key)
		if err != nil {
			return lib_error.WrapError(err)
		}
		// If value(counter) is at 1 we set the ttl for this key
		if v.(int64) == 1 {
			err = redisClient.Expire(key, self.RateLimitTTL)
			if err != nil {
				return lib_error.WrapError(err)
			}
		}
	}
	return nil
}

type RequestInterface interface {
	GetLanguageCode() string
	GetUser() string
}

type ApiRequest struct {
	Lang string
	User string
}

func (self *ApiRequest) GetLanguageCode() string { return self.Lang }
func (self *ApiRequest) GetUser() string         { return self.User }
