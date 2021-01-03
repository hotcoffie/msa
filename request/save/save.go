package save

import (
	"fmt"
	"github.com/levigross/grequests"
	"github.com/pkg/errors"
	"msa/common/conf"
	"msa/request"
	"msa/request/vscode"
	"net/http"
	"strconv"
	"time"
)

func Dail(passTime, JSESSIONID string) (*request.Result, error) {
	url := "https://www.sh.msa.gov.cn/zwzx/applyVtsDeclare1/saveVts/"
	if conf.Data.Active == conf.ActiveProd {
		url = "https://www.sh.msa.gov.cn/zwzx/applyVtsDeclare1/save/"
	}
	contentType, requestBody, err := conf.CreateRequestBody(passTime, vscode.Get(JSESSIONID))
	if err != nil {
		return nil, errors.WithMessage(err, "获取请求参数")
	}
	headers := map[string]string{
		"Accept":           "application/json, text/javascript, */*; q=0.01",
		"Accept-Encoding":  "gzip, deflate, br",
		"Accept-Language":  "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		"Connection":       "keep-alive",
		"Content-Length":   strconv.Itoa(requestBody.Len()),
		"Content-Type":     contentType,
		"Host":             "www.sh.msa.gov.cn",
		"Origin":           "https://www.sh.msa.gov.cn",
		"Referer":          "https://www.sh.msa.gov.cn/zwzx/applyVtsDeclare1",
		"Sec-Fetch-Dest":   "empty",
		"Sec-Fetch-Mode":   "cors",
		"Sec-Fetch-Site":   "same-origin",
		"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66",
		"X-Requested-With": "XMLHttpRequest",
	}
	opts := &grequests.RequestOptions{
		RequestTimeout: 5 * time.Second,
		Headers:        headers,
		Cookies:        []*http.Cookie{{Name: "isRead", Value: "y"}, {Name: "JSESSIONID", Value: JSESSIONID}},
		RequestBody:    requestBody,
	}
	res, err := grequests.Post(url, opts)
	if err != nil {
		return nil, errors.WithMessage(err, "请求")
	}
	defer res.Close()
	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("状态码：%d", res.StatusCode))
	}
	result := &request.Result{}
	err = res.JSON(result)
	if err != nil {
		return nil, errors.WithMessage(err, "解码")
	}
	return result, nil
}

func GetTime(JSESSIONID string) (int, error) {
	url := "https://www.sh.msa.gov.cn/zwzx/applyVtsDeclare1/getSeconds"
	headers := map[string]string{
		"Accept":           "*/*",
		"Accept-Encoding":  "gzip, deflate, br",
		"Accept-Language":  "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		"Connection":       "keep-alive",
		"Content-Length":   "0",
		"Host":             "www.sh.msa.gov.cn",
		"Origin":           "https://www.sh.msa.gov.cn",
		"Referer":          "https://www.sh.msa.gov.cn/zwzx/applyVtsDeclare1",
		"Sec-Fetch-Dest":   "empty",
		"Sec-Fetch-Mode":   "cors",
		"Sec-Fetch-Site":   "same-origin",
		"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66",
		"X-Requested-With": "XMLHttpRequest",
	}
	opts := &grequests.RequestOptions{
		RequestTimeout: 5 * time.Second,
		Headers:        headers,
		Cookies:        []*http.Cookie{{Name: "isRead", Value: "y"}, {Name: "JSESSIONID", Value: JSESSIONID}},
	}
	res, err := grequests.Get(url, opts)
	if err != nil {
		return 0, errors.WithMessage(err, "请求")
	}
	defer res.Close()
	if res.StatusCode != 200 {
		return 0, errors.New(fmt.Sprintf("状态码：%d", res.StatusCode))
	}
	num, err := strconv.Atoi(res.String())
	if err != nil {
		return 0, errors.WithMessage(err, "解析时间")
	}
	return num, nil
}
