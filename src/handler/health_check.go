package handler

import (
	"app/src/constant"
	lib_db "app/src/lib/db"
	lib_error "app/src/lib/error"
	lib_redis "app/src/lib/redis"
	"net/http"
)

type HealthCheckHandler struct {
	BaseHandler
}

func (self *HealthCheckHandler) ProcessError(err error) error {
	self.SetHttpStatus(http.StatusInternalServerError)
	lib_error.WrapError(err)
	return self.WriteString(err.Error())
}

func (self *HealthCheckHandler) Process() error {
	context := self.GetContext()
	ip, _ := context.GetQuery("ip")
	if ip == "1" {
		self.SetHttpStatus(http.StatusOK)
		return self.WriteString(context.ClientIP())
	}
	// DB.
	dbClient, err := self.GetDbClient(constant.DbIdolverse)
	if err != nil {
		return err
	}
	err = dbClient.GetDB().DB().Ping()
	if err != nil {
		return err
	}

	// Redis.
	redisClient, err := self.GetRedisClient(constant.RedisCache)
	if err != nil {
		return err
	}
	err = redisClient.Set("health_check", 1)
	if err != nil {
		return err
	}
	v := 0
	_, err = redisClient.Get(&v, "health_check")
	if err != nil {
		return err
	}

	//Db with Redis
	redisCache, err := lib_redis.NewClient(constant.RedisCache)
	if err != nil {
		return err
	}
	defer redisCache.Terminate()
	drClient, err := lib_db.Connect(constant.DbIdolverse, redisCache)
	if err != nil {
		return err
	}
	defer drClient.Close()
	err = drClient.GetDB().DB().Ping()
	if err != nil {
		return err
	}

	self.SetHttpStatus(http.StatusOK)
	return self.WriteString("ok")
}

func HealthCheck() HandlerInterface {
	return &HealthCheckHandler{}
}
