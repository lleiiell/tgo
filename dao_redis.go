package tgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"
)

type DaoRedis struct {
	KeyName string
}

var (
	redisPool         *pools.ResourcePool
	redisPoolMux sync.Mutex
)

type ResourceConn struct {
	redis.Conn
}

func (r ResourceConn) Close() {
	r.Conn.Close()
}

func (b *DaoRedis) InitRedis() (redis.Conn, error) {

	cacheConfig := ConfigCacheGetRedis()

	conn, err := redis.DialTimeout("tcp", fmt.Sprintf("%s:%d", cacheConfig.Address, cacheConfig.Port), time.Duration(cacheConfig.ConnectTimeout)*time.Millisecond, time.Duration(cacheConfig.ReadTimeout)*time.Millisecond, time.Duration(cacheConfig.WriteTimeout)*time.Millisecond)

	if err != nil {

		UtilLogErrorf("open redis error: %s", err.Error())

	}

	return conn, err
}

func (b *DaoRedis) dail() (redis.Conn, error) {

	cacheConfig := ConfigCacheGetRedis()
	address := fmt.Sprintf("%s:%d", cacheConfig.Address, cacheConfig.Port)
	c, err := redis.DialTimeout("tcp", address, 0, time.Duration(cacheConfig.ReadTimeout)*time.Millisecond, time.Duration(cacheConfig.WriteTimeout)*time.Millisecond)
	if err != nil {
		UtilLogErrorf("open redis pool error: %s", err.Error())
		return nil, err
	}

	return c, err

}
func (b *DaoRedis) InitRedisPool() (pools.Resource, error) {

	if redisPool == nil || redisPool.IsClosed() {

		redisPoolMux.Lock()

		defer redisPoolMux.Unlock()

		if redisPool == nil {

			cacheConfig := ConfigCacheGetRedis()

			if cacheConfig.PoolMinActive == 0 {
				cacheConfig.PoolMinActive = 1
			}

			redisPool = pools.NewResourcePool(func() (pools.Resource, error) {
				c, err := b.dail()
				return ResourceConn{c}, err
			}, cacheConfig.PoolMinActive, cacheConfig.PoolMaxActive, time.Duration(cacheConfig.PoolIdleTimeout)*time.Millisecond)
		}
	}
	if redisPool != nil {
		var r pools.Resource
		var err error
		/*
			if pool.Available() == 0 {
				var conn redis.Conn
				conn, err = b.dail()


				if err != nil {
					UtilLogErrorf("redis ava dail connection err:%s", err.Error())
				}
				return ResourceConn{conn}, err
			}*/

		ctx := context.TODO()

		r, err = redisPool.Get(ctx)

		if err != nil {
			UtilLogErrorf("redis get connection err:%s", err.Error())
		} else if r == nil {
			err = errors.New("redis pool resource is null")
		} else {
			rc := r.(ResourceConn)

			if rc.Conn.Err() != nil {
				UtilLogErrorf("redis rc connection err:%s", rc.Conn.Err().Error())

				rc.Close()
				//连接断开，重新打开
				var conn redis.Conn
				conn, err = b.dail()
				if err != nil {
					UtilLogErrorf("redis redail connection err:%s", err.Error())
					return nil, err
				} else {
					return ResourceConn{conn}, err
				}
			}
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

func (b *DaoRedis) doSet(cmd string, key string, value interface{}, fields ...string) (interface{}, error) {

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return nil, err
	}

	key = b.getKey(key)

	defer redisPool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	data, errJson := json.Marshal(value)

	if errJson != nil {
		UtilLogErrorf("redis %s marshal data to json:%s", cmd, errJson.Error())
		return nil, errJson
	}
	var reply interface{}
	var errDo error

	if len(fields) == 0 {
		reply, errDo = redisClient.Do(cmd, key, data)
	} else {
		field := fields[0]
		reply, errDo = redisClient.Do(cmd, key, field, data)
	}

	if errDo != nil {
		UtilLogErrorf("run redis command %s failed:%s", cmd, errDo.Error())
		return nil, errDo
	}
	return reply, errDo
}

func (b *DaoRedis) doSetNX(cmd string, key string, value interface{}, field ...string) (int64, bool) {

	reply, err := b.doSet(cmd, key, value, field...)

	if err != nil {
		return 0, false
	}

	replyInt, ok := reply.(int64)

	if !ok {
		UtilLogErrorf("HSetNX reply to int failed,key:%s,field:%s", key, field)

		return 0, false
	}

	return replyInt, true
}
func (b *DaoRedis) doMSet(cmd string, key string, value map[string]interface{}) (interface{}, error) {

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return nil, err
	}
	defer redisPool.Put(redisResource)

	var args []interface{}

	if key != "" {
		key = b.getKey(key)
		args = append(args, key)
	}

	for k, v := range value {
		data, errJson := json.Marshal(v)

		if errJson != nil {
			UtilLogErrorf("redis %s marshal data: %v to json:%s", cmd, v, errJson.Error())
			return nil, errJson
		}
		if key == "" {
			args = append(args, b.getKey(k), data)
		} else {
			args = append(args, k, data)
		}

	}

	redisClient := redisResource.(ResourceConn)

	var reply interface{}
	var errDo error
	reply, errDo = redisClient.Do(cmd, args...)

	if errDo != nil {
		UtilLogErrorf("run redis command %s failed:%s", cmd, errDo.Error())
		return nil, errDo
	}
	return reply, errDo
}
func (b *DaoRedis) doGet(cmd string, key string, value interface{}, fields ...string) (bool, error) {

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return false, err
	}
	defer redisPool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	key = b.getKey(key)

	var result interface{}
	var errDo error

	//if len(fields) == 0 {
	//	result, errDo = redisClient.Do(cmd, key)
	//} else {
	var args []interface{}

	args = append(args, key)

	for _, f := range fields {
		args = append(args, f)
	}

	result, errDo = redisClient.Do(cmd, args...)
	//}

	if errDo != nil {

		UtilLogErrorf("run redis %s command failed:%s", cmd, errDo.Error())

		return false, errDo
	}

	if result == nil {
		value = nil
		return false, nil
	}

	if reflect.TypeOf(result).Kind() == reflect.Slice {

		byteResult := (result.([]byte))
		strResult := string(byteResult)

		if strResult == "[]" {
			return true, nil
		}
	}

	errorJson := json.Unmarshal(result.([]byte), value)

	if errorJson != nil {

		UtilLogErrorf("get %s command result failed:%s", cmd, errorJson.Error())

		return false, errorJson
	}

	return true, nil
}

func (b *DaoRedis) doMGet(cmd string, args []interface{}, value []interface{}) error {

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return err
	}
	defer redisPool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	result, errDo := redis.ByteSlices(redisClient.Do(cmd, args...))

	if errDo != nil {
		UtilLogErrorf("run redis %s command failed:%s", cmd, errDo.Error())
		return errDo
	}

	if result == nil {
		return nil
	}
	if len(result) > 0 {

		for i := 0; i < len(result); i++ {
			r := result[i]
			if i >= len(value) {
				break
			}
			if r != nil {
				errorJson := json.Unmarshal(r, value[i])

				if errorJson != nil {

					UtilLogErrorf("%s command result failed:%s", cmd, errorJson.Error())

					return errorJson
				}
			} else {
				value[i] = nil
			}

		}
	}
	return nil
}

func (b *DaoRedis) doIncr(cmd string, key string, value int, fields ...string) (int, bool) {

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return 0, false
	}
	defer redisPool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	key = b.getKey(key)

	var data interface{}
	var errDo error

	if len(fields) == 0 {
		data, errDo = redisClient.Do(cmd, key, value)
	} else {
		field := fields[0]
		data, errDo = redisClient.Do(cmd, key, field, value)
	}

	if errDo != nil {
		UtilLogErrorf("run redis %s command failed:%s", cmd, errDo.Error())

		return 0, false
	}

	count, result := data.(int64)

	if !result {

		UtilLogErrorf("get %s command result failed:%v ,is %v", cmd, data, reflect.TypeOf(data))

		return 0, false
	}
	return int(count), true
}

func (b *DaoRedis) doDel(cmd string, data ...interface{}) error {

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return err
	}
	defer redisPool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	_, errDo := redisClient.Do(cmd, data...)

	if errDo != nil {

		UtilLogErrorf("run redis %s command failed:%s", cmd, errDo.Error())
	}

	return errDo
}

/*基础结束*/

func (b *DaoRedis) Set(key string, value interface{}) bool {

	_, err := b.doSet("SET", key, value)

	if err != nil {
		return false
	}
	return true
}
func (b *DaoRedis) MSet(datas map[string]interface{}) bool {
	_, err := b.doMSet("MSET", "", datas)
	if err != nil {
		return false
	}
	return true
}

func (b *DaoRedis) SetEx(key string, value interface{}, time int) bool {

	_, err := b.doSet("SET", key, value)

	if err != nil {
		return false
	}
	e := b.Expire(key, time)
	if e {
		return true
	} else {
		return false
	}
}

func (b *DaoRedis) Expire(key string, time int) bool {
	redisResource, err := b.InitRedisPool()
	if err != nil {
		return false
	}
	key = b.getKey(key)
	defer redisPool.Put(redisResource)
	redisClient := redisResource.(ResourceConn)
	_, err = redisClient.Do("EXPIRE", key, time)
	if err != nil {
		return false
	}

	return true
}

func (b *DaoRedis) Get(key string, data interface{}) bool {

	result, err := b.doGet("GET", key, data)

	if err == nil && result {
		return true
	}
	return false
}
func (b *DaoRedis) GetE(key string, data interface{}) error {

	_, err := b.doGet("GET", key, data)

	return err
}
func (b *DaoRedis) MGet(keys []string, data []interface{}) error {

	var args []interface{}

	for _, v := range keys {
		args = append(args, b.getKey(v))
	}

	err := b.doMGet("MGET", args, data)

	return err
}

func (b *DaoRedis) Incr(key string) (int, bool) {

	return b.doIncr("INCRBY", key, 1)
}

func (b *DaoRedis) IncrBy(key string, value int) (int, bool) {

	return b.doIncr("INCRBY", key, value)
}
func (b *DaoRedis) SetNX(key string, value interface{}) (int64, bool) {

	return b.doSetNX("SETNX", key, value)
}

func (b *DaoRedis) Del(key string) bool {

	key = b.getKey(key)

	err := b.doDel("DEL", key)

	if err != nil {
		return false
	}

	return true
}

//hash start
func (b *DaoRedis) HIncrby(key string, field string, value int) (int, bool) {

	return b.doIncr("HINCRBY", key, value, field)
}

func (b *DaoRedis) HGet(key string, field string, value interface{}) bool {

	result, err := b.doGet("HGET", key, value, field)

	if err == nil && result {
		return true
	}
	return false
}

//HGetE 返回error
func (b *DaoRedis) HGetE(key string, field string, value interface{}) error {

	_, err := b.doGet("HGET", key, value, field)

	return err
}

func (b *DaoRedis) HMGet(key string, fields []interface{}, data []interface{}) error {
	var args []interface{}

	args = append(args, b.getKey(key))

	for _, v := range fields {
		args = append(args, v)
	}

	err := b.doMGet("HMGET", args, data)

	return err
}

func (b *DaoRedis) HSet(key string, field string, value interface{}) bool {

	_, err := b.doSet("HSET", key, value, field)

	if err != nil {
		return false
	}
	return true
}
func (b *DaoRedis) HSetNX(key string, field string, value interface{}) (int64, bool) {

	return b.doSetNX("HSETNX", key, value, field)
}

//HMSet value是filed:data
func (b *DaoRedis) HMSet(key string, value map[string]interface{}) bool {
	_, err := b.doMSet("HMSet", key, value)
	if err != nil {
		return false
	}
	return true
}

func (b *DaoRedis) HLen(key string, data *int) bool {
	redisResource, err := b.InitRedisPool()

	if err != nil {
		return false
	}
	defer redisPool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	resultData, errDo := redisClient.Do("HLEN", key)

	if errDo != nil {
		UtilLogErrorf("run redis HLEN command failed:%s", errDo.Error())
		return false
	}

	lenth, resultConv := resultData.(int)

	if !resultConv {
		UtilLogErrorf("redis data convert to int failed:%v", resultConv)

	}

	data = &lenth

	return resultConv
}

func (b *DaoRedis) HDel(key string, data ...interface{}) bool {
	var args []interface{}

	key = b.getKey(key)
	args = append(args, key)

	for _, item := range data {
		args = append(args, item)
	}

	err := b.doDel("HDEL", args...)

	if err != nil {
		return false
	}

	return true
}

// hash end

// sorted set start
func (b *DaoRedis) ZAdd(key string, score int, data interface{}) bool {
	redisResource, err := b.InitRedisPool()

	if err != nil {
		return false
	}
	defer redisPool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	_, errDo := redisClient.Do("ZADD", key, score, data)

	if errDo != nil {
		UtilLogErrorf("run redis ZADD command failed:%s", errDo.Error())
		return false
	}
	return true
}

// sorted set start
func (b *DaoRedis) ZAddM(key string, value map[string]interface{}) bool {
	_, err := b.doMSet("ZADD", key, value)
	if err != nil {
		return false
	}
	return true
}

func (b *DaoRedis) ZGet(key string, sort bool, start int, end int, value []interface{}) error {

	var cmd string
	if sort {
		cmd = "ZRANGE"
	} else {
		cmd = "ZREVRANGE"
	}

	var args []interface{}
	args = append(args, b.getKey(key))
	args = append(args, start)
	args = append(args, end)
	err := b.doMGet(cmd, args, value)

	return err
}

func (b *DaoRedis) ZRevRange(key string, start int, end int, value []interface{}) error {
	return b.ZGet(key, false, start, end, value)
}

func (b *DaoRedis) ZRem(key string, data ...interface{}) bool {

	var args []interface{}

	key = b.getKey(key)
	args = append(args, key)

	for _, item := range data {
		args = append(args, item)
	}

	err := b.doDel("ZREM", args...)

	if err != nil {
		return false
	}
	return true
}

//list start

func (b *DaoRedis) LRange(key string, start int, end int, value interface{}) bool {

	cmd := "LRANGE"

	strStart := strconv.Itoa(start)
	strEnd := strconv.Itoa(end)
	_, err := b.doGet(cmd, key, value, strStart, strEnd)
	if err == nil {
		return true
	} else {
		return false
	}
}

func (b *DaoRedis) LREM(key string, count int, data interface{}) int {
	redisResource, err := b.InitRedisPool()

	if err != nil {
		return 0
	}
	defer redisPool.Put(redisResource)

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

	key := ""

	_, err := b.doSet(cmd, key, value)

	if err != nil {
		return false
	}
	return true
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
	key := ""

	_, err := b.doGet(cmd, key, value)

	if err == nil {
		return true
	} else {
		return false
	}
}

//list end
