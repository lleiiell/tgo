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
    Persistent bool // 持久化key
}

type redisPool struct {
    redisPool    *pools.ResourcePool
    redisPoolMux sync.Mutex
    redisPPool    *pools.ResourcePool // 持久化Pool
    redisPPoolMux sync.Mutex
}

func (p *redisPool) Get(persistent bool) (*pools.ResourcePool, sync.Mutex) {
    if persistent{
        return p.redisPPool, p.redisPPoolMux
    } else {
        return p.redisPool,p.redisPoolMux
    }
}

func (p *redisPool) Set(pool *pools.ResourcePool, persistent bool) {
    if persistent{
        p.redisPPool = pool
    } else {
        p.redisPool = pool
    }
}

func (p *redisPool) Put(resource pools.Resource, persistent bool) {
    if persistent{
        p.redisPPool.Put(resource)
    } else {
        p.redisPool.Put(resource)
    }
}

var daoPool redisPool

type ResourceConn struct {
	redis.Conn
	serverIndex int
}

func (r ResourceConn) Close() {
	r.Conn.Close()
}

/*
func (b *DaoRedis) InitRedis() (redis.Conn, error) {

	cacheConfig := ConfigCacheGetRedisWithConn(b.Persistent)()

	conn, err := redis.DialTimeout("tcp", fmt.Sprintf("%s:%d", cacheConfig.Address, cacheConfig.Port), time.Duration(cacheConfig.ConnectTimeout)*time.Millisecond, time.Duration(cacheConfig.ReadTimeout)*time.Millisecond, time.Duration(cacheConfig.WriteTimeout)*time.Millisecond)

	if err != nil {

		UtilLogErrorf("open redis error: %s", err.Error())

	}

	return conn, err
}
*/
func (b *DaoRedis) dial(fromIndex int) (redis.Conn, int, error) {

	cacheConfig := ConfigCacheGetRedisWithConn(b.Persistent)

	if len(cacheConfig.Address) > 0 {
		if fromIndex+1 > len(cacheConfig.Address) {
			fromIndex = 0
		}

		var c redis.Conn
		var err error
		for i, addr := range cacheConfig.Address {
			if i >= fromIndex {
				c, err = redis.DialTimeout("tcp", addr, time.Duration(cacheConfig.ConnectTimeout)*time.Millisecond, time.Duration(cacheConfig.ReadTimeout)*time.Millisecond, time.Duration(cacheConfig.WriteTimeout)*time.Millisecond)
				if err != nil {
					UtilLogErrorf("dail redis pool error: %s", err.Error())
				} else {
					return c, i, err
				}
			}
		}
		return c, 0, err
	} else {
		return nil, 0, errors.New("redis address lenth is 0")
	}
}
func (b *DaoRedis) InitRedisPool() (pools.Resource, error) {

    var poolHandler *pools.ResourcePool
    var poolMux sync.Mutex

    poolHandler, poolMux = daoPool.Get(b.Persistent)

	if poolHandler == nil || poolHandler.IsClosed() {

		poolMux.Lock()

		defer poolMux.Unlock()

		if poolHandler == nil {

			cacheConfig := ConfigCacheGetRedisWithConn(b.Persistent)

			if cacheConfig.PoolMinActive == 0 {
				cacheConfig.PoolMinActive = 1
			}

            poolHandler = pools.NewResourcePool(func() (pools.Resource, error) {
				c, serverIndex, err := b.dial(0)
				return ResourceConn{Conn: c, serverIndex: serverIndex}, err
			}, cacheConfig.PoolMinActive, cacheConfig.PoolMaxActive, time.Duration(cacheConfig.PoolIdleTimeout)*time.Millisecond)

            daoPool.Set(poolHandler, b.Persistent)

		}
	}
	if poolHandler != nil {
		var r pools.Resource
		var err error
		ctx := context.TODO()

		r, err = poolHandler.Get(ctx)

		if err != nil {
			UtilLogErrorf("redis get connection err:%s", err.Error())
		} else if r == nil {
			err = errors.New("redis pool resource is null")
		} else {
			rc := r.(ResourceConn)

			if rc.Conn.Err() != nil {
				UtilLogErrorf("redis rc connection err:%s,serverIndex:%d", rc.Conn.Err().Error(), rc.serverIndex)

				rc.Close()
				//连接断开，重新打开
				var conn redis.Conn
				var serverIndex int
				conn, serverIndex, err = b.dial(rc.serverIndex + 1)
				if err != nil {
                    poolHandler.Put(r)
					UtilLogErrorf("redis redail connection err:%s", err.Error())
					return nil, err
				} else {
					return ResourceConn{Conn: conn, serverIndex: serverIndex}, err
				}
			}
		}

		return r, err
	}


	UtilLogError("redis pool is null")

	return ResourceConn{}, errors.New("redis pool is null")
}

func (b *DaoRedis) getKey(key string) string {

	cacheConfig := ConfigCacheGetRedisWithConn(b.Persistent)

	prefixRedis := cacheConfig.Prefix

	if strings.Trim(key, " ") == "" {
		return fmt.Sprintf("%s:%s", prefixRedis, b.KeyName)
	}
	return fmt.Sprintf("%s:%s:%s", prefixRedis, b.KeyName, key)
}

func (b *DaoRedis) doSet(cmd string, key string, value interface{}, expire int, fields ...string) (interface{}, error) {

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return nil, err
	}

	key = b.getKey(key)

	defer daoPool.Put(redisResource, b.Persistent)

	redisClient := redisResource.(ResourceConn)

	data, errJson := json.Marshal(value)

	if errJson != nil {
		UtilLogErrorf("redis %s marshal data to json:%s", cmd, errJson.Error())
		return nil, errJson
	}

	if expire == 0 {
		cacheConfig := ConfigCacheGetRedisWithConn(b.Persistent)

		expire = cacheConfig.Expire
	}

	var reply interface{}
	var errDo error

	if len(fields) == 0 {
		if expire > 0 && strings.ToUpper(cmd) == "SET" {
			reply, errDo = redisClient.Do(cmd, key, data, "ex", expire)
		} else {
			reply, errDo = redisClient.Do(cmd, key, data)
		}

	} else {
		field := fields[0]

		reply, errDo = redisClient.Do(cmd, key, field, data)

	}

	if errDo != nil {
		UtilLogErrorf("run redis command %s failed:error:%s,key:%s,fields:%v,data:%v", cmd, errDo.Error(), key, fields, value)
		return nil, errDo
	}
	//set expire
	if expire > 0 && strings.ToUpper(cmd) != "SET" {
		_, errExpire := redisClient.Do("EXPIRE", key, expire)
		if errExpire != nil {
			UtilLogErrorf("run redis EXPIRE command failed: error:%s,key:%s,time:%d", errExpire.Error(), key, expire)
		}
	}

	return reply, errDo
}

func (b *DaoRedis) doSetNX(cmd string, key string, value interface{}, expire int, field ...string) (int64, bool) {

	reply, err := b.doSet(cmd, key, value, expire, field...)

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
	defer daoPool.Put(redisResource, b.Persistent)

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
	/*
		if expire == 0 {
			cacheConfig := ConfigCacheGetRedisWithConn(b.Persistent)()

			expire = cacheConfig.Expire
		}

		if expire > 0 {
			args = append(args, "ex", expire)
		}*/

	redisClient := redisResource.(ResourceConn)

	var reply interface{}
	var errDo error
	reply, errDo = redisClient.Do(cmd, args...)

	if errDo != nil {
		UtilLogErrorf("run redis command %s failed:error:%s,key:%s,value:%v", cmd, errDo.Error(), key, value)
		return nil, errDo
	}
	return reply, errDo
}
func (b *DaoRedis) doGet(cmd string, key string, value interface{}, fields ...string) (bool, error) {

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return false, err
	}
	defer daoPool.Put(redisResource, b.Persistent)

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

		UtilLogErrorf("run redis %s command failed: error:%s,key:%s,fields:%v", cmd, errDo.Error(), key, fields)

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
func (b *DaoRedis) doMGet(cmd string, args []interface{}, value interface{}) error {

	refValue := reflect.ValueOf(value)
	if refValue.Kind() != reflect.Ptr || refValue.Elem().Kind() != reflect.Slice || refValue.Elem().Type().Elem().Kind() != reflect.Ptr {
		return errors.New(fmt.Sprintf("value is not *[]*object:  %v", refValue.Elem().Type().Elem().Kind()))
	}
	//return errors.New(fmt.Sprintf("s:  %v", refValue.Elem().Type().Elem().Elem()))

	refSlice := refValue.Elem()
	refItem := refSlice.Type().Elem()

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return err
	}
	defer daoPool.Put(redisResource, b.Persistent)

	redisClient := redisResource.(ResourceConn)

	result, errDo := redis.ByteSlices(redisClient.Do(cmd, args...))

	if errDo != nil {
		UtilLogErrorf("run redis %s command failed: error:%s,args:%v", cmd, errDo.Error(), args)
		return errDo
	}

	if result == nil {
		return nil
	}
	if len(result) > 0 {

		for i := 0; i < len(result); i++ {
			r := result[i]

			if r != nil {
				item := reflect.New(refItem)

				errorJson := json.Unmarshal(r, item.Interface())

				if errorJson != nil {

					UtilLogErrorf("%s command result failed:%s", cmd, errorJson.Error())

					return errorJson
				}
				refSlice.Set(reflect.Append(refSlice, item.Elem()))
			} else {
				refSlice.Set(reflect.Append(refSlice, reflect.Zero(refItem)))
			}
		}
	}
	return nil
}

/*
func (b *DaoRedis) doMGet(cmd string, args []interface{}, value []interface{}) error {

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return err
	}
	defer daoPool.Put()()(redisResource)

	redisClient := redisResource.(ResourceConn)

	result, errDo := redis.ByteSlices(redisClient.Do(cmd, args...))

	if errDo != nil {
		UtilLogErrorf("run redis %s command failed: error:%s,args:%v", cmd, errDo.Error(), args)
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
*/
func (b *DaoRedis) doIncr(cmd string, key string, value int, expire int, fields ...string) (int, bool) {

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return 0, false
	}
	defer daoPool.Put(redisResource, b.Persistent)

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
		UtilLogErrorf("run redis %s command failed: error:%s,key:%s,fields:%v,value:%d", cmd, errDo.Error(), key, fields, value)

		return 0, false
	}

	count, result := data.(int64)

	if !result {

		UtilLogErrorf("get %s command result failed:%v ,is %v", cmd, data, reflect.TypeOf(data))

		return 0, false
	}

	if expire == 0 {
		cacheConfig := ConfigCacheGetRedisWithConn(b.Persistent)

		expire = cacheConfig.Expire
	}
	//set expire
	if expire > 0 {
		_, errExpire := redisClient.Do("EXPIRE", key, expire)
		if errExpire != nil {
			UtilLogErrorf("run redis EXPIRE command failed: error:%s,key:%s,time:%d", errExpire.Error(), key, expire)
		}
	}

	return int(count), true
}

func (b *DaoRedis) doDel(cmd string, data ...interface{}) error {

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return err
	}
	defer daoPool.Put(redisResource, b.Persistent)

	redisClient := redisResource.(ResourceConn)

	_, errDo := redisClient.Do(cmd, data...)

	if errDo != nil {

		UtilLogErrorf("run redis %s command failed: error:%s,data:%v", cmd, errDo.Error(), data)
	}

	return errDo
}

/*基础结束*/

func (b *DaoRedis) Set(key string, value interface{}) bool {
	_, err := b.doSet("SET", key, value, 0)

	if err != nil {
		return false
	}
	return true
}

//MSet mset
func (b *DaoRedis) MSet(datas map[string]interface{}) bool {
	_, err := b.doMSet("MSET", "", datas)
	if err != nil {
		return false
	}
	return true
}

//SetEx setex
func (b *DaoRedis) SetEx(key string, value interface{}, expire int) bool {

	_, err := b.doSet("SET", key, value, expire)

	if err != nil {
		return false
	}
	return true
}

//Expire expire
func (b *DaoRedis) Expire(key string, expire int) bool {
	redisResource, err := b.InitRedisPool()
	if err != nil {
		return false
	}
	key = b.getKey(key)
	defer daoPool.Put(redisResource, b.Persistent)
	redisClient := redisResource.(ResourceConn)
	_, err = redisClient.Do("EXPIRE", key, expire)
	if err != nil {
		UtilLogErrorf("run redis EXPIRE command failed: error:%s,key:%s,time:%d", err.Error(), key, expire)
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

// 返回 1. key是否存在 2. error
func (b *DaoRedis) GetRaw(key string, data interface{}) (bool, error) {

    return b.doGet("GET", key, data)
}

func (b *DaoRedis) MGet(keys []string, data interface{}) error {

	var args []interface{}

	for _, v := range keys {
		args = append(args, b.getKey(v))
	}

	err := b.doMGet("MGET", args, data)

	return err
}

func (b *DaoRedis) Incr(key string) (int, bool) {

	return b.doIncr("INCRBY", key, 1, 0)
}

func (b *DaoRedis) IncrBy(key string, value int) (int, bool) {

	return b.doIncr("INCRBY", key, value, 0)
}
func (b *DaoRedis) SetNX(key string, value interface{}) (int64, bool) {

	return b.doSetNX("SETNX", key, value, 0)
}

func (b *DaoRedis) Del(key string) bool {

	key = b.getKey(key)

	err := b.doDel("DEL", key)

	if err != nil {
		return false
	}

	return true
}

func (b *DaoRedis) MDel(key ...string) bool {
	var keys []interface{}
	for _, v := range key {
		keys = append(keys, b.getKey(v))
	}

	err := b.doDel("DEL", keys...)

	if err != nil {
		return false
	}

	return true
}

//hash start
func (b *DaoRedis) HIncrby(key string, field string, value int) (int, bool) {

	return b.doIncr("HINCRBY", key, value, 0, field)
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

//HGetRaw 返回 1. key是否存在 2. error
func (b *DaoRedis) HGetRaw(key string, field string, value interface{}) (bool, error) {
    return b.doGet("HGET", key, value, field)
}

func (b *DaoRedis) HMGet(key string, fields []interface{}, data interface{}) error {
	var args []interface{}

	args = append(args, b.getKey(key))

	for _, v := range fields {
		args = append(args, v)
	}

	err := b.doMGet("HMGET", args, data)

	return err
}

func (b *DaoRedis) HSet(key string, field string, value interface{}) bool {

	_, err := b.doSet("HSET", key, value, 0, field)

	if err != nil {
		return false
	}
	return true
}
func (b *DaoRedis) HSetNX(key string, field string, value interface{}) (int64, bool) {

	return b.doSetNX("HSETNX", key, value, 0, field)
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
	defer daoPool.Put(redisResource, b.Persistent)

	redisClient := redisResource.(ResourceConn)

	resultData, errDo := redisClient.Do("HLEN", key)

	if errDo != nil {
		UtilLogErrorf("run redis HLEN command failed: error:%s,key:%s", errDo.Error(), key)
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
		UtilLogErrorf("run redis HDEL command failed: error:%s,key:%s,data:%v", err.Error(), key, data)
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
	defer daoPool.Put(redisResource, b.Persistent)

	redisClient := redisResource.(ResourceConn)

	_, errDo := redisClient.Do("ZADD", key, score, data)

	if errDo != nil {
		UtilLogErrorf("run redis ZADD command failed: error:%s,key:%s,score:%d,data:%v", errDo.Error(), key, score, data)
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

func (b *DaoRedis) ZGet(key string, sort bool, start int, end int, value interface{}) error {

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

func (b *DaoRedis) ZRevRange(key string, start int, end int, value interface{}) error {
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

func (b *DaoRedis) LLen(key string) (int64,error) {
	cmd := "LLEN"

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return 0, err
	}
	defer daoPool.Put(redisResource, b.Persistent)

	redisClient := redisResource.(ResourceConn)
	key = b.getKey(key)

	var result interface{}
	var errDo error

	var args []interface{}
	args = append(args, key)
	result, errDo = redisClient.Do(cmd, key)

	if errDo != nil {

		UtilLogErrorf("run redis %s command failed: error:%s,key:%s", cmd, errDo.Error(), key)

		return 0, errDo
	}

	if result == nil {
		return 0, nil
	}

	num, ok := result.(int64)
	if !ok {
		return 0,errors.New("result to int64 failed")
	}

	return num, nil
}

func (b *DaoRedis) LREM(key string, count int, data interface{}) int {
	redisResource, err := b.InitRedisPool()

	if err != nil {
		return 0
	}
	defer daoPool.Put(redisResource, b.Persistent)

	redisClient := redisResource.(ResourceConn)

	result, errDo := redisClient.Do("LREM", key, count, data)

	if errDo != nil {
		UtilLogErrorf("run redis command LREM failed: error:%s,key:%s,count:%d,data:%v", errDo.Error(), key, count, data)
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

	_, err := b.doSet(cmd, key, value, -1)

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

//pipeline start

func (b *DaoRedis) PipelineHGet(key []string, fields []interface{}, data []interface{}) error {
	var args [][]interface{}

	for k, v := range key {
		var arg []interface{}
		arg = append(arg, b.getKey(v))
		arg = append(arg, fields[k])
		args = append(args, arg)
	}

	err := b.pipeDoGet("HGET", args, data)

	return err
}

func (b *DaoRedis) pipeDoGet(cmd string, args [][]interface{}, value []interface{}) error {

	redisResource, err := b.InitRedisPool()

	if err != nil {
		return err
	}
	defer daoPool.Put(redisResource, b.Persistent)

	redisClient := redisResource.(ResourceConn)

	for _, v := range args {
		if err := redisClient.Send(cmd, v...); err != nil {
			UtilLogErrorf("Send(%v) returned error %v", v, err)
			return err
		}
	}
	if err := redisClient.Flush(); err != nil {
		UtilLogErrorf("Flush() returned error %v", err)
		return err
	}
	for k, v := range args {
		result, err := redisClient.Receive()
		if err != nil {
			UtilLogErrorf("Receive(%v) returned error %v", v, err)
			return err
		}
		if result == nil {
			value[k] = nil
			continue
		}
		if reflect.TypeOf(result).Kind() == reflect.Slice {

			byteResult := (result.([]byte))
			strResult := string(byteResult)

			if strResult == "[]" {
				value[k] = nil
				continue
			}
		}

		errorJson := json.Unmarshal(result.([]byte), value[k])

		if errorJson != nil {
			UtilLogErrorf("get %s command result failed:%s", cmd, errorJson.Error())
			return errorJson
		}
	}

	return nil
}

//pipeline end
