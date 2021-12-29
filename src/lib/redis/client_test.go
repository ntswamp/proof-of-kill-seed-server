package lib_redis

import (
	"app/src/constant"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 文字列型のテスト.
func TestStringType(t *testing.T) {

	client, err := NewClient(constant.RedisDb)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		client.Del("test", "test2", "test3")
		client.Terminate()
	}()

	t.Log("del")
	err = client.Del("test", "test2", "test3")
	if err != nil {
		t.Fatal(err)
	}

	// set.
	t.Log("set")
	err = client.Set("test", 1.5)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Set("test3", 2.5)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("exists")
	isExists, err := client.Exists("test")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, isExists, true)

	t.Log("get")
	dest := float64(0)
	_, err = client.Get(&dest, "test")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, dest, float64(1.5))

	isExists, err = client.Get(&dest, "not_found")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, isExists, false)

	t.Log("mget")
	mget := []float64{}
	err = client.MGet(&mget, "test", "test3")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, mget[0], float64(1.5))

	t.Log("expire")
	err = client.Expire("test", 86400)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("ttl")
	ttl, err := client.Ttl("test")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ttl)

	t.Log("incrby")
	incr, err := client.IncrBy("test2", 100)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, incr, int64(100))

	t.Log("multi success")
	err = client.Multi(func(mclient *Client, args ...interface{}) error {
		mclient.Set("test2", args[0])
		return nil
	}, 200)
	if err != nil {
		t.Fatal(err)
	}
	multi := 0
	_, err = client.Get(&multi, "test2")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, multi, 200)

	t.Log("multi failure")
	client.Multi(func(mclient *Client, args ...interface{}) error {
		mclient.Set("test2", args[0])
		return errors.New("HogeHoge")
	}, 300)
	_, err = client.Get(&multi, "test2")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, multi, 200)

	t.Log("finish")
}

// リスト型のテスト.
func TestListType(t *testing.T) {
	client, err := NewClient(constant.RedisDb)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		client.Del("test")
		client.Terminate()
	}()

	t.Log("del")
	err = client.Del("test")
	if err != nil {
		t.Fatal(err)
	}

	// lpush.
	t.Log("lpush")
	err = client.LPush("test", 10)
	if err != nil {
		t.Fatal(err)
	}

	// rpush.
	t.Log("rpush")
	pValues := []int{100, 200, 300}
	for _, v := range pValues {
		err = client.RPush("test", v)
		if err != nil {
			t.Fatal(err)
		}
	}

	// llen.
	t.Log("llen")
	llen, err := client.LLen("test")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, llen, int64(4))

	// lrange.
	t.Log("lrange")
	values := []float64{}
	err = client.LRange(&values, "test", 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, values[0], float64(10))

	// lpop.
	t.Log("lpop")
	err = client.LPop("test")
	if err != nil {
		t.Fatal(err)
	}
	err = client.LRange(&values, "test", 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, values[0], float64(100))

	// rpop.
	t.Log("rpop")
	err = client.RPop("test")
	if err != nil {
		t.Fatal(err)
	}
	err = client.LRange(&values, "test", 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, values[0], float64(100))
	llen, err = client.LLen("test")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, llen, int64(2))

	// ltrim.
	t.Log("ltrim")
	err = client.LTrim("test", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	llen, err = client.LLen("test")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, llen, int64(1))
}

// セット型のテスト.
func TestSetType(t *testing.T) {
	client, err := NewClient(constant.RedisDb)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		client.Del("test")
		client.Terminate()
	}()

	t.Log("del")
	err = client.Del("test")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("sadd")
	err = client.SAdd("test", "a", "b", "c", "d")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("srem")
	err = client.SRem("test", "b", "d")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("smembers")
	members := []string{}
	err = client.SMembers(&members, "test")
	if err != nil {
		t.Fatal(err)
	} else if len(members) != 2 {
		t.Error("想定外の長さです")
	}

	t.Log("scard")
	scard, err := client.SCard("test")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, scard, int64(2))

	t.Log("sismember")
	isMember, err := client.SIsMember("test", "c")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, isMember, true)

	t.Log("srandmember")
	err = client.SRandMember(&members, "test", 1)
	if err != nil {
		t.Fatal(err)
	} else if len(members) != 1 {
		t.Error("想定外の長さです")
	} else if members[0] != "a" && members[0] != "c" {
		t.Error("結果が想定外です")
	}
}

// セット型のテスト.
func TestSortedSetType(t *testing.T) {
	client, err := NewClient(constant.RedisDb)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		client.Del("test")
		client.Terminate()
	}()

	t.Log("del")
	err = client.Del("test")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("zadd")
	members := map[string]uint64{
		"a": 2,
	}
	err = client.ZAdd("test", members)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("zadd nx")
	members = map[string]uint64{
		"a": 3,
		"b": 10,
		"c": 1,
	}
	err = client.ZAddNX("test", members)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("zadd xx")
	members = map[string]uint64{
		"b": 100,
		"d": 1,
	}
	err = client.ZAddXX("test", members)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("zrank")
	rank, err := client.ZRank("test", "b")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, rank, int64(2))

	t.Log("zrevrank")
	rank, err = client.ZRevRank("test", "b")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, rank, int64(0))

	t.Log("zincrby")
	err = client.ZIncrBy("test", "c", 2)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("zrange")
	zrange, err := client.ZRange("test", 0, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, zrange, []string{"a", "c", "b"})

	t.Log("zrevrange")
	zrange, err = client.ZRange("test", 0, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, zrange, []string{"b", "c", "a"})

	t.Log("zrange withscores")
	sortedSetMembers, err := client.ZRangeWithScores("test", 0, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, sortedSetMembers[0].Name, "a")
	assert.Equal(t, sortedSetMembers[0].Score, uint64(2))

	t.Log("zrevrange withscores")
	sortedSetMembers, err = client.ZRangeWithScores("test", 0, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, sortedSetMembers[0].Name, "b")
	assert.Equal(t, sortedSetMembers[0].Score, uint64(100))

	t.Log("zrangebyscore")
	zrange, err = client.ZRangeByScore("test", 3, "+inf", false)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, zrange, []string{"c", "b"})

	t.Log("zrevrangebyscore")
	zrange, err = client.ZRangeByScore("test", 3, "-inf", true)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, zrange, []string{"c", "a"})

	t.Log("zrangebyscore withscores")
	sortedSetMembers, err = client.ZRangeByScoreWithScores("test", 3, "+inf", false)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, sortedSetMembers[0].Name, "c")
	assert.Equal(t, sortedSetMembers[0].Score, uint64(3))

	t.Log("zrevrangebyscore withscores")
	sortedSetMembers, err = client.ZRangeByScoreWithScores("test", 3, "-inf", true)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, sortedSetMembers[1].Name, "a")
	assert.Equal(t, sortedSetMembers[1].Score, uint64(2))

	t.Log("zcount")
	zcount, err := client.ZCount("test", 3, "+inf")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, zcount, int64(2))

	t.Log("zcard")
	zcard, err := client.ZCard("test")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, zcard, int64(3))

	t.Log("zcard")
	zscore, err := client.ZScore("test", "b")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, zscore, uint64(100))

}

// ハッシュ型のテスト.
func TestHashType(t *testing.T) {
	client, err := NewClient(constant.RedisDb)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		client.Del("test")
		client.Terminate()
	}()

	t.Log("del")
	err = client.Del("test")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("hset")
	testMap := map[string]int{"test": 100}
	err = client.HSet("test", "a", 1.5, "mapp", testMap)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("hget")
	hget := float64(0)
	err = client.HGet(&hget, "test", "a")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("hmget")
	hmget := []map[string]int{}
	err = client.HMGet(&hmget, "test", "mapp", "mapq")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, hmget[0]["test"], 100)
	assert.Equal(t, len(hmget[1]), 0)

	t.Log("hdel")
	err = client.HDel("test", "mapp", "a", "mapq")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("hlen")
	hlen, err := client.HLen("test")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, hlen, int64(0))

	t.Log("hincrby")
	hincrby, err := client.HIncrBy("test", "b", 3)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, hincrby, int64(3))

	t.Log("hgetall")
	hgetall := map[string]int{}
	err = client.HGetAll(&hgetall, "test")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, hgetall["b"], 3)
}

// 地理空間型のテスト.
func TestGeospatial(t *testing.T) {
	client, err := NewClient(constant.RedisDb)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		client.Del("test")
		client.Terminate()
	}()
	t.Log("geoadd")
	err = client.GeoAdd("test", 13.361389, 38.115556, "Palermo", 15.087269, 37.502669, "Catania")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("geopos")
	locations, err := client.GeoPos("test", "Palermo", "Catania", "Nothing")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(locations)
	assert.Nil(t, locations[2])
	t.Log("geopos")
	members := []string{}
	err = client.GeoRadius(&members, "test", 13.361389, 38.115556, 20, "km")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(members)
}
