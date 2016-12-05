package tgo

import (
	"strconv"
	"testing"
)

type ModelRedisHello struct {
	HelloWord string
}

func Test_Redis_Set(t *testing.T) {
	redis := NewRedisTest()

	redis.Set("tonyjt", "12345678")

	var data string
	result := redis.Get("tonyjt", &data)

	if !result {
		t.Error("result false")
	}
}

func Test_Redis_SetEx(t *testing.T) {
	redis := NewRedisTest()

	result := redis.SetEx("setex", "asdfsdf", 60)

	if !result {
		t.Error("result false")
	}
}

func Test_Redis_Incr(t *testing.T) {
	redis := NewRedisTest()

	_, result := redis.Incr("incr")

	if !result {
		t.Error("result false")
	}
}

func Test_Redis_HSet(t *testing.T) {
	redis := NewRedisTest()

	result := redis.HSet("hset", "k1", "sdfsdf")

	if !result {
		t.Error("result false")
	}
}

func Test_Redis_HSetNX(t *testing.T) {
	redis := NewRedisTest()

	_, result := redis.HSetNX("hsetnx", "h1", "123123")

	if !result {
		t.Error("result false")
	}
}

func Test_Redis_Del(t *testing.T) {
	redis := NewRedisTest()

	result := redis.Del("tonyjt")
	if !result {
		t.Error("result false")
	}
}
func Test_Redis_HIncrby(t *testing.T) {
	redis := NewRedisTest()

	_, result := redis.HIncrby("hincr", "1", 1)

	if !result {
		t.Error("result false")
	}
}

func Test_Redis_HMSet(t *testing.T) {
	redis := NewRedisTest()

	datas := make(map[string]ModelRedisHello)

	datas["1"] = ModelRedisHello{HelloWord: "HelloWord1"}
	datas["2"] = ModelRedisHello{HelloWord: "HelloWord2"}

	result := redis.HMSet("hmset1", datas)

	if !result {
		t.Errorf("result false")
	}
}

func Test_Redis_ZAddM(t *testing.T) {
	redis := NewRedisTest()

	datas := make(map[int]int)

	datas[3] = 3
	datas[2] = 2
	datas[1] = 1

	result := redis.ZAddM("zaddm1", datas)

	if !result {
		t.Errorf("result false")
	}
}

func Test_Redis_ZRem(t *testing.T) {
	redis := NewRedisTest()

	result := redis.ZRem("zaddm1", 2, 3)

	if !result {
		t.Errorf("result false")
	}
}
func Test_Redis_MSet(t *testing.T) {
	redis := NewRedisTest()
	value := make(map[string]ModelRedisHello)
	value["mset1"] = ModelRedisHello{HelloWord: "1"}
	value["mset2"] = ModelRedisHello{HelloWord: "2"}
	value["mset3"] = ModelRedisHello{HelloWord: "3"}
	result := redis.MSet(value)
	if !result {
		t.Error("result false")
	}
}

func Test_Redis_MGet(t *testing.T) {
	redis := NewRedisTest()

	value, err := redis.MGet("mset1", "mset2", "mset4", "mset3")

	if err != nil {
		t.Errorf("result false:%s", err.Error())
	} else if len(value) != 4 {
		t.Errorf("len is < 4:%d", len(value))
	}

}

func Test_Redis_HDel(t *testing.T) {
	redis := NewRedisTest()

	key := "hmset1"
	result := redis.HDel(key, "1", "2")

	if !result {
		t.Error("result false")
	}
}
func Test_Redis_HMGet(t *testing.T) {
	redis := NewRedisTest()

	data, err := redis.HMGet("hmset1", "1", "2", "3")

	if err != nil {
		t.Errorf("result false:%s", err.Error())
	} else if len(data) != 3 {
		t.Errorf("len is lt :%d", len(data))
	}
}

func Test_Redis_ZRevRange(t *testing.T) {
	redis := NewRedisTest()

	data, err := redis.ZRevRange("zaddm", 0, 1)

	if err != nil {
		t.Errorf("result false:%s", err.Error())
	} else {
		t.Errorf("result,len:%d,d:%v", len(data), data)
	}
}

type TestDaoRedis struct {
	DaoRedis
}

func NewRedisTest() *TestDaoRedis {

	componentDao := &TestDaoRedis{DaoRedis{KeyName: "test"}}

	return componentDao
}

func (c *TestDaoRedis) Set(name string, key string) bool {

	return c.DaoRedis.Set(name, key)
}

func (c *TestDaoRedis) Get(name string, key *string) bool {
	return c.DaoRedis.Get(name, key)
}
func (c *TestDaoRedis) MGet(keys ...string) ([]*ModelRedisHello, error) {

	var datas []*ModelRedisHello

	err := c.DaoRedis.MGet(keys, &datas)

	return datas, err
}

func (c *TestDaoRedis) Del(key string) bool {
	return c.DaoRedis.Del(key)
}

func (c *TestDaoRedis) HMSet(key string, value map[string]ModelRedisHello) bool {
	datas := make(map[string]interface{})

	for k, v := range value {
		datas[k] = v
	}
	return c.DaoRedis.HMSet(key, datas)
}

func (c *TestDaoRedis) HMGet(key string, fields ...string) ([]*ModelRedisHello, error) {
	var datas []*ModelRedisHello

	var args []interface{}

	for _, item := range fields {
		args = append(args, item)
		//datas = append(datas, &ModelRedisHello{})
	}
	err := c.DaoRedis.HMGet(key, args, &datas)

	return datas, err
}

func (c *TestDaoRedis) ZAddM(key string, value map[int]int) bool {
	datas := make(map[string]interface{})

	for k, v := range value {
		datas[strconv.Itoa(k)] = v
	}
	return c.DaoRedis.ZAddM(key, datas)
}

func (c *TestDaoRedis) ZRem(key string, data ...interface{}) bool {

	return c.DaoRedis.ZRem(key, data...)
}

func (c *TestDaoRedis) ZRevRange(key string, start int, end int) ([]ModelRedisHello, error) {

	var data []*ModelRedisHello

	err := c.DaoRedis.ZRevRange(key, start, end, &data)
	var value []ModelRedisHello
	if err == nil {

		for _, item := range data {
			if item != nil {
				value = append(value, *item)
			} else {
				value = append(value, ModelRedisHello{})
			}
		}
	}
	return value, err
}

func (c *TestDaoRedis) MSet(value map[string]ModelRedisHello) bool {
	datas := make(map[string]interface{})

	for k, v := range value {
		datas[k] = v
	}
	return c.DaoRedis.MSet(datas)
}
