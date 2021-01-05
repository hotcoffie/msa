package request

import (
	"github.com/sirupsen/logrus"
	"msa/request/login"
	"msa/request/save"
	"sync"
)

func Dail(passTime string, index int, wg *sync.WaitGroup) {
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
