package tgo

import (
	"github.com/gin-gonic/gin"
)

func UtilRequestGetParam(c *gin.Context, key string) string {
	if c.Request.Method == "GET" {
		return c.Query(key)
	}
	return c.PostForm(key)
}
