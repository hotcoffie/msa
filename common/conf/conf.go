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
var info map[string]string

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
	Data.Points = strings.ReplaceAll(Data.Points, ",", " ")
	Data.Active = strings.ToLower(Data.Active)
	wg.Done()
}
func initInfo(wg *sync.WaitGroup) {
	file, err := os.Open("info.yml")
	if err != nil {
		logrus.WithError(err).Panic("读取申报信息")
	}
	defer file.Close()
	br := bufio.NewReader(file)
	info = make(map[string]string)
	for {
		line, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		lineStr := string(line)
		index := strings.Index(lineStr, ":")
		k := strings.TrimSpace(lineStr[:index])
		v := strings.TrimSpace(lineStr[index+1:])
		info[k] = v
	}
	wg.Done()
}

func CreateRequestBody(passTime, saveCode string) (string, *bytes.Buffer, error) {
	var requestBody = new(bytes.Buffer)
	w := multipart.NewWriter(requestBody)
	defer w.Close()
	for k, v := range info {
		if k == "savePostion" {
			tmp := make(map[string]interface{})
			err := json.Unmarshal([]byte(v), &tmp)
			if err != nil {
				return "", requestBody, errors.WithMessage(err, "解析savePostion字段")
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
		if err := w.WriteField(k, v); err != nil {
			return "", requestBody, errors.WithMessage(err, "构建multipart/form-data")
		}
	}
	return w.FormDataContentType(), requestBody, nil
}
