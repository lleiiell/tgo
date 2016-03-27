package tgo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func configGet(name string, data interface{}, defaultData interface{}) {

	//mux.Lock()

	//defer mux.Unlock()

	absPath, _ := filepath.Abs(fmt.Sprintf("configs/%s.json", name))

	file, err := os.Open(absPath)

	if err != nil {

		UtilLogError(fmt.Sprintf("open %s config file failed:%s", name, err.Error()))

		data = defaultData

	} else {

		defer file.Close()

		decoder := json.NewDecoder(file)

		errDecode := decoder.Decode(data)

		//if name == "cache" {
		//}
		if errDecode != nil {
			//记录日志
			UtilLogError(fmt.Sprintf("decode %s config error:%s", name, errDecode.Error()))
			data = defaultData
		}
	}
}

func ConfigReload() {
	configAppClear()
	configCacheClear()
	configCodeClear()
	configDbClear()
}
