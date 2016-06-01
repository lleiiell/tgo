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

  result:=redis.ZRem("zaddm1", 2,3)

  if !result{
    t.Errorf("result false")
  }
}
func Test_MSet(t *testing.T){
	redis:=NewRedisTest()
	value:=make(map[string]ModelRedisHello)
	value["mset1"] = ModelRedisHello{HelloWord:"1"}
	value["mset2"] = ModelRedisHello{HelloWord:"2"}
	value["mset3"] = ModelRedisHello{HelloWord:"3"}
	result:= redis.MSet(value)
	if !result{
		t.Error("result false")
	}
}

func Test_MGet(t *testing.T){
	redis:=NewRedisTest()

	data,err:=redis.MGet("mset1","mset2","mset3","mset4")


  if err!=nil{
    t.Errorf("result false:%s",err.Error())
  }else if len(data)!=3{
		t.Errorf("result false,len:%d",len(data))
	}else{
		for k,v:=range data{
			t.Errorf("k:%s,v:%v", k,v)
		}
	}
}

func Test_HDel(t *testing.T){
	redis:=NewRedisTest()

	key := "hmset1"
	result:=redis.HDel(key, "1","2")

	if !result{
		t.Error("result false")
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
func (c *TestDaoRedis) MGet(keys ...string)(map[string]interface{},error){

	data := ModelRedisHello{}
	value,err:= c.DaoRedis.doMGet("MGET","", data,keys...)

	if err==nil{
		return value,nil/*
		datas := make(map[string]ModelRedisHello)

		for k,v:=range value{
				datas[k]= v.(ModelRedisHello)
		}
		return datas,nil*/
	}
	return nil,err
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

func (c *TestDaoRedis) MSet(value map[string]ModelRedisHello) bool{
	datas := make(map[string]interface{})

  for k,v := range value{
    datas[k] = v
  }
	return c.DaoRedis.MSet(datas)
}
