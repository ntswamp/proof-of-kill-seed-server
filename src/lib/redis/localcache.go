package lib_redis

import (
	"app/src/constant"
	"reflect"
	"time"
)

/**
 * 各フロントエンド毎にデータをキャッシュする.
 * マスターデータ等の不変な値はこっちに入れる.
 */
type LocalCache struct {
	data map[string]string
}

/**
 * ローカルのキャッシュからデータを取得.
 */
func (self *LocalCache) Get(key string, dest interface{}) bool {
	s, ok := self.data[key]
	if !ok {
		return false
	}
	err := Unmarshal(s, dest)
	if err != nil {
		return false
	}
	return true
}

/**
 * ローカルのキャッシュからデータをまとめて取得.
 */
func (self *LocalCache) GetMany(keys []string, dest interface{}) {
	if len(keys) == 0 {
		return
	}
	destV := reflect.ValueOf(dest)
	elem := destV.Elem()
	sl := reflect.MakeSlice(elem.Type(), 0, 0)
	valType := sl.Type().Elem()
	for _, key := range keys {
		var itemV reflect.Value
		if valType.Kind() == reflect.Ptr {
			itemV = reflect.New(valType.Elem())
			if !self.Get(key, itemV.Interface()) {
				// 無かった.
				continue
			}
		} else {
			itemV = reflect.New(valType).Elem()
			if !self.Get(key, itemV.Addr().Interface()) {
				// 無かった.
				continue
			}
		}
		sl = reflect.Append(sl, itemV)
	}
	elem.Set(sl)
}

/**
 * ローカルのキャッシュからデータをまとめてmapで取得.
 * sampleにはmapの要素の型のサンプルを渡す.
 */
func (self *LocalCache) GetManyMap(keys []string, dest interface{}) {
	if len(keys) == 0 {
		return
	}
	destV := reflect.ValueOf(dest)
	elem := destV.Elem()
	mp := reflect.MakeMap(elem.Type())
	valType := mp.Type().Elem()
	for _, key := range keys {
		var itemV reflect.Value
		if valType.Kind() == reflect.Ptr {
			itemV = reflect.New(valType.Elem())
			if !self.Get(key, itemV.Interface()) {
				// 無かった.
				continue
			}
		} else {
			itemV = reflect.New(valType).Elem()
			if !self.Get(key, itemV.Addr().Interface()) {
				// 無かった.
				continue
			}
		}
		mp.SetMapIndex(reflect.ValueOf(key), itemV)
	}
	elem.Set(mp)
}

/**
 * データをローカルのキャッシュに保存.
 */
func (self *LocalCache) Set(key string, value interface{}) error {
	s, err := Marshal(value)
	if err != nil {
		return err
	}
	self.data[key] = s
	return nil
}

var localcache *LocalCache = nil
var localcacheTime int64 = 0

func InitLocalCache(client *Client) error {
	var err error = nil
	if client == nil {
		client, err = NewClient(constant.RedisCache)
		if err != nil {
			return err
		}
	}
	var t int64 = 0
	_, err = client.Get(&t, "localcacheTime")
	if err != nil {
		return err
	} else if t == 0 {
		// 設定がない.
		t = time.Now().UnixNano()
		err = client.SetNx("localcacheTime", t)
		if err != nil {
			return err
		}
	}
	if localcacheTime < t {
		// ローカルのキャッシュを破棄.
		localcache = nil
		localcacheTime = t
	}
	return nil
}
func DeleteLocalCache(client *Client) error {
	var err error = nil
	if client == nil {
		client, err = NewClient(constant.RedisCache)
		if err != nil {
			return err
		}
		defer client.Terminate()
	}
	err = client.Del("localcacheTime")
	return err
}
func newLocalCache() *LocalCache {
	return &LocalCache{
		data: map[string]string{},
	}
}
func getLocalCache() *LocalCache {
	if localcache == nil {
		localcache = newLocalCache()
	}
	return localcache
}

func GetLocalCache(key string, dest interface{}) bool {
	obj := getLocalCache()
	return obj.Get(key, dest)
}

func GetLocalCacheMany(keys []string, dest interface{}) {
	obj := getLocalCache()
	obj.GetMany(keys, dest)
}

func GetLocalCacheManyMap(keys []string, dest interface{}) {
	obj := getLocalCache()
	obj.GetManyMap(keys, dest)
}

func SetLocalCache(key string, value interface{}) error {
	obj := getLocalCache()
	return obj.Set(key, value)
}
