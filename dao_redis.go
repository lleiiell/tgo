package tgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"
	"reflect"
	"strings"
	"sync"
	"time"
)

type DaoRedis struct {
	KeyName string
}

var (
	pool         *pools.ResourcePool
	redisPoolMux sync.Mutex
)

type ResourceConn struct {
	redis.Conn
}

func (r ResourceConn) Close() {
	r.Conn.Close()
}

func InitRedis() (redis.Conn, error) {

	cacheConfig := ConfigCacheGetRedis()

	conn, err := redis.DialTimeout("tcp", fmt.Sprintf("%s:%d", cacheConfig.Address, cacheConfig.Port), time.Duration(cacheConfig.ConnectTimeout)*time.Millisecond, time.Duration(cacheConfig.ReadTimeout)*time.Millisecond, time.Duration(cacheConfig.WriteTimeout)*time.Millisecond)

	if err != nil {

		UtilLogErrorf("open redis error: %s", err.Error())

	}

	return conn, err
}

func InitRedisPool() (pools.Resource, error) {

	if pool == nil || pool.IsClosed() {

		redisPoolMux.Lock()

		defer redisPoolMux.Unlock()

		if pool == nil {
			cacheConfig := ConfigCacheGetRedis()

			address := fmt.Sprintf("%s:%d", cacheConfig.Address, cacheConfig.Port)

			pool = pools.NewResourcePool(func() (pools.Resource, error) {
				c, err := redis.Dial("tcp", address)
				if err != nil {
					UtilLogErrorf("open redis pool error: %s", err.Error())
					return nil, err
				}
				return ResourceConn{c}, err
			}, 1, cacheConfig.PoolMaxActive, time.Duration(cacheConfig.PoolIdleTimeout)*time.Millisecond)
		}
	}
	if pool != nil {

		var r pools.Resource
		var err error

		for i := 0; i < 1 || i < 10; i++ {
			ctx := context.TODO()
			r, err = pool.Get(ctx)
			if err != nil {
				UtilLogErrorf("redis get connection err:%s", err.Error())
			}
			if i > 0 {
				UtilLogErrorf("index is %d", i)
			}

			rc := r.(ResourceConn)

			if rc.Conn.Err() != nil {
				r.Close()
				//连接断开，重新打开
				UtilLogErrorf("redis rc connection err:%s", rc.Conn.Err().Error())
				err = rc.Conn.Err()
				continue
			}
			break
		}
		if r == nil {
			err = errors.New("redis pool resource is null")
		}
		return r, err
	}

	UtilLogError("redis pool is null")

	return ResourceConn{}, errors.New("redis pool is null")
}

func (b *DaoRedis) getKey(key string) string {

	cacheConfig := ConfigCacheGetRedis()

	prefixRedis := cacheConfig.Prefix

	if strings.Trim(key, " ") == "" {
		return fmt.Sprintf("%s:%s", prefixRedis, b.KeyName)
	}
	return fmt.Sprintf("%s:%s:%s", prefixRedis, b.KeyName, key)
}

func (b *DaoRedis) Set(key string, value interface{}) bool {

	redisResource, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

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

func (b *DaoRedis) Get(key string, data interface{}) bool {

	redisResource, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

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

func (b *DaoRedis) Incr(key string) (interface{}, bool) {

	redisResource, err := InitRedisPool()

	if err != nil {
		return 0, false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	data, errDo := redisClient.Do("INCR", b.getKey(key))

	if errDo != nil {
		UtilLogErrorf("run redis INCR command failed:%s", errDo.Error())

		return 0, false
	}
	return data, true
}

//hash start
func (b *DaoRedis) HIncrby(key string, field string, value int) (int, bool) {

	redisResource, err := InitRedisPool()

	if err != nil {
		return 0, false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

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

func (b *DaoRedis) HGet(key string, field string, value interface{}) bool {

	redisResource, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

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

func (b *DaoRedis) HSet(key string, field string, value interface{}) bool {

	redisResource, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

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
func (b *DaoRedis) HMSet(key string, datas ...interface{}) bool {

	redisResource, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	_, errDo := redisClient.Do("HMSET", b.getKey(key), datas)

	if errDo != nil {
		UtilLogErrorf("run redis HMSET command failed:%s", errDo.Error())
		return false
	}
	return true
}

func (b *DaoRedis) HLen(key string, data *int) bool {
	redisResource, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

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

func (b *DaoRedis) HDel(key string, field string) bool {

	redisResource, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	_, errDo := redisClient.Do("HDel", b.getKey(key), field)

	if errDo != nil {

		UtilLogErrorf("run redis HDel command failed:%s", errDo.Error())

		return false
	}

	return true
}

// hash end

// sorted set start
func (b *DaoRedis) ZAdd(key string, score int, data interface{}) bool {
	redisResource, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	_, errDo := redisClient.Do("ZADD", key, score, data)

	if errDo != nil {
		UtilLogErrorf("run redis ZADD command failed:%s", errDo.Error())
		return false
	}
	return true
}
func (b *DaoRedis) ZGet(key string, sort bool, start int, end int, data interface{}) bool {
	redisResource, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

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

func (b *DaoRedis) LRange(key string, start int, end int, data interface{}) bool {
	redisResource, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

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

func (b *DaoRedis) LREM(key string, count int, data interface{}) int {
	redisResource, err := InitRedisPool()

	if err != nil {
		return 0
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

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

func (b *DaoRedis) RPush(value interface{}) bool {
	return b.Push(value, false)
}
func (b *DaoRedis) LPush(value interface{}) bool {
	return b.Push(value, true)
}

func (b *DaoRedis) Push(value interface{}, isLeft bool) bool {

	var cmd string
	if isLeft {
		cmd = "LPUSH"
	} else {
		cmd = "RPUSH"
	}

	key := b.getKey("")

	return b.DoSet(cmd, key, value)
}

func (b *DaoRedis) RPop(value interface{}) bool {
	return b.Pop(value, false)
}

func (b *DaoRedis) LPop(value interface{}) bool {
	return b.Pop(value, true)
}

func (b *DaoRedis) Pop(value interface{}, isLeft bool) bool {

	var cmd string
	if isLeft {
		cmd = "LPOP"
	} else {
		cmd = "RPOP"
	}
	key := b.getKey("")

	return b.DoGet(cmd, key, value)
}
func (b *DaoRedis) DoSet(cmd string, key string, value interface{}) bool {

	redisResource, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

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
func (b *DaoRedis) DoGet(cmd string, key string, value interface{}) bool {

	redisResource, err := InitRedisPool()

	if err != nil {
		return false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

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
