package util

import (
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"path/filepath"
	"sync"
)

var (
	logMux sync.Mutex

	logInitFlag bool
)

func initLog() {

	if !logInitFlag {
		logMux.Lock()

		defer logMux.Unlock()
		if !logInitFlag {
			filePath, err := filepath.Abs("configs/log.xml")

			if err == nil {
				l4g.LoadConfiguration(filePath)
				logInitFlag = true
			}
		}
	}
}
func LogError(msg interface{}) {

	initLog()

	//defer l4g.Close()

	l4g.Error(msg)

}

func LogErrorf(format string, a ...interface{}) {

	msg := fmt.Sprintf(format, a...)

	initLog()

	//defer l4g.Close()

	l4g.Error(msg)

}

func LogInfo(msg interface{}) {

	initLog()

	//defer l4g.Close()

	l4g.Info(msg)

}
func LogInfof(format string, a ...interface{}) {

	msg := fmt.Sprintf(format, a...)

	initLog()

	//defer l4g.Close()

	l4g.Info(msg)

}

func LogDebug(msg interface{}) {

	initLog()

	//defer l4g.Close()

	l4g.Debug(msg)

}
func LogDebugf(format string, a ...interface{}) {

	msg := fmt.Sprintf(format, a...)

	initLog()

	//defer l4g.Close()

	l4g.Debug(msg)
}
