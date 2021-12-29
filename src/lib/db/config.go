package lib_db

import (
	"app/src/constant"
	"fmt"
)

type connectionSetting struct {
	Host     string
	Port     string
	User     string
	Database string
	Password string
}

// "host=myhost port=myport user=gorm dbname=gorm password=mypassword"
func (cs *connectionSetting) Address() string {
	connection := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s TimeZone=Asia/Tokyo", cs.Host, cs.Port, cs.User, cs.Database, cs.Password)
	if constant.IsDebug() {
		connection = fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable TimeZone=Asia/Tokyo", cs.Host, cs.Port, cs.User, cs.Database, cs.Password)
	}
	return connection
}

// データベースの接続用マッピング情報
var settings map[string]*connectionSetting = map[string]*connectionSetting{}

func initSettings() {
	dbConnectionSettings := constant.GetDbConnectionSettings()
	settings = make(map[string]*connectionSetting, len(dbConnectionSettings))
	for name, paramMap := range dbConnectionSettings {
		host, _ := paramMap["Host"]
		if len(host) == 0 {
			host = "127.0.0.1"
		}
		port, _ := paramMap["Port"]
		if len(port) == 0 {
			port = "5432"
		}
		database := paramMap["Database"]
		user := paramMap["User"]
		password := paramMap["Password"]
		settings[name] = &connectionSetting{
			Host:     host,
			Port:     port,
			Database: database,
			User:     user,
			Password: password,
		}
	}
}

func GetConnectionSetting(name string) *connectionSetting {
	dbConnectionSettings := constant.GetDbConnectionSettings()
	if len(settings) != len(dbConnectionSettings) {
		// 未ロード.
		initSettings()
	}
	if _, ok := settings[name]; !ok {
		// 無い.
		return nil
	}
	return settings[name]
}
