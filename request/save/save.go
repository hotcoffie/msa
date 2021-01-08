package save

import (
	"fmt"
	"github.com/levigross/grequests"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"msa/common/conf"
	"msa/request/vo"
	"msa/request/vscode"
	"net/http"
	"strconv"
	"time"
)

var finishedMsgs = map[string]bool{
	"操作成功": true,
	"进出口申报同一条船舶一天只能申请一次！":  false,
	"进出口申报同一条船舶一天内只能申请一次！": false,
	"当前船舶已经申报成功,请勿重新提交！":   false,
	"当前时间点已被申报，是否选择排队？":    false,
	"当前时间点已被申报,是否选择排队？":    false,
}

func howLong() int64 {
	now := time.Now()
	times := fmt.Sprintf("%d-%02d-%02d 14:00", now.Year(), now.Month(), now.Day())
	limit, _ := time.Parse("2006-01-02 15:04", times)
	return int64(now.Sub(limit).Seconds())
}

func Dail(passTime, JSESSIONID string, log *logrus.Entry) {
	url := "https://www.sh.msa.gov.cn/zwzx/applyVtsDeclare1/saveVts/"
	if conf.Data.Active == conf.ActiveProd {
		url = "https://www.sh.msa.gov.cn/zwzx/applyVtsDeclare1/save/"
	}
	opts, err := generateSaveOpts(passTime, JSESSIONID)
	if err != nil {
		log.WithError(err).Error("组装请求参数失败")
		return
	}
	finished := false
	start := false
	for !finished {
		if start || conf.Data.Active == conf.ActiveDev {
			result, err := save(url, opts)
			if err != nil {
				log.WithError(err).Error("提交失败")
				finished = conf.Data.Active == conf.ActiveDev
				continue
			}
			_, finished = finishedMsgs[result.ResultDesc]
			log.WithField("result", result.ResultDesc).WithField("finished", finished).Info("提交成功")
		} else {
			serverResult, err := getTime(JSESSIONID)
			localResult := howLong()
			if err != nil {
				log.WithError(err).Error("获取申报开始时间失败")
				serverResult = localResult
			} else if serverResult > localResult {
				log.Error("使用相对更小的本地时间校准")
				serverResult = localResult
			}
			remainLog := log.WithField("remain", serverResult)
			if serverResult > 30 {
				remainLog.Debug("休息20秒")
				time.Sleep(20 * time.Second)
			} else {
				remainLog.Info("抢点开始")
				start = true
			}
		}
	}
}
func generateSaveOpts(passTime, JSESSIONID string) (*grequests.RequestOptions, error) {
	contentType, requestBody, err := conf.CreateRequestBody(passTime, vscode.Get(JSESSIONID))
	if err != nil {
		return nil, err
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
	return opts, nil
}
func save(url string, opts *grequests.RequestOptions) (*vo.Result, error) {
	res, err := grequests.Post(url, opts)
	if err != nil {
		return nil, errors.WithMessage(err, "请求")
	}
	defer res.Close()
	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("状态码：%d", res.StatusCode))
	}
	result := &vo.Result{}
	err = res.JSON(result)
	if err != nil {
		return nil, errors.WithMessage(err, "解码")
	}
	return result, nil
}

func getTime(JSESSIONID string) (int64, error) {
	url := "https://www.sh.msa.gov.cn/zwzx/applyVtsDeclare1/getSeconds"
	opts := &grequests.RequestOptions{
		RequestTimeout: 5 * time.Second,
		Headers: map[string]string{
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
		},
		Cookies: []*http.Cookie{{Name: "isRead", Value: "y"}, {Name: "JSESSIONID", Value: JSESSIONID}},
	}
	res, err := grequests.Get(url, opts)
	if err != nil {
		return 0, errors.WithMessage(err, "请求")
	}
	defer res.Close()
	if res.StatusCode != 200 {
		return 0, errors.New(fmt.Sprintf("状态码：%d", res.StatusCode))
	}
	num, err := strconv.ParseInt(res.String(), 10, 64)
	if err != nil {
		return 0, errors.WithMessage(err, "解析时间")
	}
	return num, nil
}
