package tgo

import (
	"testing"
)

func TestConfigPool_configPoolGet(t *testing.T) {
	configPoolInit()

	poolConfig := configPoolGet("test")

	if poolConfig == nil {
		t.Error("config pool is null")
	}
}

func TestConfigPool_GetAddressRandom(t *testing.T) {
	configPoolInit()

	poolConfig := configPoolGet("test")

	_, err := poolConfig.GetAddressRandom()

	if err != nil {
		t.Errorf("get error:%v", err.Error())
	}
}
