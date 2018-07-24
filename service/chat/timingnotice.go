package chat

import (
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service"
)

func init() {
	service.RegConfigLoadFinish(onConfigLoadFinish)
}

func onConfigLoadFinish() {
	for _, conf := range g.GSystemTimingNoticeConfig {
		conf.Time = g.ParseTime(conf.TimeType, conf.TimeSubType, conf.Time)
	}
}
