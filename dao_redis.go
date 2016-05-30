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
	pool         *pools.ResourcePool
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

	if pool == nil || pool.IsClosed() {

		redisPoolMux.Lock()

		defer redisPoolMux.Unlock()

		if pool == nil {

			cacheConfig := ConfigCacheGetRedis()

			if cacheConfig.PoolMinActive == 0 {
				cacheConfig.PoolMinActive = 1
			}

			pool = pools.NewResourcePool(func() (pools.Resource, error) {
				c, err := b.dail()
				return ResourceConn{c}, err
			}, cacheConfig.PoolMinActive, cacheConfig.PoolMaxActive, time.Duration(cacheConfig.PoolIdleTimeout)*time.Millisecond)
		}
	}
	if pool != nil {
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

		r, err = pool.Get(ctx)

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

	defer pool.Put(redisResource)

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
	defer pool.Put(redisResource)

	var args []interface{}

	if key !=""{
			key = b.getKey(key)
			args = append(args,key)
	}

	for k,v:=range value{
		data, errJson := json.Marshal(v)

		if errJson != nil {
			UtilLogErrorf("redis %s marshal data: %v to json:%s", cmd,v, errJson.Error())
			return nil, errJson
		}
		args = append(args,k,data)
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
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	key = b.getKey(key)

	var result interface{}
	var errDo error

	if len(fields) == 0 {
		result, errDo = redisClient.Do(cmd, key)
	} else {
		field := fields[0]
		result, errDo = redisClient.Do(cmd, key, field)
	}

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

func (b *DaoRedis) doIncr(cmd string, key string, value int, fields ...string) (int, bool) {

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return 0, false
	}
	defer pool.Put(redisResource)

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

func (b *DaoRedis) doDel(cmd string,key string,data ...interface{}) error{

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return err
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	var args []interface{}

	if key !=""{
			key = b.getKey(key)
			args = append(args,key)
	}

	for _,item:= range data{
		args = append(args,item)
	}

	_, errDo := redisClient.Do(cmd, args...)

	if errDo != nil {

		UtilLogErrorf("run redis %s command failed:%s",cmd, errDo.Error())
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

func (b *DaoRedis) Incr(key string) (int, bool) {

	return b.doIncr("INCRBY", key, 1)
}

func (b *DaoRedis) IncrBy(key string, value int) (int, bool) {

	return b.doIncr("INCRBY", key, value)
}
func (b *DaoRedis) SetNX(key string, value interface{}) (int64, bool) {

	return b.doSetNX("SETNX", key, value)
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

func (b *DaoRedis) HMGet(key string, value interface{}, fields ...string) bool {
	redisResource, err := b.InitRedisPool()

	if err != nil {
		return false
	}
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	key = b.getKey(key)

	var result interface{}
	var errDo error

	if len(fields) == 0 {
		return false
	} else {
		result, errDo = redisClient.Do("HMGET", redis.Args{}.Add(key).AddFlat(fields)...)
	}

	if errDo != nil {

		UtilLogErrorf("run redis HMGET command failed:%s", errDo.Error())
		return false
	}

	if result == nil {
		value = nil
		return false
	}

	errorJson := json.Unmarshal(result.([]byte), value)

	if errorJson != nil {

		UtilLogErrorf("HMGET command result failed:%s", errorJson.Error())

		return false
	}

	return true
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
	_,err:= b.doMSet("HMSet", key, value)
	if err!=nil{
		return false
	}
	return true
}

func (b *DaoRedis) HLen(key string, data *int) bool {
	redisResource, err := b.InitRedisPool()

	if err != nil {
		return false
	}
	defer pool.Put(redisResource)

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

func (b *DaoRedis) HDel(key string, field string) bool {

	err := b.doDel("HDel", key, field)

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
	defer pool.Put(redisResource)

	redisClient := redisResource.(ResourceConn)

	_, errDo := redisClient.Do("ZADD", key, score, data)

	if errDo != nil {
		UtilLogErrorf("run redis ZADD command failed:%s", errDo.Error())
		return false
	}
	return true
}

// sorted set start
func (b *DaoRedis) ZAddM(key string, value map[string]interface{})bool {
	_,err:= b.doMSet("ZADD", key, value)
	if err!=nil{
		return false
	}
	return true
}

func (b *DaoRedis) ZGet(key string, sort bool, start int, end int, value interface{}) bool {

	var cmd string
	if sort {
		cmd = "ZRANGE"
	} else {
		cmd = "ZREVRANGE"
	}

	strStart := strconv.Itoa(start)
	strEnd := strconv.Itoa(end)
	_, err := b.doGet(cmd, key, value, strStart, strEnd)

	if err == nil {
		return true
	} else {
		return false
	}
}

func (b *DaoRedis) ZRevRange(key string,start int,end int,value interface{}) bool{
	return b.ZGet(key, false, start, end, value)
}

func (b *DaoRedis) ZRem(key string,data ...interface{}) bool{
	err:= b.doDel("ZREM", key, data...)

	if err!=nil{
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
