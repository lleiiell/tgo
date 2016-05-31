package tgo

import (
	"testing"
	"strconv"
)

type ModelRedisHello struct {
  HelloWord string
}

func Test_Call(t *testing.T) {
	redis := NewRedisTest()

	redis.Set("tonyjt", "1234567")

	var data string
	result := redis.Get("tonyjt", &data)

	if !result {
		t.Error("result false")
	}
}

func Test_Del(t *testing.T){
	redis := NewRedisTest()

	result:=redis.Del("tonyjt")
	if !result {
		t.Error("result false")
	}
}

func Test_HMSet(t *testing.T){
  redis:=NewRedisTest()

  datas := make(map[string]ModelRedisHello)

  datas["1"] = ModelRedisHello{HelloWord:"HelloWord1"}
  datas["2"] = ModelRedisHello{HelloWord:"HelloWord2"}


  result:=redis.HMSet("hmset1", datas)

	if !result{
    t.Errorf("result false")
  }
}

func Test_ZAddM(t *testing.T){
	redis:=NewRedisTest()

  datas := make(map[int]int)

  datas[3] = 3
  datas[2] = 2
	datas[1] = 1

  result:=redis.ZAddM("zaddm1", datas)

  if !result{
    t.Errorf("result false")
  }
}

func Test_ZRem(t *testing.T){
	redis:=NewRedisTest()

  result:=redis.ZRem("zaddm1", 1)

  if !result{
    t.Errorf("result false")
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

func (c *TestDaoRedis) Del(key string)bool{
	return c.DaoRedis.Del(key)
}

func (c *TestDaoRedis) HMSet(key string,value map[string]ModelRedisHello)bool{
  datas := make(map[string]interface{})

  for k,v := range value{
    datas[k] = v
  }
  return c.DaoRedis.HMSet(key, datas)
}

func (c *TestDaoRedis) ZAddM(key string,value map[int]int)bool{
	datas := make(map[string]interface{})

  for k,v := range value{
    datas[strconv.Itoa(k)] = v
  }
  return c.DaoRedis.ZAddM(key, datas)
}

func (c *TestDaoRedis) ZRem(key string, data ...interface{})bool{

	return c.DaoRedis.ZRem(key, data...)
}
