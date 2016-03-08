package tgo

import (
	"strconv"
	"sync"
)

var (
	codeConfig    *ConfigCodeList
	codeConfigMux sync.Mutex
)

type ConfigCodeList struct {
	Codes map[string]string
}

func configCodeInit() {
	if codeConfig == nil || len(codeConfig.Codes) == 0 {

		codeConfigMux.Lock()

		defer codeConfigMux.Unlock()

		codeConfig = &ConfigCodeList{}

		defaultData := configCodeGetDefault()

		configGet("code", codeConfig, defaultData)
	}
}

func configCodeGetDefault() *ConfigCodeList {
	return &ConfigCodeList{Codes: map[string]string{"0": "success"}}
}

func ConfigCodeGetMessage(code int) string {

	configCodeInit()

	msg, exists := codeConfig.Codes[strconv.Itoa(code)]

	if !exists {
		return "system error"
	}
	return msg
}
