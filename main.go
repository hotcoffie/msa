package main

import (
	"github.com/sirupsen/logrus"
	"msa/common/conf"
	_ "msa/common/conf"
	_ "msa/common/logger"
	"msa/common/util"
	"msa/request/login"
	"msa/request/save"
	"strings"
	"sync"
	"time"
)

func main() {
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
			go task(date+passTime, i, wg)
		}
	}
	wg.Wait()
}

func task(passTime string, index int, wg *sync.WaitGroup) {
	log := logrus.WithField("point", passTime).WithField("index", index)
	defer func() {
		err := recover()
		if err != nil {
			logrus.Errorf("异常终止：%s", err)
		}
		log.Info("任务结束")
		wg.Done()
	}()
	JSESSIONID := login.Dail()
	log.Info("登录成功")
	save.Dail(passTime, JSESSIONID, log)
}
