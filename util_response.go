package tgo

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UtilResponseReturnJson(c *gin.Context, code int, model interface{}) {

	if code == 0 {
		code = 1001
	}
	msg := ConfigCodeGetMessage(code)

	UtilResponseReturnJsonWithMsg(c, code, msg, model)
}

func UtilResponseReturnJsonWithMsg(c *gin.Context, code int, msg string, model interface{}) {

	var rj interface{}

	//添加结果
	if code == 1001 {
		c.Set("result", true)
	} else {
		c.Set("result", false)
	}
	rj = gin.H{
		"code":    code,
		"message": msg,
		"data":    model}

	callback := c.Query("callback")

	if UtilIsEmpty(callback) {

		c.JSON(200, rj)
	} else {
		b, err := json.Marshal(rj)
		if err != nil {
			UtilLogErrorf("jsonp marshal error:%s", err.Error())
		} else {
			c.String(200, "%s(%s)", callback, string(b))
		}
	}
}

func UtilResponseReturnJsonFailed(c *gin.Context, code int) {

	UtilResponseReturnJson(c, code, nil)
}

func UtilResponseReturnJsonSuccess(c *gin.Context, data interface{}) {
	UtilResponseReturnJson(c, 0, data)
}

func UtilResponseRedirect(c *gin.Context, url string) {
	c.Redirect(http.StatusMovedPermanently, url)
}
