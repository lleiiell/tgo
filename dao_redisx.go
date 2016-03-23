package tgo

import (
	"encoding/json"
	"fmt"
	//"github.com/garyburd/redigo/redis"
	"github.com/chasex/redis-go-cluster"
	"reflect"
	"strings"
	"sync"
	"time"
)

type DaoRedisX struct {
	KeyName string
}

var (
	cluster *redis.Cluster

	redisMux sync.Mutex
)

/*
func InitRedisX() (redis.Conn, error) {

	cacheConfig := ConfigCacheGetRedis()

	conn, err := redis.DialTimeout("tcp", fmt.Sprintf("%s:%d", cacheConfig.Address, cacheConfig.Port), time.Duration(cacheConfig.ConnectTimeout)*time.Millisecond, time.Duration(cacheConfig.ReadTimeout)*time.Millisecond, time.Duration(cacheConfig.WriteTimeout)*time.Millisecond)

	if err != nil {

		UtilLogErrorf("open redis error: %s", err.Error())

	}

	return conn, err
}*/

func InitRedisPoolX() (*redis.Cluster, error) {

	var err error

	if cluster == nil {
		redisMux.Lock()

		defer redisMux.Unlock()

		if cluster == nil {
			cacheConfig := ConfigCacheGetRedis()

			cluster, err = redis.NewCluster(
				&redis.Options{
					StartNodes:   []string{"172.172.177.52:7000", "172.172.177.52:8000", "172.172.177.53:7000", "172.172.177.53:8001", "172.172.177.54:7000", "172.172.177.54:8000"},
					ConnTimeout:  time.Duration(cacheConfig.ConnectTimeout) * time.Millisecond,
					ReadTimeout:  time.Duration(cacheConfig.ReadTimeout) * time.Millisecond,
					WriteTimeout: time.Duration(cacheConfig.WriteTimeout) * time.Millisecond,
					KeepAlive:    16,
					AliveTime:    60 * time.Second,
				})

			if err != nil {
				UtilLogErrorf("open redis cluster error: %s", err.Error())
			}
		}
	}

	return cluster, err
}

func (b *DaoRedisX) getKey(key string) string {

	cacheConfig := ConfigCacheGetRedis()

	prefixRedis := cacheConfig.Prefix

	if strings.Trim(key, " ") == "" {
		return fmt.Sprintf("%s:%s", prefixRedis, b.KeyName)
	}
	return fmt.Sprintf("%s:%s:%s", prefixRedis, b.KeyName, key)
}

func (b *DaoRedisX) Set(key string, value interface{}) bool {

	redisClient, err := InitRedisPoolX()

	if err != nil {
		return false
	}
	defer redisClient.Close()
	data, errJson := json.Marshal(value)

	if errJson != nil {
		UtilLogErrorf("redis Set marshal data to json:%s", errJson.Error())
		return false
	}
	_, errDo := redisClient.Do("SET", b.getKey(key), data)

	if errDo != nil {
		UtilLogErrorf("run redis command Set failed:%s", errDo.Error())
		return false
	}
	return true
}

func (b *DaoRedisX) Get(key string, data interface{}) bool {

	redisClient, err := InitRedisPoolX()

	if err != nil {
		return false
	}
	defer redisClient.Close()

	result, errDo := redisClient.Do("GET", b.getKey(key))

	if errDo != nil {
		UtilLogErrorf("run redis command GET failed:%s", errDo.Error())
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
		UtilLogErrorf("redis GET data  unmarshal failed:%s", errorJson.Error())
		return false
	}
	return true
}

func (b *DaoRedisX) Incr(key string) (interface{}, bool) {

	redisClient, err := InitRedisPoolX()

	if err != nil {
		return 0, false
	}

	defer redisClient.Close()

	data, errDo := redisClient.Do("INCR", b.getKey(key))

	if errDo != nil {
		UtilLogErrorf("run redis INCR command failed:%s", errDo.Error())

		return 0, false
	}
	return data, true
}

//hash start
func (b *DaoRedisX) HIncrby(key string, field string, value int) (int, bool) {

	redisClient, err := InitRedisPoolX()

	if err != nil {
		return 0, false
	}

	defer redisClient.Close()

	data, errDo := redisClient.Do("HINCRBY", b.getKey(key), field, value)

	if errDo != nil {
		UtilLogErrorf("run redis HINCRBY command failed:%s", errDo.Error())

		return 0, false
	}

	count, result := data.(int64)

	if !result {

		UtilLogErrorf("get HINCRBY command result failed:%v ,is %v", data, reflect.TypeOf(data))

		return 0, false
	}
	return int(count), true
}

func (b *DaoRedisX) HGet(key string, field string, value interface{}) bool {

	redisClient, err := InitRedisPoolX()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	result, errDo := redisClient.Do("HGet", b.getKey(key), field)

	if errDo != nil {
		UtilLogErrorf("run redis HGet command failed:%s", errDo.Error())

		return false
	}

	if result == nil {

		return false
	}

	errorJson := json.Unmarshal(result.([]byte), value)

	if errorJson != nil {

		UtilLogErrorf("get HGet command result failed:%s", errorJson.Error())

		return false
	}

	return true
}

func (b *DaoRedisX) HSet(key string, field string, value interface{}) bool {

	redisClient, err := InitRedisPoolX()

	if err != nil {
		return false
	}
	defer redisClient.Close()
	data, errJson := json.Marshal(value)

	if errJson != nil {
		UtilLogErrorf("redis HSET marshal data to json:%s", errJson.Error())
		return false
	}
	_, errDo := redisClient.Do("HSET", b.getKey(key), field, data)

	if errDo != nil {
		UtilLogErrorf("run redis HSET command failed:%s", errDo.Error())
		return false
	}
	return true
}
func (b *DaoRedisX) HMSet(key string, datas ...interface{}) bool {

	redisClient, err := InitRedisPoolX()

	if err != nil {
		return false
	}
	defer redisClient.Close()

	_, errDo := redisClient.Do("HMSET", b.getKey(key), datas)

	if errDo != nil {
		UtilLogErrorf("run redis HMSET command failed:%s", errDo.Error())
		return false
	}
	return true
}

func (b *DaoRedisX) HLen(key string, data *int) bool {
	redisClient, err := InitRedisPoolX()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	resultData, errDo := redisClient.Do("HLEN", key)

	if errDo != nil {
		UtilLogErrorf("run redis ZADD command failed:%s", errDo.Error())
		return false
	}

	lenth, resultConv := resultData.(int)

	if !resultConv {
		UtilLogErrorf("redis data convert to int failed:%v", resultConv)

	}

	data = &lenth

	return resultConv
}

func (b *DaoRedisX) HDel(key string, field string) bool {

	redisClient, err := InitRedisPoolX()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	_, errDo := redisClient.Do("HDel", b.getKey(key), field)

	if errDo != nil {

		UtilLogErrorf("run redis HDel command failed:%s", errDo.Error())

		return false
	}

	return true
}

// hash end

// sorted set start
func (b *DaoRedisX) ZAdd(key string, score int, data interface{}) bool {
	redisClient, err := InitRedisPoolX()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	_, errDo := redisClient.Do("ZADD", key, score, data)

	if errDo != nil {
		UtilLogErrorf("run redis ZADD command failed:%s", errDo.Error())
		return false
	}
	return true
}
func (b *DaoRedisX) ZGet(key string, sort bool, start int, end int, data interface{}) bool {
	redisClient, err := InitRedisPoolX()

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
		UtilLogErrorf("run redis command ZREVRANGE failed:%s", errDo.Error())
		return false
	}

	if result == nil {
		return false
	}

	errorJson := json.Unmarshal(result.([]byte), data)

	if errorJson != nil {
		UtilLogErrorf("redis data unmarshal failed:%s", errorJson.Error())
		return false
	}
	return true
}

//list start

func (b *DaoRedisX) LRange(key string, start int, end int, data interface{}) bool {
	redisClient, err := InitRedisPoolX()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	result, errDo := redisClient.Do("LRANGE", key, start, end)

	if errDo != nil {
		UtilLogErrorf("run redis command LRANGE failed:%s", errDo.Error())
		return false
	}

	if result == nil {
		return false
	}

	errorJson := json.Unmarshal(result.([]byte), data)

	if errorJson != nil {
		UtilLogErrorf("redis data unmarshal failed:%s", errorJson.Error())
		return false
	}
	return true
}

func (b *DaoRedisX) LREM(key string, count int, data interface{}) int {
	redisClient, err := InitRedisPoolX()

	if err != nil {
		return 0
	}

	defer redisClient.Close()

	result, errDo := redisClient.Do("LREM", key, count, data)

	if errDo != nil {
		UtilLogErrorf("run redis command LREM failed:%s", errDo.Error())
		return 0
	}

	countRem, ok := result.(int)

	if !ok {
		UtilLogErrorf("redis data convert to int failed:%v", result)
		return 0
	}

	return countRem
}

func (b *DaoRedisX) RPush(value interface{}) bool {
	return b.Push(value, false)
}
func (b *DaoRedisX) LPush(value interface{}) bool {
	return b.Push(value, true)
}

func (b *DaoRedisX) Push(value interface{}, isLeft bool) bool {

	var cmd string
	if isLeft {
		cmd = "LPUSH"
	} else {
		cmd = "RPUSH"
	}

	key := b.getKey("")

	return b.DoSet(cmd, key, value)
}

func (b *DaoRedisX) RPop(value interface{}) bool {
	return b.Pop(value, false)
}

func (b *DaoRedisX) LPop(value interface{}) bool {
	return b.Pop(value, true)
}

func (b *DaoRedisX) Pop(value interface{}, isLeft bool) bool {

	var cmd string
	if isLeft {
		cmd = "LPOP"
	} else {
		cmd = "RPOP"
	}
	key := b.getKey("")

	return b.DoGet(cmd, key, value)
}
func (b *DaoRedisX) DoSet(cmd string, key string, value interface{}) bool {

	redisClient, err := InitRedisPoolX()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	data, errJson := json.Marshal(value)

	if errJson != nil {
		UtilLogErrorf("redis %s marshal data to json:%s", cmd, errJson.Error())
		return false
	}

	_, errDo := redisClient.Do(cmd, key, data)

	if errDo != nil {
		UtilLogErrorf("run redis command %s failed:%s", cmd, errDo.Error())
		return false
	}
	return true
}
func (b *DaoRedisX) DoGet(cmd string, key string, value interface{}) bool {

	redisClient, err := InitRedisPoolX()

	if err != nil {
		return false
	}

	defer redisClient.Close()

	result, errDo := redisClient.Do(cmd, key)

	if errDo != nil {

		UtilLogErrorf("run redis %s command failed:%s", cmd, errDo.Error())

		return false
	}

	if result == nil {
		value = nil
		return true
	}

	errorJson := json.Unmarshal(result.([]byte), value)

	if errorJson != nil {

		UtilLogErrorf("get %s command result failed:%s", cmd, errorJson.Error())

		return false
	}

	return true
}

//list end
