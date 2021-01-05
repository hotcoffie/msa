package vscode

import (
	"bytes"
	"fmt"
	"github.com/levigross/grequests"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

func Get(JSESSIONID string) string {
	url := `https://www.sh.msa.gov.cn/zwzx/views/image.jsp?ts=` + strconv.Itoa(time.Now().Second()*1000)
	opts := &grequests.RequestOptions{
		RequestTimeout: 30 * time.Second,
		Headers: map[string]string{
			"Accept":          "image/webp,image/apng,image/*,*/*;q=0.8",
			"Accept-Encoding": "gzip, deflate, br",
			"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
			"Connection":      "keep-alive",
			"Host":            "www.sh.msa.gov.cn",
			"Sec-Fetch-Dest":  "image",
			"Sec-Fetch-Mode":  "no-cors",
			"Sec-Fetch-Site":  "same-origin",
			"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66",
		},
		Cookies: []*http.Cookie{{Name: "isRead", Value: "y"}, {Name: "JSESSIONID", Value: JSESSIONID}},
	}
	for {
		code, err := getImg(url, opts)
		if err != nil {
			logrus.WithError(err).Debug("验证码")
			continue
		}
		return code
	}
}

func getImg(url string, opts *grequests.RequestOptions) (string, error) {
	res, err := grequests.Get(url, opts)
	if err != nil {
		return "", errors.WithMessage(err, "获取图片")
	}
	defer res.Close()
	if res.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("状态码：%d", res.StatusCode))
	}
	cmd := exec.Command("tesseract", "stdin", "stdout", "--dpi", "96", "--psm", "10", "--oem", "3", "-c", "tessedit_char_whitelist=0123456789")
	var result = new(bytes.Buffer)
	cmd.Stdout = result
	cmd.Stdin = res.RawResponse.Body
	err = cmd.Run()
	if err != nil {
		return "", errors.WithMessage(err, "执行tesseract")
	}

	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		return "", errors.WithMessage(err, "正则处理验证码")
	}
	resultStr := result.String()
	code := reg.ReplaceAllString(resultStr, "")
	if len(code) != 4 {
		return "", errors.New("识别结果：" + code)
	}
	return code, nil
}
