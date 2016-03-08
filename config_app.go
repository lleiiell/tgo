package tgo

import (
	"sync"
)

var (
	appConfigMux sync.Mutex

	appConfig *ConfigList
)

type ConfigList struct {
	Configs map[string]interface{}
}

func getAppConfigs() {

	if appConfig == nil || len(appConfig.Configs) == 0 {

		appConfigMux.Lock()
		defer appConfigMux.Unlock()

		appConfig = &ConfigList{}

		defaultConfig := getDefaultAppConfig()

		getConfigs("app", appConfig, defaultConfig)
	}
}
func getDefaultAppConfig() *ConfigList {
	return &ConfigList{map[string]interface{}{"Env": "idc", "UrlUserLogin": "http://user.haiziwang.com/user/CheckLogin"}}
}
func GetAppConfig(key string) interface{} {

	getAppConfigs()

	config, exists := appConfig.Configs[key]

	if !exists {
		return nil
	}
	return config
}

func GetEnv() string {
	strEnv := GetAppConfig("Env")

	return strEnv.(string)
}

func IsEnvDev() bool {
	env := GetEnv()

	if env == "dev" {
		return true
	}
	return false
}

func IsEnvBeta() bool {
	env := GetEnv()

	if env == "beta" {
		return true
	}
	return false
}
