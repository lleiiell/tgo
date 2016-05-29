package tgo

import (
	"fmt"
	l4g "github.com/tonyjt/log4go"
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
func UtilLogError(msg interface{}) {

	initLog()

	//defer l4g.Close()
	l4g.Error(msg)

}

func UtilLogErrorf(format string, a ...interface{}) {

	msg := fmt.Sprintf(format, a...)

	initLog()

	l4g.Error(msg)

}

func UtilLogInfo(msg interface{}) {
	initLog()

	l4g.Info(msg)

}
func UtilLogInfof(format string, a ...interface{}) {

	msg := fmt.Sprintf(format, a...)

	initLog()

	//defer l4g.Close()

	l4g.Info(msg)

}

func UtilLogDebug(msg interface{}) {

	initLog()

	//defer l4g.Close()

	l4g.Debug(msg)

}
func UtilLogDebugf(format string, a ...interface{}) {

	msg := fmt.Sprintf(format, a...)

	initLog()

	//defer l4g.Close()

	l4g.Debug(msg)
}


type Log struct{

}

func NewUtilLog()*Log{
  return &Log{}
}

func (l *Log) Error(format string, a ...interface{}){
  UtilLogErrorf(format, a...)
}

func (l *Log) Info(format string, a ...interface{}){
  UtilLogInfof(format, a...)
}
