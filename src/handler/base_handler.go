package handler

import (
	"fmt"
	"net/http"
	"time"

	"app/src/constant"
	lib_db "app/src/lib/db"
	lib_error "app/src/lib/error"
	lib_log "app/src/lib/log"
	lib_redis "app/src/lib/redis"

	"github.com/gin-gonic/gin"
)

type HandlerInterface interface {
	Setup(*gin.Context) error
	Terminate()
	GetContext() *gin.Context
	GetDbClient(string) (*lib_db.Client, error)
	GetModelManager(...string) (*lib_db.ModelManager, error)
	GetRedisClient(string) (*lib_redis.Client, error)
	GetNow() time.Time
	SetHttpStatus(int)
	WriteString(string) error
	WriteHtml(string) error
	WriteJson() error

	CheckGlobalRateLimit() error
	CheckRateLimit() error
	PreProcess() error
	Process() error
	CheckUser() error
	IsEnd() bool
	ProcessError(error) error
}

type BaseHandler struct {
	context        *gin.Context
	dbClient       map[string]*lib_db.Client
	redisClient    map[string]*lib_redis.Client
	now            time.Time
	IsDebugRequest bool
	httpStatus     int
	HtmlParams     gin.H
	JsonParams     map[string]interface{}
	staff          bool
	RedisCache     string
	RateLimit      bool
	RateLimitCap   int
	RateLimitTTL   int64
}

// ベースハンドラのセットアップ.
func (self *BaseHandler) Setup(context *gin.Context) error {
	self.context = context
	self.dbClient = map[string]*lib_db.Client{}
	self.redisClient = map[string]*lib_redis.Client{}
	self.now = time.Now().UTC()
	self.httpStatus = http.StatusOK
	self.HtmlParams = gin.H{}
	self.IsDebugRequest = constant.IsDebug()
	self.JsonParams = map[string]interface{}{}
	self.staff, _ = constant.DEV_IP[context.ClientIP()]
	return nil
}

// 終了処理.
func (self *BaseHandler) Terminate() {
	for _, client := range self.dbClient {
		if client == nil {
			continue
		}
		client.Close()
	}
	for _, client := range self.redisClient {
		if client == nil {
			continue
		}
		client.Terminate()
	}
}

// ginのContext.
func (self *BaseHandler) GetContext() *gin.Context {
	return self.context
}

// レスポンス返却済みフラグ.
func (self *BaseHandler) IsEnd() bool {
	return self.GetContext().GetBool("isEnd")
}

// レスポンス返却済みフラグを設定.
func (self *BaseHandler) setEnd() {
	self.GetContext().Set("isEnd", true)
}

// 開発者フラグ.
func (self *BaseHandler) GetIsStaff() bool {
	return self.staff
}

// dbの接続クライアントを取得.
func (self *BaseHandler) GetDbClient(name string) (*lib_db.Client, error) {
	client, _ := self.dbClient[name]
	if client != nil {
		return client, nil
	}
	var err error = nil
	// キャッシュ用redisクライアント.
	dbName := self.RedisCache
	if len(dbName) == 0 {
		dbName = constant.RedisCache
	}
	redis, err := self.GetRedisClient(dbName)
	// DBクライアント.
	client, err = lib_db.Connect(name, redis)
	if err != nil {
		return nil, lib_error.WrapError(err)
	}
	self.dbClient[name] = client
	return client, nil
}

// dbのモデル管理インスタンスを取得.
func (self *BaseHandler) GetModelManager(args ...string) (*lib_db.ModelManager, error) {
	name := constant.DbIdolverse
	if 0 < len(args) {
		name = args[0]
	}
	client, err := self.GetDbClient(name)
	if err != nil {
		return nil, lib_error.WrapError(err)
	}
	return client.GetModelManager(), nil
}

// Redisの接続クライアントを取得.
func (self *BaseHandler) GetRedisClient(name string) (*lib_redis.Client, error) {
	redisClient, _ := self.redisClient[name]
	if redisClient != nil {
		return redisClient, nil
	}
	var err error = nil
	redisClient, err = lib_redis.NewClient(name)
	if err != nil {
		return nil, lib_error.WrapError(err)
	}
	self.redisClient[name] = redisClient
	return redisClient, nil
}

// 現在時刻.
func (self *BaseHandler) GetNow() time.Time {
	return self.now
}

func (self *BaseHandler) CheckGlobalRateLimit() error {
	return nil
}

func (self *BaseHandler) CheckRateLimit() error {
	return nil
}

// processの直前に呼ばれます.
func (self *BaseHandler) PreProcess() error {
	return nil
}

// メイン処理.
func (self *BaseHandler) Process() error {
	return nil
}

// ユーザーの確認.
func (self *BaseHandler) CheckUser() error {
	return nil
}

// エラー処理.
func (self *BaseHandler) ProcessError(err error) error {
	self.httpStatus = http.StatusInternalServerError
	self.WriteString(err.Error())
	lib_log.Error("ProcessError:StatusInternalServerError: %+v", err)
	return nil
}

func (self *BaseHandler) SetHttpStatus(status int) {
	self.httpStatus = status
}

// レスポンスのbodyにtextを書き込む.
func (self *BaseHandler) WriteString(text string) error {
	context := self.GetContext()
	context.String(self.httpStatus, text)
	err := context.Err()
	if err != nil {
		return err
	}
	self.setEnd()
	return nil
}

// レスポンスのbodyにhtmlを書き込む.
func (self *BaseHandler) WriteHtml(htmlName string) error {
	context := self.GetContext()
	context.Writer.Header().Set("Cache-Control", "private")
	context.HTML(self.httpStatus, htmlName, self.HtmlParams)
	err := context.Err()
	if err != nil {
		return err
	}
	self.setEnd()
	return nil
}

// レスポンスのbodyにjsonを書き込む.
func (self *BaseHandler) WriteJson() error {
	context := self.GetContext()
	context.Writer.Header().Set("Cache-Control", "no-cache")
	if self.IsDebugRequest {
		// デバッグ時はインデントを付けた形で返す.
		context.IndentedJSON(self.httpStatus, self.JsonParams)
	} else {
		context.SecureJSON(self.httpStatus, self.JsonParams)
	}
	err := context.Err()
	if err != nil {
		return err
	}
	self.setEnd()
	return nil
}

// リダイレクト.
func (self *BaseHandler) Redirect(url string) {
	context := self.GetContext()
	context.Writer.Header().Set("Cache-Control", "no-cache")
	context.Redirect(http.StatusMovedPermanently, url)
	self.setEnd()
}

// ログ.
func (self *BaseHandler) DebugLog(msg string, args ...interface{}) {
	lib_log.Debug(msg, args...)
}
func (self *BaseHandler) InfoLog(msg string, args ...interface{}) {
	lib_log.Info(msg, args...)
}
func (self *BaseHandler) ErrorLog(msg string, args ...interface{}) {
	lib_log.Error(msg, args...)
}

// 実行処理.
func Run(handler HandlerInterface) error {
	userAgent := handler.GetContext().Request.Header.Get("User-Agent")
	if userAgent == "Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)" {
		// ウィルスバスターのアクセス.IE6も死ぬけど仕方ない.
		handler.SetHttpStatus(http.StatusBadRequest)
		return handler.WriteString("No support")
	}
	// ローカルキャッシュ初期化.
	redis, err := handler.GetRedisClient(constant.RedisCache)
	if err != nil {
		return handler.ProcessError(err)
	}
	err = lib_redis.InitLocalCache(redis)
	if err != nil {
		return handler.ProcessError(err)
	}

	// Check Global Rate Limiter
	err = handler.CheckGlobalRateLimit()
	if err != nil {
		return handler.ProcessError(err)
	} else if handler.IsEnd() {
		return nil
	}

	// Check Rate Limiter
	err = handler.CheckRateLimit()
	if err != nil {
		return handler.ProcessError(err)
	} else if handler.IsEnd() {
		return nil
	}

	// ユーザー確認.
	err = handler.CheckUser()
	if err != nil {
		return handler.ProcessError(err)
	} else if handler.IsEnd() {
		return nil
	}
	// プロセス前の処理.
	err = handler.PreProcess()
	if err != nil {
		return handler.ProcessError(err)
	} else if handler.IsEnd() {
		return nil
	}
	// メイン処理.
	err = handler.Process()
	if err != nil {
		return handler.ProcessError(err)
	} else if !handler.IsEnd() {
		// レスポンスを返していない.
		return handler.ProcessError(lib_error.NewAppError(lib_error.DefaultErrorCode, "no response"))
	}
	return nil
}

// ルータのラッパ.
func Wrap(f func() HandlerInterface) gin.HandlerFunc {
	return func(context *gin.Context) {
		var handler HandlerInterface = nil
		defer func() {
			if err := recover(); err != nil {
				lib_log.Error("panic: %+v\n%s", err, lib_error.StackTrace())
				if handler != nil && !handler.IsEnd() {
					handler.ProcessError(fmt.Errorf("panic: %+v", err))
				}
			}
			if handler != nil {
				handler.Terminate()
			}
		}()
		// ハンドラ作成.
		handler = f()
		// セットアップ.
		err := handler.Setup(context)
		if err == nil {
			// 実行.
			err = Run(handler)
		}
		if err != nil {
			message := ""
			if constant.IsDebug() {
				message = err.Error()
			}
			lib_log.Error("internal server error: %+v", err)
			context.String(http.StatusInternalServerError, message)
		}
	}
}
