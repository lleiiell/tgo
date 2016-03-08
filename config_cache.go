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
	Address         string
	Port            int
	Prefix          string
	ReadTimeout     int
	WriteTimeout    int
	ConnectTimeout  int
	PoolMaxIdle     int
	PoolMaxActive   int
	PoolIdleTimeout int
}

func configConfigGet() {

	if cacheConfig == nil || cacheConfig.Redis.Address == "" {

		cacheConfigMux.Lock()

		defer cacheConfigMux.Unlock()

		cacheConfig = &ConfigCache{}

		defaultCacheConfig := configCacheGetDefault()

		configGet("cache", cacheConfig, defaultCacheConfig)
	}
}

func configCacheGetDefault() *ConfigCache {
	return &ConfigCache{Redis: ConfigCacheRedis{"172.172.177.15", 33062, "component", 1000, 1000, 1000, 10, 100, 180000}}
}

func ConfigCacheGetRedis() *ConfigCacheRedis {

	configConfigGet()

	redisConfig := cacheConfig.Redis

	if &redisConfig == nil {
		//log
		redisConfig = configCacheGetDefault().Redis
	}

	return &redisConfig
}
