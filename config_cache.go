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
	return &ConfigCache{Redis: ConfigCacheRedis{[]string{"172.172.177.15:33062"}, "component", 604800, 1000, 1000, 1000, 10, 100, 180000, 2}}
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
