package tgo

import (
	"sync"
)

var (
	esConfigMux sync.Mutex

	esConfig *ConfigES
)

type ConfigES struct {
	Address         []string
}


func configESInit() {

	if esConfig == nil || len(esConfig.Address) == 0 {

		esConfigMux.Lock()

		defer esConfigMux.Unlock()

		if esConfig == nil || len(esConfig.Address) == 0 {
			esConfig = &ConfigES{}

			defaultESConfig := configESGetDefault()

			configGet("es", esConfig, defaultESConfig)
		}
	}
}

func configESClear() {
	esConfigMux.Lock()

	defer esConfigMux.Unlock()

	esConfig = nil
}

func configESGetDefault() *ConfigES {

	return &ConfigES{Address:[]string{"http://172.172.177.16:9200"}}
}

func configESGetAddress() []string{
  configESInit()

  return esConfig.Address
}
