package tgo

import (
	"testing"
)

type ModelHello struct {
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
