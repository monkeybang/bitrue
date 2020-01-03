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
	var paramStr string
	if len(params) > 0 {
		keys, values := SortByKey(params)
		paramStr = Slice2UrlQuery(keys, values)
	}
	timestamp := cast.ToString(time.Now().UnixNano() / 1000000)
	var sigStr string
	//if method == GET {
	//	if len(paramStr) > 0 {
	//		url = url + "?" + paramStr
	//	}
	//	sigStr = method + url + timestamp
	//} else if method == POST {
	//	sigStr = paramStr + "&timestamp=" + timestamp
	//}
	if method == GET {
		//if len(paramStr) > 0 {
		//	url = url + "?" + paramStr
		//}
	}

	sigStr = "timestamp=" + timestamp
	if len(paramStr) != 0 {
		sigStr = paramStr + "&" + sigStr
	}

	signature := GetSigned(sigStr)
	d := sigStr + "&signature=" + signature

	//log.Println("url:",url)
	request, err := http.NewRequest(method, url, strings.NewReader(d))
	//log.Println(url)
	//log.Println(d)
	if err != nil {
		log.Println(err)
	}
	request.Header.Add("X-MBX-APIKEY", accessKey)
	//request.Header.Add("FC-ACCESS-TIMESTAMP", timestamp)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if nil != err {
		return err.Error()
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if nil != err {
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
