package tgo

import "testing"

type ConfigTest struct {
	Configs map[string]interface{}
}

func configAppGetTest() *ConfigTest {
	return &ConfigTest{map[string]interface{}{"Env": "idc", "UrlUserLogin": "http://user.haiziwang.com/user/CheckLogin"}}
}
func Test_ConfigGet(t *testing.T) {
	config := &ConfigTest{}

	defaultConfig := configAppGetTest()

	configGet("app", config, defaultConfig)

	if config == defaultConfig {
		t.Errorf("not find config file")
	} else {
		t.Errorf("%v", config)
	}
}
