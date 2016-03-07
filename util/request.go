package util

import (
	"github.com/gin-gonic/gin"
	"github.com/tonyjt/tgo/util"
)

func RequestGetParam(c *gin.Context, key string) string {
	if c.Request.Method == "GET" {
		return c.Query(key)
	}
	return c.PostForm(key)
}
