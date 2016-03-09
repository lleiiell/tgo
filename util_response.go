package tgo

import (
	"encoding/json"
	//"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func UtilResponseReturnJson(c *gin.Context, code int, model interface{}) {

	var rj interface{}

	msg := ConfigCodeGetMessage(code)

	if code == 0 {
		code = 1001
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
