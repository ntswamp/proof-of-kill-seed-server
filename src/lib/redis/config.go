package lib_redis

import (
	"app/src/constant"
	"strconv"
)

type connectionSetting struct {
	Host string
	Port string
	Db   int
	Pool int
}

// データベースの接続用マッピング情報
var settings map[string]*connectionSetting = map[string]*connectionSetting{}

func initSettings() {
	redisConnectionSettings := constant.GetRedisConnectionSettings()
	settings = make(map[string]*connectionSetting, len(redisConnectionSettings))
	for name, paramMap := range redisConnectionSettings {
		host, _ := paramMap["Host"]
		if len(host) == 0 {
			host = "127.0.0.1"
		}
		port, _ := paramMap["Port"]
		if len(port) == 0 {
			port = "6379"
		}
		db := 0
		if _, ok := paramMap["Db"]; ok {
			db, _ = strconv.Atoi(paramMap["Db"])
		}
		pool := 0
		if _, ok := paramMap["Pool"]; ok {
			pool, _ = strconv.Atoi(paramMap["Pool"])
		}
		if pool < 1 {
			pool = 1
		}
		settings[name] = &connectionSetting{
			Host: host,
			Port: port,
			Db:   db,
			Pool: pool,
		}
	}
}

func GetConnectionSetting(name string) *connectionSetting {
	redisConnectionSettings := constant.GetRedisConnectionSettings()
	if len(settings) != len(redisConnectionSettings) {
		// 未ロード.
		initSettings()
	}
	if _, ok := settings[name]; !ok {
		// 無い.
		return nil
	}
	return settings[name]
}
