package conf

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"strings"
	"sync"
)

const ActiveDev = "dev"
const ActiveProd = "prod"

type entity struct {
	Active    string
	Username  string
	Password  string
	Points    string
	ThreadNum int `yaml:"threadNum"`
}

var Data entity
var info map[string]interface{}

func init() {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	initConf(wg)
	initInfo(wg)
	wg.Wait()
}
func initConf(wg *sync.WaitGroup) {
	yamlFile, err := ioutil.ReadFile("conf.yml")
	if err != nil {
		logrus.WithError(err).Panic("读取配置文件")
	}

	err = yaml.Unmarshal(yamlFile, &Data)
	if err != nil {
		logrus.WithError(err).Panic("解析配置文件")
	}
	Data.Active = strings.ToLower(Data.Active)
	if Data.Active == ActiveDev {
		Data.ThreadNum = 1
	}
	wg.Done()
}
func initInfo(wg *sync.WaitGroup) {
	file, err := os.Open("info.yml")
	if err != nil {
		logrus.WithError(err).Panic("读取申报信息")
	}
	defer file.Close()
	br := bufio.NewReader(file)
	info = make(map[string]interface{})
	for {
		line, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		lineStr := string(line)
		index := strings.Index(lineStr, ":")
		k := strings.TrimSpace(lineStr[:index])
		v := strings.TrimSpace(lineStr[index+1:])
		if k != "savePostion" {
			info[k] = v
			continue
		}
		mapV := map[string]string{}
		err = json.Unmarshal([]byte(v), &mapV)
		if err != nil {
			logrus.WithError(err).Panic("解析申报信息savePostion字段")
		}
		info[k] = mapV
	}
	wg.Done()
}

func CreateRequestBody(passTime, saveCode string) (string, *bytes.Buffer, error) {
	var requestBody = new(bytes.Buffer)
	w := multipart.NewWriter(requestBody)
	defer w.Close()
	for k, v := range info {
		if k == "savePostion" {
			tmp, ok := v.(map[string]string)
			if !ok {
				return "", requestBody, errors.New("savePostion字段类型异常")
			}
			tmp["passTime"] = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:00", passTime)))
			m, err := json.Marshal(tmp)
			if err != nil {
				return "", requestBody, errors.WithMessage(err, "序列化savePostion字段")
			}
			v = string(m)
		} else if k == "saveCode" {
			v = saveCode
		}
		strV, ok := v.(string)
		if !ok {
			return "", requestBody, errors.New(k + "字段类型异常")
		}
		if err := w.WriteField(k, strV); err != nil {
			return "", requestBody, errors.WithMessage(err, "构建multipart/form-data")
		}
	}
	return w.FormDataContentType(), requestBody, nil
}
