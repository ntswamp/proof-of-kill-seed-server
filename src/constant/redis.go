package constant

const RedisDb string = "db"
const RedisCache string = "cache" // セールサイトと同じにするべき.
const RedisSession string = "session"

var RedisDefault string = RedisDb

func GetRedisConnectionSettings() map[string]map[string]string {
	switch SERVER_TYPE {
	case "development":
		return map[string]map[string]string{
			RedisDb: {
				"Host": "127.0.0.1",
				"Port": "6379",
				"Db":   "0",
				"Pool": "100",
			},
			RedisCache: {
				"Host": "127.0.0.1",
				"Port": "6379",
				"Db":   "1",
				"Pool": "100",
			},
			RedisSession: {
				"Host": "127.0.0.1",
				"Port": "6379",
				"Db":   "2",
				"Pool": "100",
			},
		}
	case "production", "management":
		return map[string]map[string]string{
			RedisDb: {
				"Host": "token-rel-redis.a3ytn5.ng.0001.apne1.cache.amazonaws.com",
				"Port": "6379",
				"Db":   "0",
				"Pool": "100",
			},
			RedisCache: {
				"Host": "token-rel-redis.a3ytn5.ng.0001.apne1.cache.amazonaws.com",
				"Port": "6379",
				"Db":   "1",
				"Pool": "100",
			},
			RedisSession: {
				"Host": "token-rel-redis.a3ytn5.ng.0001.apne1.cache.amazonaws.com",
				"Port": "6379",
				"Db":   "2",
				"Pool": "100",
			},
		}
	default:
		return map[string]map[string]string{
			RedisDb: {
				"Host": "redis",
				"Port": "6379",
				"Db":   "0",
				"Pool": "10",
			},
			RedisCache: {
				"Host": "redis",
				"Port": "6379",
				"Db":   "1",
				"Pool": "10",
			},
			RedisSession: {
				"Host": "redis",
				"Port": "6379",
				"Db":   "2",
				"Pool": "10",
			},
		}
	}
}
