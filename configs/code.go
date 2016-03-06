package configs

import (
	"strconv"
	"sync"
)

var (
	codeConfig    *CodeList
	codeConfigMux sync.Mutex
)

type CodeList struct {
	Codes map[string]string
}

func getCodeConfigs() {
	if codeConfig == nil || len(codeConfig.Codes) == 0 {

		codeConfigMux.Lock()

		defer codeConfigMux.Unlock()

		codeConfig = &CodeList{}

		defaultData := getDefaultCodeConfig()

		getConfigs("code", codeConfig, defaultData)
	}
}

func getDefaultCodeConfig() *CodeList {
	return &CodeList{Codes: map[string]string{"0": "success"}}
}

func GetCodeMessage(code int) string {

	getCodeConfigs()

	msg, exists := codeConfig.Codes[strconv.Itoa(code)]

	if !exists {
		return "system error"
	}
	return msg
}
