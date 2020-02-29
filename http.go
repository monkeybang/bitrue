package bitrue

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/spf13/cast"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	POST   = "POST"
	GET    = "GET"
	DELETE = "DELETE"
)

func HttpGetRequest(strUrl string, mapParams map[string]string) string {
	httpClient := &http.Client{}

	var strRequestUrl string
	if nil == mapParams {
		strRequestUrl = strUrl
	} else {
		strParams := Map2UrlQuery(mapParams)
		strRequestUrl = strUrl + "?" + strParams
	}

	// 构建Request, 并且按官方要求添加Http Header
	request, err := http.NewRequest("GET", strRequestUrl, nil)
	if nil != err {
		log.Println(strRequestUrl)
		log.Println(request)
		return err.Error()
	}
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.71 Safari/537.36")

	//	log.Println(strUrl)
	//log.Println(request)
	// 发出请求
	response, err := httpClient.Do(request)
	if nil != err {
		log.Print(err)
		return err.Error()
	}
	defer response.Body.Close()

	// 解析响应内容
	body, err := ioutil.ReadAll(response.Body)
	if nil != err {
		log.Print(err)
		return err.Error()
	}

	return string(body)
}

// 将map格式的请求参数转换为字符串格式的
// mapParams: map格式的参数键值对
// return: 查询字符串
func Map2UrlQuery(data map[string]string) string {
	var strParams string
	for key, value := range data {
		strParams += key + "=" + value + "&"
	}
	if 0 < len(strParams) {
		strParams = string([]rune(strParams)[:len(strParams)-1])
	}
	return strParams
}

func SignedRequest(method string, url string, params map[string]string) string {
	return SignedRequestWithKey(method, url, params, accessKey, secretKey)
}

func SignedRequestWithKey(method string, url string, params map[string]string, ak, sk string) string {
	var paramStr string
	if len(params) > 0 {
		keys, values := SortByKey(params)
		paramStr = Slice2UrlQuery(keys, values)
	}
	timestamp := cast.ToString(time.Now().UnixNano() / 1000000)
	if len(paramStr) > 0 {
		paramStr += "&timestamp=" + timestamp
	} else {
		paramStr += "timestamp=" + timestamp
	}

	signature := GetSignedWithSecretKey(paramStr, sk)
	paramStr += "&signature=" + signature

	request, err := http.NewRequest(method, url, strings.NewReader(paramStr))

	if err != nil {
		log.Println(err)
		return err.Error()
	}

	request.Header.Add("X-MBX-APIKEY", ak)
	//request.Header.Add("FC-ACCESS-TIMESTAMP", timestamp)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if nil != err {
		log.Println("HTTP request error， url:", url, "params:", paramStr, "info:", err)
		return ""
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if nil != err {
		log.Println("HTTP response data read error， url:", url, "params:", paramStr, "info:", err)
		return err.Error()
	}

	return string(body)
}

func Slice2UrlQuery(keys []string, values []string) string {
	var strParams string
	for i, key := range keys {
		value := values[i]
		strParams += key + "=" + value + "&"
	}
	if 0 < len(strParams) {
		strParams = string([]rune(strParams)[:len(strParams)-1])
	}
	return strParams
}

func SortByKey(mapValue map[string]string) (keys []string, values []string) {
	for key := range mapValue {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		//value := url.QueryEscape(mapValue[key])
		value := mapValue[key]
		values = append(values, value)
	}
	return
}

func GetSigned(sigStr string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(sigStr))
	return hex.EncodeToString(mac.Sum(nil))
}

func GetSignedWithSecretKey(str, sk string) string {
	mac := hmac.New(sha256.New, []byte(sk))
	mac.Write([]byte(str))
	return hex.EncodeToString(mac.Sum(nil))
}
