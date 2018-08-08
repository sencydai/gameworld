package chat

import (
	"fmt"
	"time"

	"github.com/sencydai/gameworld/base"
	c "github.com/sencydai/gameworld/constdefine"
	g "github.com/sencydai/gameworld/gconfig"
	//"github.com/sencydai/gameworld/log"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/timer"
)

func init() {
	service.RegConfigLoadFinish(onConfigLoadFinish)
	service.RegSystemTimeChange(onSystemTimeChange)
}

func onConfigLoadFinish(isGameStart bool) {
	for _, v := range g.GSystemNoticeConfig {
		for _, vv := range v {
			for _, vvv := range vv {
				vvv.SpecId = base.ReverseKeyValue(vvv.SpecId)
			}
		}
	}

	for _, conf := range g.GSystemTimingNoticeConfig {
		conf.Time = g.ParseTime(conf.TimeType, conf.TimeSubType, conf.Time)
		checkOpen(conf.Id)
	}
}

func onSystemTimeChange() {
	for _, conf := range g.GSystemTimingNoticeConfig {
		checkOpen(conf.Id)
	}
}

func checkOpen(id int) {
	timerNotice := fmt.Sprintf("timenotice_%d", id)
	timerCheck := fmt.Sprintf("timenotice_check_%d", id)
	timer.StopTimer(nil, timerNotice)
	timer.StopTimer(nil, timerCheck)

	conf, ok := g.GSystemTimingNoticeConfig[id]
	if !ok {
		return
	}

	now := time.Now()
	status, start, end := g.CheckTimeStatus(now, conf.TimeType, conf.TimeSubType, conf.Time)
	//log.Infof("checkOpen begin %d %d,%d,%d", id, status, start, end)
	switch status {
	case c.TimeStatusUnlimit:
		timer.Loop(nil, timerNotice, 0, time.Second*time.Duration(conf.Interval), -1, sendTimingNotice, id)
	case c.TimeStatusOutside:
		timer.After(nil, timerCheck, time.Second*time.Duration(start), checkOpen, id)
	case c.TimeStatusInRange:
		timer.Loop(nil, timerNotice, 0, time.Second*time.Duration(conf.Interval), -1, sendTimingNotice, id)
		timer.After(nil, timerCheck, time.Second*time.Duration(end+1), checkOpen, id)
	}
}

func sendTimingNotice(id int) {
	conf, ok := g.GSystemTimingNoticeConfig[id]
	if !ok {
		timerNotice := fmt.Sprintf("timenotice_%d", id)
		timerCheck := fmt.Sprintf("timenotice_check_%d", id)
		timer.StopTimer(nil, timerNotice)
		timer.StopTimer(nil, timerCheck)
		return
	}

	BroadcastMsg(byte(conf.NoticeType), conf.Channel, byte(conf.Roll), conf.Content)
}
