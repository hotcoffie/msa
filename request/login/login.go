package login

import (
	"crypto/md5"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/levigross/grequests"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"msa/common/conf"
	"msa/common/util"
	"msa/request/vo"
	"msa/request/vscode"
	"net/http"
	"time"
)

func Dail() string {
	publicKey, JSESSIONID := getPublicKey()
	loginCheck(publicKey, JSESSIONID)
	return JSESSIONID
}

func getPublicKey() (string, string) {
	url := "https://www.sh.msa.gov.cn/zwzx/loginOut"
	opts := &grequests.RequestOptions{
		RequestTimeout: 30 * time.Second,
		Headers: map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
			"Accept-Encoding":           "gzip, deflate, br",
			"Accept-Language":           "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
			"Cache-Control":             "max-age=0",
			"Connection":                "keep-alive",
			"Host":                      "www.sh.msa.gov.cn",
			"Origin":                    "https://www.sh.msa.gov.cn",
			"Referer":                   "https://www.sh.msa.gov.cn/zwzx/applyVtsDeclare1",
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "same-origin",
			"Sec-Fetch-User":            "?1",
			"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66",
			"Upgrade-Insecure-Requests": "1",
		},
	}
	for {
		if publicKey, JSESSIONID, err := parsePublicKey(url, opts); err != nil {
			logrus.WithError(err).Error("解析登录页面")
		} else {
			return publicKey, JSESSIONID
		}
	}
}

func parsePublicKey(url string, opts *grequests.RequestOptions) (string, string, error) {
	res, err := grequests.Get(url, opts)
	if err != nil {
		return "", "", errors.WithMessage(err, "请求")
	}
	defer res.Close()
	if res.StatusCode != 200 {
		return "", "", errors.New(fmt.Sprintf("状态码：%d", res.StatusCode))
	}
	doc, err := goquery.NewDocumentFromReader(res.RawResponse.Body)
	if err != nil {
		return "", "", errors.WithMessage(err, "读取")
	}
	publicKey, exists := doc.Find("#publicKey").Attr("value")
	if !exists {
		return "", "", errors.New("未找到publicKey元素")
	}
	var JSESSIONID string
	for _, v := range res.RawResponse.Cookies() {
		if v.Name == "JSESSIONID" {
			JSESSIONID = v.Value
			break
		}
	}
	return publicKey, JSESSIONID, nil
}

func loginCheck(publicKey, JSESSIONID string) {
	url := "https://www.sh.msa.gov.cn/zwzx/loginCheck"
	opts := &grequests.RequestOptions{
		RequestTimeout: 30 * time.Second,
		Headers: map[string]string{
			"Accept":           "*/*",
			"Accept-Encoding":  "gzip, deflate, br",
			"Accept-Language":  "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7",
			"Connection":       "keep-alive",
			"Content-Type":     "application/x-www-form-urlencoded; charset=UTF-8",
			"Origin":           "https://www.sh.msa.gov.cn",
			"Referer":          "https://www.sh.msa.gov.cn/zwzx/loginOut",
			"Sec-Fetch-Dest":   "empty",
			"Sec-Fetch-Mode":   "cors",
			"Sec-Fetch-Site":   "same-origin",
			"Host":             "www.sh.msa.gov.cn",
			"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66",
			"X-Requested-With": "XMLHttpRequest",
		},
		Cookies: []*http.Cookie{{Name: "isRead", Value: "y"}, {Name: "JSESSIONID", Value: JSESSIONID}},
	}
	logincode := vscode.Get(JSESSIONID)
	for {
		if result, err := tryLogin(url, opts, publicKey, logincode); err != nil {
			logrus.WithError(err).Error("登录")
		} else if result.ResultDesc == "帐号或密码错误！" || result.ResultDesc == "重试超过5次，请稍后再试！" {
			panic(result.ResultDesc)
		} else if result.ResultDesc == "验证码错误！" {
			logincode = vscode.Get(JSESSIONID)
		} else if result.ResultCode != 0 {
			logrus.Error(result.ResultDesc)
		} else {
			break
		}
	}
}

func tryLogin(url string, opts *grequests.RequestOptions, publicKey, logincode string) (*vo.Result, error) {
	params, err := loginParams(publicKey, logincode)
	if err != nil {
		return nil, errors.WithMessage(err, "加密参数")
	}
	opts.Data = params
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
	return result, err
}

func loginParams(publicKey, logincode string) (map[string]string, error) {
	params := map[string]string{
		"publicKey": publicKey,
		"username":  conf.Data.Username,
		"logincode": logincode,
	}
	key, err := util.NewEncrypt(publicKey)
	if err != nil {
		return nil, errors.WithMessage(err, "初始化公钥")
	}
	params["user_name"], err = key.RsaEncrypt(conf.Data.Username)
	if err != nil {
		return nil, errors.WithMessage(err, "用户名加密")
	}
	password := fmt.Sprintf("%x", md5.Sum([]byte(conf.Data.Password)))
	params["password"], err = key.RsaEncrypt(password)
	if err != nil {
		return nil, errors.WithMessage(err, "密码加密")
	}
	params["login_code"], err = key.RsaEncrypt(logincode)
	if err != nil {
		return nil, errors.WithMessage(err, "验证码加密")
	}
	return params, nil
}
