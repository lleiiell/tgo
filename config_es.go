package tgo

import (
	"sync"
)

var (
	esConfigMux sync.Mutex

	esConfig *ConfigES
)

type ConfigES struct {
	Address             []string
	ClientPool          bool
	ClientMaxTotal      int
	ClientMaxIdle       int
	ClientMinIdle       int
	ClientLifo          bool
	ClientMaxWaitMillis int64
	Timeout             int
	TransportMaxIdel    int
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

	return &ConfigES{Address: []string{"url"},
		ClientPool:       true,
		ClientMaxTotal:   20,
		ClientMinIdle:    2,
		ClientMaxIdle:    20,
		Timeout:          3000,
		TransportMaxIdel: 10}
}

func configESGetAddress() []string {
	configESInit()

	return esConfig.Address
}

func configESGet() *ConfigES {
	configESInit()

	return esConfig
}
