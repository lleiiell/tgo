package tgo

import (
	"testing"
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

func Test_HMSet(t *testing.T){
  redis:=NewRedisTest()

  datas := make(map[string]ModelRedisHello)

  datas["1"] = ModelRedisHello{HelloWord:"HelloWord1"}
  datas["2"] = ModelRedisHello{HelloWord:"HelloWord2"}


  _,err:=redis.HMSet("hmset1", datas)

  if err!=nil{
    t.Errorf("result false:%s",err.Error())
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

func (c *TestDaoRedis) HMSet(key string,value map[string]ModelRedisHello)(interface{},error){
  datas := make(map[string]interface{})

  for k,v := range value{
    datas[k] = v
  }
  return c.DaoRedis.doMSet("HMSET",key, datas)
}
