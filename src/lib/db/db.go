package lib_db

import (
	lib_error "app/src/lib/error"
	lib_redis "app/src/lib/redis"
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// データベースに接続してクライアントを返す.
func Connect(name string, redis *lib_redis.Client) (*Client, error) {
	setting := GetConnectionSetting(name)
	if setting == nil {
		panic(fmt.Sprintf("%s does not exist in DB Settings.", name))
	}
	db, err := gorm.Open("postgres", setting.Address())
	if err != nil {
		log.Println(err)
		return nil, lib_error.WrapError(err)
	}
	// コールバックを登録.
	db.Callback().Create().Replace("gorm:create", CreateCallback)
	db.Callback().Update().Replace("gorm:update", UpdateCallback)
	return NewClient(db, redis), nil
}
