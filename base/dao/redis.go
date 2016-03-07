package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/tonyj/tgo/configs"
	"reflect"
	"strings"
	"sync"
	"time"
	"util"
)

type Redis struct {
	KeyName string
}

var (
	pool         *redis.Pool
	redisPoolMux sync.Mutex
)

func InitRedis() (redis.Conn, error) {

	cacheConfig := configs.GetCacheRedisConfig()

	conn, err := redis.DialTimeout("tcp", fmt.Sprintf("%s:%d", cacheConfig.Address, cacheConfig.Port), time.Duration(cacheConfig.ConnectTimeout)*time.Millisecond, time.Duration(cacheConfig.ReadTimeout)*time.Millisecond, time.Duration(cacheConfig.WriteTimeout)*time.Millisecond)

	if err != nil {

		util.LogErrorf("open redis error: %s", err.Error())

	}

	return conn, err
}

func InitRedisPool() (redis.Conn, error) {

	if pool == nil {

		redisPoolMux.Lock()

		defer redisPoolMux.Unlock()

		cacheConfig := configs.GetCacheRedisConfig()

		address := fmt.Sprintf("%s:%d", cacheConfig.Address, cacheConfig.Port)

		pool = &redis.Pool{
			// 从配置文件获取maxidle以及maxactive，取不到则用后面的默认值
			MaxIdle:     cacheConfig.PoolMaxIdle,
			MaxActive:   cacheConfig.PoolMaxActive,
			IdleTimeout: time.Duration(cacheConfig.PoolIdleTimeout) * time.Millisecond,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", address)
				if err != nil {
					util.LogErrorf("open redis pool error: %s", err.Error())

					return nil, err
				}
				return c, nil
			},
		}
	}
	if pool != nil {
		return pool.Get(), nil
	}

	util.LogError("redis pool is null")

	return nil, errors.New("redis pool is null")
}

func (b *Redis) getKey(key string) string {

	cacheConfig := configs.GetCacheRedisConfig()

	prefixRedis := cacheConfig.Prefix

	if strings.Trim(key, " ") == "" {
		return fmt.Sprintf("%s:%s", prefixRedis, b.KeyName)
	}
	return fmt.Sprintf("%s:%s:%s", prefixRedis, b.KeyName, key)
}

func (b *Redis) Set(key string, value interface{}) bool {

	redisClient, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer redisClient.Close()
	data, errJson := json.Marshal(value)

	if errJson != nil {
		util.LogErrorf("redis Set marshal data to json:%s", errJson.Error())
		return false
	}
	_, errDo := redisClient.Do("SET", b.getKey(key), data)

	if errDo != nil {
		util.LogErrorf("run redis command Set failed:%s", errDo.Error())
		return false
	}
	return true
}

func (b *Redis) Get(key string, data interface{}) bool {

	redisClient, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer redisClient.Close()

	result, errDo := redisClient.Do("GET", b.getKey(key))

	if errDo != nil {
		util.LogErrorf("run redis command GET failed:%s", errDo.Error())
		return false
	}
	if result == nil {
		//util.LogInfof("run GET failed:%s", key)

		return false
	}

	if reflect.TypeOf(result).Kind() == reflect.Slice {

		byteResult := (result.([]byte))
		strResult := string(byteResult)

		if strResult == "[]" {
			return true
		}
	}

	errorJson := json.Unmarshal(result.([]byte), data)

	if errorJson != nil {
		util.LogErrorf("redis GET data  unmarshal failed:%s", errorJson.Error())
		return false
	}
	return true
}

func (b *Redis) Incr(key string) (interface{}, bool) {

	redisClient, err := InitRedisPool()

	if err != nil {
		return 0, false
	}

	defer redisClient.Close()

	data, errDo := redisClient.Do("INCR", b.getKey(key))

	if errDo != nil {
		util.LogErrorf("run redis INCR command failed:%s", errDo.Error())

		return 0, false
	}
	return data, true
}

//hash start
func (b *Redis) HIncrby(key string, field string, value int) (int, bool) {

	redisClient, err := InitRedisPool()

	if err != nil {
		return 0, false
	}

	defer redisClient.Close()

	data, errDo := redisClient.Do("HINCRBY", b.getKey(key), field, value)

	if errDo != nil {
		util.LogErrorf("run redis HINCRBY command failed:%s", errDo.Error())

		return 0, false
	}

	count, result := data.(int64)

	if !result {

		util.LogErrorf("get HINCRBY command result failed:%v ,is %v", data, reflect.TypeOf(data))

		return 0, false
	}
	return int(count), true
}

func (b *Redis) HGet(key string, field string, value interface{}) bool {

	redisClient, err := InitRedisPool()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	result, errDo := redisClient.Do("HGet", b.getKey(key), field)

	if errDo != nil {
		util.LogErrorf("run redis HGet command failed:%s", errDo.Error())

		return false
	}

	if result == nil {

		return false
	}

	errorJson := json.Unmarshal(result.([]byte), value)

	if errorJson != nil {

		util.LogErrorf("get HGet command result failed:%s", errorJson.Error())

		return false
	}

	return true
}

func (b *Redis) HSet(key string, field string, value interface{}) bool {

	redisClient, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer redisClient.Close()
	data, errJson := json.Marshal(value)

	if errJson != nil {
		util.LogErrorf("redis HSET marshal data to json:%s", errJson.Error())
		return false
	}
	_, errDo := redisClient.Do("HSET", b.getKey(key), field, data)

	if errDo != nil {
		util.LogErrorf("run redis HSET command failed:%s", errDo.Error())
		return false
	}
	return true
}
func (b *Redis) HMSet(key string, datas ...interface{}) bool {

	redisClient, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer redisClient.Close()

	_, errDo := redisClient.Do("HMSET", b.getKey(key), datas)

	if errDo != nil {
		util.LogErrorf("run redis HMSET command failed:%s", errDo.Error())
		return false
	}
	return true
}

func (b *Redis) HLen(key string, data *int) bool {
	redisClient, err := InitRedisPool()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	resultData, errDo := redisClient.Do("HLEN", key)

	if errDo != nil {
		util.LogErrorf("run redis ZADD command failed:%s", errDo.Error())
		return false
	}

	lenth, resultConv := resultData.(int)

	if !resultConv {
		util.LogErrorf("redis data convert to int failed:%v", resultConv)

	}

	data = &lenth

	return resultConv
}

func (b *Redis) HDel(key string, field string) bool {

	redisClient, err := InitRedisPool()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	_, errDo := redisClient.Do("HDel", b.getKey(key), field)

	if errDo != nil {

		util.LogErrorf("run redis HDel command failed:%s", errDo.Error())

		return false
	}

	return true
}

// hash end

// sorted set start
func (b *Redis) ZAdd(key string, score int, data interface{}) bool {
	redisClient, err := InitRedisPool()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	_, errDo := redisClient.Do("ZADD", key, score, data)

	if errDo != nil {
		util.LogErrorf("run redis ZADD command failed:%s", errDo.Error())
		return false
	}
	return true
}
func (b *Redis) ZGet(key string, sort bool, start int, end int, data interface{}) bool {
	redisClient, err := InitRedisPool()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	var command string
	if sort {
		command = "ZRANGE"
	} else {
		command = "ZREVRANGE"
	}

	result, errDo := redisClient.Do(command, key, start, end)

	if errDo != nil {
		util.LogErrorf("run redis command ZREVRANGE failed:%s", errDo.Error())
		return false
	}

	if result == nil {
		return false
	}

	errorJson := json.Unmarshal(result.([]byte), data)

	if errorJson != nil {
		util.LogErrorf("redis data unmarshal failed:%s", errorJson.Error())
		return false
	}
	return true
}

//list start

func (b *Redis) LRange(key string, start int, end int, data interface{}) bool {
	redisClient, err := InitRedisPool()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	result, errDo := redisClient.Do("LRANGE", key, start, end)

	if errDo != nil {
		util.LogErrorf("run redis command LRANGE failed:%s", errDo.Error())
		return false
	}

	if result == nil {
		return false
	}

	errorJson := json.Unmarshal(result.([]byte), data)

	if errorJson != nil {
		util.LogErrorf("redis data unmarshal failed:%s", errorJson.Error())
		return false
	}
	return true
}

func (b *Redis) LREM(key string, count int, data interface{}) int {
	redisClient, err := InitRedisPool()

	if err != nil {
		return 0
	}

	defer redisClient.Close()

	result, errDo := redisClient.Do("LREM", key, count, data)

	if errDo != nil {
		util.LogErrorf("run redis command LREM failed:%s", errDo.Error())
		return 0
	}

	countRem, ok := result.(int)

	if !ok {
		util.LogErrorf("redis data convert to int failed:%v", result)
		return 0
	}

	return countRem
}

func (b *Redis) RPush(value interface{}) bool {
	return b.Push(value, false)
}
func (b *Redis) LPush(value interface{}) bool {
	return b.Push(value, true)
}

func (b *Redis) Push(value interface{}, isLeft bool) bool {

	var cmd string
	if isLeft {
		cmd = "LPUSH"
	} else {
		cmd = "RPUSH"
	}

	key := b.getKey("")

	return b.DoSet(cmd, key, value)
}

func (b *Redis) RPop(value interface{}) bool {
	return b.Pop(value, false)
}

func (b *Redis) LPop(value interface{}) bool {
	return b.Pop(value, true)
}

func (b *Redis) Pop(value interface{}, isLeft bool) bool {

	var cmd string
	if isLeft {
		cmd = "LPOP"
	} else {
		cmd = "RPOP"
	}
	key := b.getKey("")

	return b.DoGet(cmd, key, value)
}
func (b *Redis) DoSet(cmd string, key string, value interface{}) bool {

	redisClient, err := InitRedisPool()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	data, errJson := json.Marshal(value)

	if errJson != nil {
		util.LogErrorf("redis %s marshal data to json:%s", cmd, errJson.Error())
		return false
	}

	_, errDo := redisClient.Do(cmd, key, data)

	if errDo != nil {
		util.LogErrorf("run redis command %s failed:%s", cmd, errDo.Error())
		return false
	}
	return true
}
func (b *Redis) DoGet(cmd string, key string, value interface{}) bool {

	redisClient, err := InitRedisPool()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	result, errDo := redisClient.Do(cmd, key)

	if errDo != nil {

		util.LogErrorf("run redis %s command failed:%s", cmd, errDo.Error())

		return false
	}

	if result == nil {
		value = nil
		return true
	}

	errorJson := json.Unmarshal(result.([]byte), value)

	if errorJson != nil {

		util.LogErrorf("get %s command result failed:%s", cmd, errorJson.Error())

		return false
	}

	return true
}

//list end
