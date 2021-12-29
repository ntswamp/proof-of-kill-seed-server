package lib_redis

import (
	lib_error "app/src/lib/error"
	"fmt"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
)

const NX string = "NX"
const XX string = "XX"

type Client struct {
	conn redis.Conn
}

func (self *Client) Terminate() {
	if self.conn != nil {
		self.conn.Close()
		self.conn = nil
	}
}

func NewClient(name string) (*Client, error) {
	conn, err := GetConnection(name)
	if err != nil {
		return nil, lib_error.WrapError(err)
	}
	return &Client{conn}, nil
}

//===========================================================================
// common.
func (self *Client) MakeKey(key, namespace string) string {
	// redisのクラスターを使う場合のキー作成.
	// multiとか使うなら必要.
	if len(namespace) == 0 {
		return key
	}
	return fmt.Sprintf("{%s}%s", namespace, key)
}

func (self *Client) Exists(key string) (bool, error) {
	v, err := self.conn.Do("EXISTS", key)
	return v == int64(1), lib_error.WrapError(err)
}

func (self *Client) Expire(key string, ttl int64) error {
	_, err := self.conn.Do("EXPIRE", key, ttl)
	return lib_error.WrapError(err)
}

func (self *Client) Ttl(key string) (int64, error) {
	i, err := self.conn.Do("TTL", key)
	if err != nil {
		return 0, lib_error.WrapError(err)
	}
	return i.(int64), nil
}

func (self *Client) Del(keys ...interface{}) error {
	_, err := self.conn.Do("DEL", keys...)
	return lib_error.WrapError(err)
}

func (self *Client) FlushDb(db int) error {
	_, err := self.conn.Do("FLUSHDB", db)
	return lib_error.WrapError(err)
}

func (self *Client) FlushAll() error {
	_, err := self.conn.Do("FLUSHALL")
	return lib_error.WrapError(err)
}

func (self *Client) Watch(keys ...interface{}) error {
	_, err := self.conn.Do("WATCH", keys...)
	return lib_error.WrapError(err)
}

func (self *Client) Multi(f func(*Client, ...interface{}) error, args ...interface{}) error {
	_, err := self.conn.Do("MULTI")
	if err != nil {
		return lib_error.WrapError(err)
	}
	defer func() {
		if err != nil {
			self.conn.Do("DISCARD")
		}
	}()
	err = f(self, args...)
	if err != nil {
		return lib_error.WrapError(err)
	}
	_, err = self.conn.Do("EXEC")
	return lib_error.WrapError(err)
}

//===========================================================================
// String.
func (self *Client) Get(dest interface{}, key string) (bool, error) {
	value, err := redis.String(self.conn.Do("GET", key))
	if err != nil {
		if 0 <= strings.Index(err.Error(), "nil returned") {
			return false, nil
		}
		return false, lib_error.WrapError(err)
	}
	err = unmarshal(value, dest)
	return true, lib_error.WrapError(err)
}

func (self *Client) MGet(dest interface{}, keys ...interface{}) error {
	values, err := redis.Strings(self.conn.Do("MGET", keys...))
	if err != nil {
		return lib_error.WrapError(err)
	}
	return unmarshalSlice(values, dest)
}

func (self *Client) Set(key string, value interface{}) error {
	redisValue, err := marshal(value)
	if err != nil {
		return lib_error.WrapError(err)
	}
	_, err = self.conn.Do("SET", key, redisValue)
	return lib_error.WrapError(err)
}

func (self *Client) SetNx(key string, value interface{}) error {
	redisValue, err := marshal(value)
	if err != nil {
		return lib_error.WrapError(err)
	}
	_, err = self.conn.Do("SETNX", key, redisValue)
	return lib_error.WrapError(err)
}

func (self *Client) Incr(key string) (interface{}, error) {
	v, err := self.conn.Do("INCR", key)
	if err != nil {
		return 0, lib_error.WrapError(err)
	}
	return v, lib_error.WrapError(err)
}

func (self *Client) IncrBy(key string, incr int64) (int64, error) {
	value, err := self.conn.Do("INCRBY", key, incr)
	if err != nil {
		return 0, lib_error.WrapError(err)
	}
	return value.(int64), nil
}

//===========================================================================
// List.
func (self *Client) LPush(key string, value interface{}) error {
	redisValue, err := marshal(value)
	if err != nil {
		return lib_error.WrapError(err)
	}
	_, err = self.conn.Do("LPUSH", key, redisValue)
	return lib_error.WrapError(err)
}

func (self *Client) LPop(key string) error {
	_, err := self.conn.Do("LPOP", key)
	return lib_error.WrapError(err)
}

func (self *Client) RPush(key string, value interface{}) error {
	redisValue, err := marshal(value)
	if err != nil {
		return lib_error.WrapError(err)
	}
	_, err = self.conn.Do("RPUSH", key, redisValue)
	return lib_error.WrapError(err)
}

func (self *Client) RPop(key string) error {
	_, err := self.conn.Do("RPOP", key)
	return lib_error.WrapError(err)
}

func (self *Client) LLen(key string) (int64, error) {
	value, err := self.conn.Do("LLEN", key)
	if err != nil {
		return 0, lib_error.WrapError(err)
	}
	return value.(int64), nil
}

func (self *Client) LRange(dest interface{}, key string, start, end int) error {
	values, err := redis.Strings(self.conn.Do("LRANGE", key, start, end))
	if err != nil {
		return lib_error.WrapError(err)
	}
	return unmarshalSlice(values, dest)
}

func (self *Client) LTrim(key string, start, end int) error {
	_, err := self.conn.Do("LTRIM", key, start, end)
	return lib_error.WrapError(err)
}

//===========================================================================
// Set.
func (self *Client) makeSetCommandArgs(key string, members []interface{}) ([]interface{}, error) {
	args := []interface{}{interface{}(key)}
	for _, member := range members {
		s, err := marshal(member)
		if err != nil {
			return args, lib_error.WrapError(err)
		}
		args = append(args, interface{}(s))
	}
	return args, nil
}

func (self *Client) SAdd(key string, members ...interface{}) error {
	args, err := self.makeSetCommandArgs(key, members)
	if err != nil {
		return lib_error.WrapError(err)
	}
	_, err = self.conn.Do("SADD", args...)
	return lib_error.WrapError(err)
}

func (self *Client) SRem(key string, members ...interface{}) error {
	args, err := self.makeSetCommandArgs(key, members)
	if err != nil {
		return lib_error.WrapError(err)
	}
	_, err = self.conn.Do("SREM", args...)
	return lib_error.WrapError(err)
}

func (self *Client) SMembers(dest interface{}, key string) error {
	values, err := redis.Strings(self.conn.Do("SMEMBERS", key))
	if err != nil {
		return lib_error.WrapError(err)
	}
	return unmarshalSlice(values, dest)
}

func (self *Client) SCard(key string) (int64, error) {
	value, err := self.conn.Do("SCARD", key)
	if err != nil {
		return 0, lib_error.WrapError(err)
	}
	return value.(int64), nil
}

func (self *Client) SIsMember(key string, member interface{}) (bool, error) {
	memberValue, err := marshal(member)
	if err != nil {
		return false, lib_error.WrapError(err)
	}
	v, err := self.conn.Do("SISMEMBER", key, memberValue)
	return v == int64(1), lib_error.WrapError(err)
}

func (self *Client) SRandMember(dest interface{}, key string, count int) error {
	values, err := redis.Strings(self.conn.Do("SRANDMEMBER", key, count))
	if err != nil {
		return lib_error.WrapError(err)
	}
	return unmarshalSlice(values, dest)
}

//===========================================================================
// Sorted Set.
func (self *Client) zadd(key string, members map[string]uint64, nxxx string, incr bool) error {
	args := []interface{}{interface{}(key)}

	// NX|XX.
	switch nxxx {
	case NX:
		fallthrough
	case XX:
		args = append(args, interface{}(nxxx))
	}

	// INCR.
	if incr {
		args = append(args, interface{}("INCR"))
	}

	// score member...
	for name, score := range members {
		args = append(args, interface{}(score), interface{}(name))
	}

	_, err := self.conn.Do("ZADD", args...)
	return lib_error.WrapError(err)
}

func (self *Client) ZAdd(key string, members map[string]uint64) error {
	return self.zadd(key, members, "", false)
}

func (self *Client) ZAddNX(key string, members map[string]uint64) error {
	return self.zadd(key, members, NX, false)
}

func (self *Client) ZAddXX(key string, members map[string]uint64) error {
	return self.zadd(key, members, XX, false)
}

func (self *Client) ZRem(key string, members ...interface{}) error {
	args := append([]interface{}{interface{}(key)}, members...)
	_, err := self.conn.Do("ZREM", args...)
	return lib_error.WrapError(err)
}

func (self *Client) ZRemRangeByScore(key string, min, max interface{}) error {
	_, err := self.conn.Do("ZREMRANGEBYSCORE", key, min, max)
	return lib_error.WrapError(err)
}

func (self *Client) ZRank(key string, member string) (int64, error) {
	v, err := self.conn.Do("ZRANK", key, member)
	return v.(int64), lib_error.WrapError(err)
}

func (self *Client) ZRevRank(key string, member string) (int64, error) {
	v, err := self.conn.Do("ZREVRANK", key, member)
	return v.(int64), lib_error.WrapError(err)
}

func (self *Client) ZIncrBy(key string, member string, score uint64) error {
	_, err := self.conn.Do("ZINCRBY", key, score, member)
	return lib_error.WrapError(err)
}

func (self *Client) ZRange(key string, start, end int64, reverse bool) ([]string, error) {
	var cmd string
	if reverse {
		cmd = "ZREVRANGE"
	} else {
		cmd = "ZRANGE"
	}
	dest, err := redis.Strings(self.conn.Do(cmd, key, start, end))
	return dest, lib_error.WrapError(err)
}

func (self *Client) ZRangeWithScores(key string, start, end int64, reverse bool) ([]*SortedSetMember, error) {
	var cmd string
	if reverse {
		cmd = "ZREVRANGE"
	} else {
		cmd = "ZRANGE"
	}
	arr, err := redis.Strings(self.conn.Do(cmd, key, start, end, "WITHSCORES"))
	if err != nil {
		return []*SortedSetMember{}, lib_error.WrapError(err)
	}
	return loadSortedSetMembers(arr), nil
}

func (self *Client) ZRangeByScore(key string, min, max interface{}, reverse bool) ([]string, error) {
	var cmd string
	if reverse {
		cmd = "ZREVRANGEBYSCORE"
	} else {
		cmd = "ZRANGEBYSCORE"
	}
	dest, err := redis.Strings(self.conn.Do(cmd, key, min, max))
	return dest, lib_error.WrapError(err)
}

func (self *Client) ZRangeByScoreWithScores(key string, min, max interface{}, reverse bool) ([]*SortedSetMember, error) {
	var cmd string
	if reverse {
		cmd = "ZREVRANGEBYSCORE"
	} else {
		cmd = "ZRANGEBYSCORE"
	}
	arr, err := redis.Strings(self.conn.Do(cmd, key, min, max, "WITHSCORES"))
	if err != nil {
		return []*SortedSetMember{}, lib_error.WrapError(err)
	}
	return loadSortedSetMembers(arr), nil
}

func (self *Client) ZCount(key string, min, max interface{}) (int64, error) {
	value, err := self.conn.Do("ZCOUNT", key, min, max)
	if err != nil {
		return 0, lib_error.WrapError(err)
	}
	return value.(int64), nil
}

func (self *Client) ZCard(key string) (int64, error) {
	value, err := self.conn.Do("ZCARD", key)
	if err != nil {
		return 0, lib_error.WrapError(err)
	}
	return value.(int64), nil
}

func (self *Client) ZScore(key, member string) (uint64, error) {
	s, err := redis.String(self.conn.Do("ZSCORE", key, member))
	if err != nil {
		return 0, lib_error.WrapError(err)
	}
	v, err := strconv.ParseUint(s, 10, 64)
	return v, lib_error.WrapError(err)
}

//===========================================================================
// HyperLogLog.

func (self *Client) PFAdd(key string, value interface{}) error {
	redisValue, err := marshal(value)
	if err != nil {
		return lib_error.WrapError(err)
	}
	_, err = self.conn.Do("PFADD", key, redisValue)
	return lib_error.WrapError(err)
}

func (self *Client) PFCount(key string) (uint64, error) {
	value, err := redis.Uint64(self.conn.Do("PFCOUNT", key))
	return value, lib_error.WrapError(err)
}

//===========================================================================
// Hash.

func (self *Client) HGet(dest interface{}, key, member string) error {
	value, err := redis.String(self.conn.Do("HGET", key, member))
	if err != nil {
		return lib_error.WrapError(err)
	}
	err = unmarshal(value, dest)
	return lib_error.WrapError(err)
}

func (self *Client) HExists(key, member string) (int, error) {
	value, err := redis.Int(self.conn.Do("HExists", key, member))
	return value, lib_error.WrapError(err)
}

func (self *Client) HMGet(dest interface{}, key string, members ...interface{}) error {
	args := append([]interface{}{interface{}(key)}, members...)
	values, err := redis.Strings(self.conn.Do("HMGET", args...))
	if err != nil {
		return lib_error.WrapError(err)
	}
	return unmarshalSlice(values, dest)
}

func (self *Client) HSet(key string, nameAndValues ...interface{}) error {
	args := []interface{}{interface{}(key)}
	for i := 0; i < len(nameAndValues); i += 2 {
		name := nameAndValues[i]
		redisValue, err := marshal(nameAndValues[i+1])
		if err != nil {
			return lib_error.WrapError(err)
		}
		args = append(args, name, redisValue)
	}
	_, err := self.conn.Do("HMSET", args...)
	return lib_error.WrapError(err)
}

func (self *Client) HDel(key string, members ...interface{}) error {
	args := append([]interface{}{interface{}(key)}, members...)
	_, err := self.conn.Do("HDEL", args...)
	return lib_error.WrapError(err)
}

func (self *Client) HIncrBy(key string, member string, incr int64) (int64, error) {
	value, err := self.conn.Do("HINCRBY", key, member, incr)
	if err != nil {
		return 0, lib_error.WrapError(err)
	}
	return value.(int64), nil
}

func (self *Client) HLen(key string) (int64, error) {
	value, err := self.conn.Do("HLEN", key)
	if err != nil {
		return 0, lib_error.WrapError(err)
	}
	return value.(int64), nil
}

func (self *Client) HGetAll(dest interface{}, key string) error {
	values, err := redis.Strings(self.conn.Do("HGETALL", key))
	if err != nil {
		return lib_error.WrapError(err)
	}
	return unmarshalMap(values, dest)
}

//===========================================================================
// Geospatial.

func (self *Client) GeoAdd(key string, args ...interface{}) error {
	var err error = nil
	tmp := make([]interface{}, len(args)+1)
	tmp[0] = key
	for i := 0; i < len(args); i += 3 {
		tmp[i+3], err = marshal(args[i+2])
		if err != nil {
			return lib_error.WrapError(err)
		}
		tmp[i+1] = args[i]
		tmp[i+2] = args[i+1]
	}
	_, err = self.conn.Do("GEOADD", tmp...)
	return lib_error.WrapError(err)
}
func (self *Client) GeoDel(key string, args ...interface{}) error {
	var err error = nil
	members := make([]interface{}, len(args))
	for i, v := range args {
		members[i], err = marshal(v)
		if err != nil {
			return lib_error.WrapError(err)
		}
	}
	return self.ZRem(key, members...)
}
func (self *Client) GeoPos(key string, members ...interface{}) ([][]float64, error) {
	var err error = nil
	tmp := make([]interface{}, len(members)+1)
	tmp[0] = key
	for i, member := range members {
		tmp[i+1], err = marshal(member)
		if err != nil {
			return nil, lib_error.WrapError(err)
		}
	}
	result, err := self.conn.Do("GEOPOS", tmp...)
	if err != nil {
		return nil, lib_error.WrapError(err)
	}
	arr := result.([]interface{})
	datas := make([][]float64, len(arr))
	for i, v := range arr {
		if v == nil {
			continue
		}
		lonlat := v.([]interface{})
		datas[i] = make([]float64, len(lonlat))
		for j, loc := range lonlat {
			locf, err := strconv.ParseFloat(string(loc.([]byte)), 64)
			if err != nil {
				return nil, lib_error.WrapError(err)
			}
			datas[i][j] = locf
		}
	}
	return datas, nil
}
func (self *Client) GeoRadius(dest interface{}, key string, lon, lat, radius float64, unit string) error {
	values, err := redis.Strings(self.conn.Do("GEORADIUS", key, lon, lat, radius, unit))
	if err != nil {
		return lib_error.WrapError(err)
	}
	return unmarshalSlice(values, dest)
}
