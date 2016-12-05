package tgo

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/go-stomp/stomp"
)

var (
	urlMqActive string
)

func UtilMQSend(key string, data interface{}) error {

	var urlMqActiveArray []string

	var err error

	var mqUrl string

	if urlMqActive == "" {
		urlMqActiveArray, err = utilMQGetUrlArray()

		if err != nil {
			return err
		}

		mqUrl = urlMqActiveArray[0]
	} else {
		mqUrl = urlMqActive
	}

	conn, errConn := utilMQDail(mqUrl)
	if errConn != nil {
		if urlMqActiveArray == nil {
			urlMqActiveArray, err = utilMQGetUrlArray()

			if err != nil {
				return err
			}
		}

		for _, mu := range urlMqActiveArray {

			if mu != mqUrl {
				conn, errConn = utilMQDail(mu)
				if errConn == nil {
					urlMqActive = mu
					break
				}
			}
		}

		if errConn != nil {
			return errConn
		}
	} else {
		urlMqActive = mqUrl
	}

	defer conn.Disconnect()

	var msg []byte

	if reflect.TypeOf(data).Kind() == reflect.String {
		msg = []byte(data.(string))
	} else {
		msg, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	err = conn.Send(key, "text/plain", msg)

	if err != nil {
		UtilLogErrorf("mq send msg error,sever:%s,key:%s,data:%v,error:%s",
			mqUrl, key, data, err.Error())
	}
	return err
}

func utilMQGetUrlArray() ([]string, error) {
	var urlMqActiveArray []string
	var err error
	mqUrl := ConfigAppGet("UrlMqServer").(string)

	if UtilIsEmpty(mqUrl) {
		err = errors.New("config UrlMqServer is empty")
	} else {
		urlMqActiveArray = strings.Split(mqUrl, ",")

		if len(urlMqActiveArray) == 0 {
			err = errors.New("config UrlMqServer is empty")
		}
	}

	return urlMqActiveArray, err
}

func utilMQDail(url string) (*stomp.Conn, error) {
	conn, err := stomp.Dial("tcp", url)
	if err != nil {
		UtilLogErrorf("cannot connect to server:%s,error:%s", url, err.Error())
	}
	return conn, err
}
