package main

import (
	"github.com/sirupsen/logrus"
	"msa/common/conf"
	_ "msa/common/logger"
	"msa/common/util"
	"msa/request"
	"strings"
	"sync"
	"time"
)

func main() {
	run()
}

func run() {
	date := time.Now().Add(time.Hour * 24).Format("2006-01-02 ")
	logrus.Infof("抢点日期：%s", date)
	logrus.Infof("运行模式：%s", conf.Data.Active)
	logrus.Infof("当前账号：%s", conf.Data.Username)
	logrus.Infof("时间列表：%s", conf.Data.Points)
	points := strings.Split(conf.Data.Points, ",")
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
