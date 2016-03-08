package tgo

import (
	"encoding/json"
	//"fmt"
	"github.com/gin-gonic/gin"
)

func UtilResponseReturnJson(c *gin.Context, code int, msg string, model interface{}) {

	var rj interface{}

	if code == 0 {
		code = 1001
	}
	rj = gin.H{
		"code":    code,
		"message": msg,
		"data":    model}

	callback := c.Query("callback")

	if IsEmpty(callback) {

		c.JSON(200, rj)
	} else {
		b, err := json.Marshal(rj)
		if err != nil {
			LogErrorf("jsonp marshal error:%s", err.Error())
		} else {
			c.String(200, "%s(%s)", callback, string(b))
		}
	}
}

func UtilResponseReturnJsonFailed(c *gin.Context, code int, message string) {

	ResponseReturnJson(c, code, message, nil)
}

func UtilResponseReturnJsonSuccess(c *gin.Context, data interface{}) {
	ResponseReturnJson(c, 0, "", data)
}
