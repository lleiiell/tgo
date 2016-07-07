package tgo

//验证签名

import (
	"github.com/gin-gonic/gin"
	"net/url"
	"strconv"
	"time"
	"sort"
	"strings"
)

var (
	SignAppSecretKey string
)

/*
sha1的签名算法
   appsecret = "std::string"
   signature = sha1(appsecret+"babybirthday=1457578839&city=南京市&mobile=15324893018&province=江苏省
&signtimestamp=1457578839&username=张三"+appsecret)

1. 签名: signature, 时间戳: timestamp
2. 参数列表按参数Key字典序升序排列
3. 编码使用 UTF-8
*/

func UtilSignCheckSign(c *gin.Context) bool {
	SignSwitch := ConfigAppGetString("SignSwitch", "1")
	if SignSwitch == "0" {
		return true
	}
	SignAppSecretKey = ConfigAppGetString("AppSecretKey", "")
	ps := UtilRequestGetAllParams(c)
	if (!utilSignCheckSignTimestamp(c)) {
		return false
	}
	signTimestamp, _ := c.Cookie("signtimestamp")
	ps["signtimestamp"] = []string{signTimestamp}
	sortedParams := utilSignGetSortUpParamsString(ps)
	signString := SignAppSecretKey + sortedParams + SignAppSecretKey
	signature := UtilCryptoSha1(signString)
	signCookie, _ := c.Cookie("signature")

	if signature == signCookie {
		return true
	}

	return false
}

//升序排序的参数拼接的字符串
func utilSignGetSortUpParamsString(ps url.Values) string {
	psKey := []string{}
	for k, _ := range ps {
		psKey = append(psKey, k)
	}
	sort.Strings(psKey)
	ret := []string{}
	for _, v := range psKey {
		ret = append(ret, v + "=" + ps.Get(v))
	}
	return strings.Join(ret, "&")
}

func utilSignCheckSignTimestamp(c *gin.Context) bool {
	appLimitTime, _ := strconv.Atoi(ConfigAppGetString("AppAccessLimitTime", ""))
	if appLimitTime == 0 {
		return true
	}
	cookie, _ := c.Cookie("signtimestamp")
	signTimestamp, _ := strconv.Atoi(cookie)
	now := time.Now().Unix()
	if now < int64(appLimitTime + signTimestamp) {
		return true
	}

	return false
}