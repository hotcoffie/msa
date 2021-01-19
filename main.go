package main

import (
	"github.com/sirupsen/logrus"
	"io"
	"msa/common/conf"
	_ "msa/common/logger"
	"msa/common/util"
	"msa/request"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const logPath = "logs"

func main() {
	logFile := getLogFile()
	defer logFile.Close()
	fileAndStdoutWriter := io.MultiWriter(logFile, os.Stdout)
	logrus.SetOutput(fileAndStdoutWriter)

	run()
}

func getLogFile() *os.File {
	logFileName := filepath.Join(logPath, "run"+strconv.FormatInt(time.Now().Unix(), 10)+".log")
	f, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		logrus.Fatal("打开日志文件：", err)
	}
	return f
}

func run() {
	date := time.Now().Add(time.Hour * 24).Format("2006-01-02 ")
	logrus.Infof("抢点日期：%s", date)
	logrus.Infof("运行模式：%s", conf.Data.Active)
	points := strings.Split(conf.Data.Points, " ")
	wn := len(points) * conf.Data.ThreadNum
	logrus.Infof("线程总数：%d", wn)

	wg := &sync.WaitGroup{}
	wg.Add(wn)
	for i := 0; i < conf.Data.ThreadNum; i++ {
		// 每组携程，都随机打乱抢点顺序
		util.RandSlice(points)
		for _, passTime := range points {
			go request.Dail(date+passTime, i, wg)
		}
	}
	wg.Wait()
}
