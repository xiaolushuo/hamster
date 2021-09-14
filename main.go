/*
 * @Descripttion:
 * @version:
 * @Author: 1314mylove
 * @Date: 2021-08-09 18:32:57
 * @LastEditors: 1314mylove
 * @LastEditTime: 2021-09-14 20:59:00
 */
package main

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/elazarl/goproxy"
)

var blackList = [...]string{"/api/userinfo"}
var n int = 0
var valueList list.List

// Convert json to string
func JsonToString(jsonStr string) (string, error) {
	v, err := jsonvalue.UnmarshalString(jsonStr)
	if err != nil {
		log.Printf("Unmarshal with error: %+v\n", err)
		return "valueList", err
	}
	n = n + 1
	// var newStr string
	for it := range v.IterObjects() {
		// log.Printf("第 %v 层 \n", n)
		// log.Println(it.K, "-", it.V)
		bb, _ := v.Get(it.K)
		// fmt.Println(bb.IsString())
		// log.Println("get value: ", bb.String())
		if strings.Contains(bb.String(), "{") {
			JsonToString(bb.String())
		} else {
			if bb.IsString() && len(bb.String()) > 0 {
				if strings.Contains(strings.ToLower(bb.String()), "token") {
					fmt.Println("PASS")
				} else {
					valueList.PushBack(bb.String())
				}
			}
		}
	}
	return "valueList", nil
}

func interfaceToString(value interface{}) string {
	// interface 转 string
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

func in(target string, str_array []string) bool {
	for _, element := range str_array {
		if strings.Contains(target, element) {
			// fmt.Println(strings.Contains(element, target))
			return false
		}

	}
	return true
}
func RequestBody(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	// fmt.Println("=========1", "\n", r.RequestURI)
	// fmt.Println("=========2", in(r.RequestURI, blackList[:]))
	if r.Method == "POST" && in(r.RequestURI, blackList[:]) {
		postType := r.Header.Get("Content-Type")
		// fmt.Println(postType)
		var bodyBytes []byte
		if r.Body != nil {
			// 获取数据流
			bodyBytes, _ = ioutil.ReadAll(r.Body)
		}
		var body_str string
		switch {
		// 判断post请求类型
		case strings.Contains(postType, "x-www-form-urlencoded"):
			// fmt.Println(string(bodyBytes))
			strList := strings.Split(string(bodyBytes), "&")
			var l1v string
			for i := 0; i < len(strList); i++ {
				l1k := strings.Split(strList[i], "=")[0]
				if strings.Contains(strings.ToLower(l1k), "token") {
					fmt.Println("PASS")
					l1v = strings.Split(strList[i], "=")[1]
				} else {
					l1v = "secscan_" + strings.Split(strList[i], "=")[1]
				}
				l1 := l1k + "=" + l1v
				if i == 0 {
					body_str = body_str + l1
				} else {
					body_str = body_str + "&" + l1
				}

			}
			fmt.Println(body_str, len(body_str))
			rl := len(body_str)
			// 动态设置Content-Length
			r.ContentLength = int64(rl)
			// 重新写入
			r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(body_str)))
		case strings.Contains(postType, "application/json"):
			// fmt.Println("101", postType)
			body_str := string(bodyBytes)
			JsonToString(body_str)
			// fmt.Println("999999", valueList)
			for e := valueList.Front(); e != nil; e = e.Next() {
				fmt.Printf("%v\n", e.Value)
				tt := interfaceToString(e.Value)
				body_str = strings.Replace(body_str, "\""+tt, "\"secscan_"+tt, 1)
			}
			log.Println(body_str)
			rl := len(body_str)
			// 动态设置Content-Length
			r.ContentLength = int64(rl)
			// 重新写入
			r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(body_str)))
		}

	}
	return r, nil
}

func main() {
	log.Printf("proxy Start success... \n")
	proxy := goproxy.NewProxyHttpServer()

	proxy.Verbose = true
	// 增加一层代理方便观察请求包，例如burp
	proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
		return url.Parse("http://127.0.0.1:8088")
	}
	proxy.OnRequest().DoFunc(RequestBody)
	proxy.OnResponse().DoFunc(
		func(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			return r
		})
	log.Fatal(http.ListenAndServe(":8080", proxy))

}
