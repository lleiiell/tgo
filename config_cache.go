package tgo

import (
	"sync"
)

var (
	cacheConfigMux sync.Mutex

	cacheConfig *ConfigCache
)

type ConfigCache struct {
	Redis ConfigCacheRedis
	RedisP ConfigCacheRedis // 持久化Redis
}

type ConfigCacheRedis struct {
	Address         []string
	Prefix          string
	Expire          int
	ReadTimeout     int
	WriteTimeout    int
	ConnectTimeout  int
	PoolMaxIdle     int
	PoolMaxActive   int
	PoolIdleTimeout int
	PoolMinActive   int
}

func configCacheGet() {

	if cacheConfig == nil || len(cacheConfig.Redis.Address) == 0 {

		cacheConfigMux.Lock()

		defer cacheConfigMux.Unlock()

		if cacheConfig == nil || len(cacheConfig.Redis.Address) == 0 {
			cacheConfig = &ConfigCache{}

			defaultCacheConfig := configCacheGetDefault()

			configGet("cache", cacheConfig, defaultCacheConfig)
		}
	}
}

func configCacheClear() {
	cacheConfigMux.Lock()

	defer cacheConfigMux.Unlock()

	cacheConfig = nil
}

func configCacheGetDefault() *ConfigCache {
	return &ConfigCache{Redis: ConfigCacheRedis{[]string{"ip:port"}, "prefix", 604800, 1000, 1000, 1000, 10, 100, 180000, 2}, RedisP: ConfigCacheRedis{[]string{"ip:port"}, "prefix", 604800, 1000, 1000, 1000, 10, 100, 180000, 2}}
}

func ConfigCacheGetRedis() *ConfigCacheRedis {

	configCacheGet()

	redisConfig := cacheConfig.Redis

	if &redisConfig == nil {
		//log
		redisConfig = configCacheGetDefault().Redis
	}

	return &redisConfig
}

func ConfigCacheGetRedisWithConn(persistent bool) *ConfigCacheRedis {

	configCacheGet()

    var redisConfig ConfigCacheRedis
    if !persistent {
        redisConfig = cacheConfig.Redis
    } else {
        redisConfig = cacheConfig.RedisP
    }

	return &redisConfig
}
