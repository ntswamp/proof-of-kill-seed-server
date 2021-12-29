package lib_session

import (
	"app/src/constant"
	lib_redis "app/src/lib/redis"
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
)

const DefaultSessionTTL int64 = 60 * 10 // 10 Minutes

type SessionUtil struct {
	prefix      string
	redisClient *lib_redis.Client
}

func NewSessionUtil(prefix string, redisClient *lib_redis.Client) (*SessionUtil, error) {
	if prefix == "" || redisClient == nil {
		return nil, fmt.Errorf("Invalid SessionUtil Settings")
	}
	return &SessionUtil{
		prefix:      prefix,
		redisClient: redisClient,
	}, nil
}

func (self *SessionUtil) newSessionId() (string, error) {
	key := securecookie.GenerateRandomKey(32)
	if key == nil {
		return "", fmt.Errorf("KeyGen Failed")
	}
	sessId := strings.TrimRight(base32.StdEncoding.EncodeToString(key), "=")
	return sessId, nil
}

func (self *SessionUtil) formSessionKey(sessionId string) string {
	return fmt.Sprintf("%s-SESS:%s", self.prefix, sessionId)
}

func (self *SessionUtil) saveSessionInfo(sessionId string, userId uint64, ttl int64) error {
	return self.redisClient.Multi(func(redisClient *lib_redis.Client, args ...interface{}) error {
		key := args[0].(string)
		data := args[1]
		ttl := args[2].(int64)
		err := redisClient.Set(key, data)
		if err != nil {
			return err
		}
		err = redisClient.Expire(key, ttl)
		if err != nil {
			return err
		}
		return nil
	}, self.formSessionKey(sessionId), userId, ttl)
}

func (self *SessionUtil) updateSessionStore(sessionId string, ttl int64) error {
	key := self.formSessionKey(sessionId)
	return self.redisClient.Expire(key, ttl)
}

func (self *SessionUtil) GetSessionInfo(sessionId string, dest interface{}) error {
	key := self.formSessionKey(sessionId)
	_, err := self.redisClient.Get(dest, key)
	if err != nil {
		return err
	}
	return nil
}

func (self *SessionUtil) MakeNewSession(context *gin.Context, sessIdKey string, ttl int64, userId uint64, rm bool) error {
	// Make session id
	sessionId, err := self.newSessionId()
	if err != nil {
		return err
	}

	// Save session info
	err = self.saveSessionInfo(sessionId, userId, ttl)
	if err != nil {
		return err
	}

	// Save Session
	session := sessions.Default(context)
	session.Set(constant.SESSION_ID_KEY, sessionId)
	session.Set(constant.SESSION_HALF_AUTHKEY, rm)
	session.Save()
	return nil
}

func (self *SessionUtil) RefreshSession(sessionId string, ttl int64) error {
	return self.updateSessionStore(sessionId, ttl)
}

func (self *SessionUtil) ClearSession(context *gin.Context, sessionId string) error {
	session := sessions.Default(context)
	session.Clear()
	session.Save()
	return self.updateSessionStore(sessionId, 0)
}
