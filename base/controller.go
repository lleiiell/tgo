package controllers

import (
	"configs"
	"fmt"
	"github.com/gin-gonic/gin"
	"modules/service"
	"pconst"
	"strconv"
	"util"
)

func GetParam(c *gin.Context, key string) string {
	if c.Request.Method == "GET" {
		return c.Query(key)
	}
	return c.PostForm(key)

}

func getLoginUid(c *gin.Context) (int64, int) {

	uid, result := GetUidFromCookie(c)

	if result {
		var skey string
		skey, result = GetSkeyFromCookie(c)
		if result {
			if service.CheckLogin(uid, skey) {
				return uid, 0
			}
		}
	}
	return 0, pconst.ERROR_AUTH_USER
}

func ReturnJson(c *gin.Context, code int, data interface{}) {

	msg := configs.GetCodeMessage(code)

	util.ReturnJson(c, code, msg, data)
}

func GetCid(c *gin.Context) (int, bool) {
	return getInt(c, "cid")
}

func GetCids(c *gin.Context) []int {
	key := "cids"
	cids := GetParam(c, key)

	cidArray := util.SplitToIntArray(cids, ",")

	return cidArray
}

func GetObjectId(c *gin.Context) string {
	key := "object_id"

	return GetParam(c, key)
}

func GetUid(c *gin.Context) (int64, bool) {

	var str string

	key := "uid"

	str = GetParam(c, key)

	uid, err := strconv.ParseInt(str, 10, 64)

	if err != nil {
		return 0, false
	}
	return uid, true
}

func GetFlag(c *gin.Context) (int, bool) {
	return getInt(c, "flag")
}

func GetUidFromCookie(c *gin.Context) (int64, bool) {

	cookieUid, err := c.Request.Cookie("uid")

	if err != nil {
		fmt.Printf("uid err is :%s", err.Error())

		return 0, false
	}

	uid, err := strconv.ParseInt(cookieUid.Value, 10, 64)

	if err != nil {
		return 0, false
	}
	return uid, true
}

func GetRecordId(c *gin.Context) (int64, bool) {

	var str string
	key := "record_id"
	str = GetParam(c, key)

	id, err := strconv.ParseInt(str, 10, 64)

	if err != nil {
		return 0, false
	}
	return id, true
}

func GetSuccess(c *gin.Context) (bool, bool) {

	var str string
	key := "success"
	str = GetParam(c, key)

	id, err := strconv.ParseBool(str)

	if err != nil {
		return false, false
	}
	return id, true
}

func getInt(c *gin.Context, key string) (int, bool) {
	var strId string

	strId = GetParam(c, key)

	id, err := strconv.Atoi(strId)

	if err != nil {
		return 0, false
	}
	return id, true
}

func GetSkeyFromCookie(c *gin.Context) (string, bool) {

	skey, err := c.Request.Cookie("skey")

	if err != nil {
		fmt.Printf("skey err is :%s", err.Error())

		return "", false
	}
	return skey.Value, true
}
