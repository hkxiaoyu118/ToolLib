package main

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)
import "fmt"

const g_ipFormat = "https://www.ip.cn/api/index?ip=%s&type=1"
const g_regStr = `((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)`

//{"rs":1,"code":0,"address":"中国  四川省 成都市 电信","ip":"182.150.63.184","isDomain":0}
type IpResult struct {
	RS       string `json:"rs"`
	Code     int    `json:"code"`
	Address  string `json:"address"`
	IP       string `json:"ip"`
	IsMomain string `json:"isDomain"`
}

//读取文件的内容
func readStringFromFile(filePath string) string {
	bytes, err := ioutil.ReadFile(filePath)
	if err == nil {
		return string(bytes)
	} else {
		panic(err.Error())
	}
}

func appendStringToFile(filePath string, content string) {
	fd, _ := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	buf := []byte(content + "\n")
	_, _ = fd.Write(buf)
	_ = fd.Close()
}

func HttpGet(url string) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}
	return client.Get(url)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: ipshow /var/xxx/log.txt log.txt(choice)")
		os.Exit(0)
	}

	filePath := os.Args[1]
	ipArray := make(map[string]int)
	msg := readStringFromFile(filePath)
	r := regexp.MustCompile(g_regStr)
	matched := r.FindAllStringSubmatch(msg, -1)
	for _, item := range matched {
		ipString := item[0]
		if strings.HasPrefix(ipString, "192.") == false && strings.HasPrefix(ipString, "172.") == false && strings.HasPrefix(ipString, "10.") == false && strings.HasPrefix(ipString, "127.") == false {
			if _, ok := ipArray[ipString]; ok {
				count := ipArray[ipString] + 1
				ipArray[ipString] = count
			} else {
				ipArray[ipString] = 1
			}
		}
	}

	logFilePath := ""
	if len(os.Args) == 3 {
		logFilePath = os.Args[2]
	}

	for key, value := range ipArray {
		urlString := fmt.Sprintf(g_ipFormat, key)
		response, err := HttpGet(urlString)
		if err == nil {
			defer response.Body.Close()
			data, err := ioutil.ReadAll(response.Body)
			ipResult := new(IpResult)
			if err == nil {
				_ = json.Unmarshal(data, ipResult)
			}
			msg := key + " " + ipResult.Address + " count:" + strconv.Itoa(value)
			if logFilePath == "" {
				fmt.Println(msg)
			} else {
				appendStringToFile(logFilePath, msg)
			}
		} else {
			msg := key + " 获取IP地址归属地发生错误" + " count:" + strconv.Itoa(value)
			if logFilePath == "" {
				fmt.Println(msg)
			} else {
				appendStringToFile(logFilePath, msg)
			}
		}
	}
}
