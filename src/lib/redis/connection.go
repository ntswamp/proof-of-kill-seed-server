package lib_redis

import (
	lib_error "app/src/lib/error"
	"fmt"

	"github.com/garyburd/redigo/redis"
)

var pools = map[string]*redis.Pool{}

func newPool(name string) (*redis.Pool, error) {

	setting := GetConnectionSetting(name)
	if setting == nil {
		return nil, lib_error.NewAppErrorWithStackTrace(lib_error.DefaultErrorCode, "Illegal redis db name:%s", name)
	}

	pool := redis.NewPool(func() (redis.Conn, error) {
		connectTo := fmt.Sprintf("%s:%s", setting.Host, setting.Port)
		opt := redis.DialDatabase(setting.Db)
		conn, err := redis.Dial("tcp", connectTo, opt)
		return conn, err
	}, setting.Pool)

	return pool, nil

}

func GetConnection(name string) (redis.Conn, error) {
	var err error = nil
	pool, ok := pools[name]
	if !ok || pool == nil {
		pool, err = newPool(name)
		if err != nil {
			return nil, err
		}
	}
	conn := pool.Get()
	return conn, nil
}
