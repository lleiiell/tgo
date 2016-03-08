package tgo

import (
	"sync"
)

var (
	cacheConfigMux sync.Mutex

	cacheConfig *CacheConfig
)

type CacheConfig struct {
	Redis CacheRedisConfig
}

type CacheRedisConfig struct {
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

func getCacheConfigs() {

	if cacheConfig == nil || cacheConfig.Redis.Address == "" {

		cacheConfigMux.Lock()

		defer cacheConfigMux.Unlock()

		cacheConfig = &CacheConfig{}

		defaultCacheConfig := getDefaultCacheConfig()

		getConfigs("cache", cacheConfig, defaultCacheConfig)
	}
}

func getDefaultCacheConfig() *CacheConfig {
	return &CacheConfig{Redis: CacheRedisConfig{"172.172.177.15", 33062, "component", 1000, 1000, 1000, 10, 100, 180000}}
}

func GetCacheRedisConfig() *CacheRedisConfig {

	getCacheConfigs()

	redisConfig := cacheConfig.Redis

	if &redisConfig == nil {
		//log
		redisConfig = getDefaultCacheConfig().Redis
	}

	return &redisConfig
}
