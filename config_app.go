package tgo

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

var (
	appConfigMux sync.Mutex

	appConfig *ConfigApp
)

type ConfigApp struct {
	Configs map[string]interface{}
}

func configAppInit() {

	if appConfig == nil || len(appConfig.Configs) == 0 {

		appConfigMux.Lock()
		defer appConfigMux.Unlock()

		if appConfig == nil {
			appConfig = &ConfigApp{}

			defaultConfig := configAppGetDefault()

			configGet("app", appConfig, defaultConfig)
		}
	}
}
func configAppClear() {
	appConfigMux.Lock()
	defer appConfigMux.Unlock()

	appConfig = nil
}

func configAppGetDefault() *ConfigApp {
	return &ConfigApp{map[string]interface{}{"Env": "idc", "UrlUserLogin": "http://user.haiziwang.com/user/CheckLogin"}}
}
func ConfigAppGetString(key string, defaultConfig string) string {

	config := ConfigAppGet(key)

	var configStr string
	if config != nil {
		configStr = config.(string)
	}

	if UtilIsEmpty(configStr) {
		configStr = defaultConfig
	}
	return configStr
}

func ConfigAppGet(key string) interface{} {

	configAppInit()

	config, exists := appConfig.Configs[key]

	if !exists {
		return nil
	}
	return config
}

func ConfigAppFailoverGet(key string) (string, error) {

	var server string

	var err error

	failoverConfig := ConfigAppGet(key)

	if failoverConfig == nil {
		err = errors.New(fmt.Sprintf("config %s is null", key))
	} else {

		failoverUrl := failoverConfig.(string)

		if UtilIsEmpty(failoverUrl) {
			err = errors.New(fmt.Sprintf("config %s is empty", key))
		} else {
			failoverArray := strings.Split(failoverUrl, ",")

			randomMax := len(failoverArray)
			if randomMax == 0 {
				err = errors.New(fmt.Sprintf("config %s is empty", key))
			} else {
				var randomValue int
				if randomMax > 1 {

					rand.Seed(time.Now().UnixNano())

					randomValue = rand.Intn(randomMax)

				} else {
					randomValue = 0
				}
				server = failoverArray[randomValue]

			}
		}
	}
	return server, err
}

func ConfigEnvGet() string {
	strEnv := ConfigAppGet("Env")

	return strEnv.(string)
}

func ConfigEnvIsDev() bool {
	env := ConfigEnvGet()

	if env == "dev" {
		return true
	}
	return false
}

func ConfigEnvIsBeta() bool {
	env := ConfigEnvGet()

	if env == "beta" {
		return true
	}
	return false
}
