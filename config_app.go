package tgo

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
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

	if config == nil {
		return defaultConfig
	} else {
		configStr := config.(string)

		if UtilIsEmpty(configStr) {
			configStr = defaultConfig
		}
		return configStr
	}
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

//ConfigAppGetSlice 获取slice配置，data必须是指针slice *[]，目前支持string,int,int64,bool,float64,float32
func ConfigAppGetSlice(key string, data interface{}) error {

	dataStrConfig := ConfigAppGetString(key, "")

	if UtilIsEmpty(dataStrConfig) {
		return errors.New("config is empty")
	}

	dataStrSlice := strings.Split(dataStrConfig, ",")

	dataType := reflect.ValueOf(data)

	//不是指针Slice
	if dataType.Kind() != reflect.Ptr || dataType.Elem().Kind() != reflect.Slice {
		return errors.New("reflect is not pt or slice")
	}

	dataSlice := dataType.Elem()

	//dataSlice = dataSlice.Slice(0, dataSlice.Cap())

	dataElem := dataSlice.Type().Elem()

	for _, dataStr := range dataStrSlice {

		if UtilIsEmpty(dataStr) {
			continue
		}
		var errConv error
		var item interface{}

		switch dataElem.Kind() {
		case reflect.String:
			item = dataStr
		case reflect.Int:
			item, errConv = strconv.Atoi(dataStr)
		case reflect.Int64:
			item, errConv = strconv.ParseInt(dataStr, 10, 64)
		case reflect.Bool:
			item, errConv = strconv.ParseBool(dataStr)
		case reflect.Float64:
			item, errConv = strconv.ParseFloat(dataStr, 64)
		case reflect.Float32:
			var item64, errConv = strconv.ParseFloat(dataStr, 32)
			if errConv == nil {
				item = float32(item64)
			}
		/*
			case reflect.Struct:
				var de
				errConv = json.Unmarshal([]byte(dataStr), de.Interface())*/
		default:
			return errors.New("type not support")
		}
		if errConv != nil {
			return errors.New(fmt.Sprintf("convert config failed error:%s", errConv.Error()))
		}

		dataSlice.Set(reflect.Append(dataSlice, reflect.ValueOf(item)))
	}
	return nil
}
