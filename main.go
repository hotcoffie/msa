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

var finishedMsgs = map[string]bool{
	"操作成功": true,
	"进出口申报同一条船舶一天只能申请一次！": false,
	"当前船舶已经申报成功,请勿重新提交！":  false,
	"当前时间点已被申报，是否选择排队？":   false,
	"当前时间点已被申报,是否选择排队？":   false,
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
	finished := false
	start := false
	for !finished {
		if start || conf.Data.Active == conf.ActiveDev {
			result, err := save.Dail(passTime, JSESSIONID)
			if err != nil {
				log.WithError(err).Error("提交失败")
				continue
			}
			log.WithField("result", result.ResultDesc).Info("提交成功")
			_, finished = finishedMsgs[result.ResultDesc]
		} else {
			result, err := save.GetTime(JSESSIONID)
			if err != nil {
				log.WithError(err).Error("获取申报开始时间失败")
				continue
			}
			remainLog := log.WithField("remain", result)
			if result < 0 {
				remainLog.Info("抢点结束")
				break
			} else if result > 30 {
				remainLog.Debug("休息20秒")
				time.Sleep(20 * time.Second)
			} else {
				remainLog.Info("抢点开始")
				start = true
			}
		}
	}
}
