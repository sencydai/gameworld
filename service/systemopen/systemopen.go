package systemopen

import (
	"fmt"
	"time"

	"github.com/sencydai/gameworld/data"

	c "github.com/sencydai/gameworld/constdefine"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/timer"
	t "github.com/sencydai/gameworld/typedefine"
)

type systemOpenData struct {
	start int
	end   int
}

const (
	openStatusClosed = 0
	openStatusOpen   = 1
	openStatusComing = 2
)

var (
	systemOpens map[int]map[int]*systemOpenData
	openIds     map[int]int

	openPackData []byte
)

func init() {
	service.RegConfigLoadFinish(onConfigLoadFinish)
	service.RegActorLogin(onActorLogin)
}

func IsOpen(actor *t.Actor, id int) bool {
	if _, ok := systemOpens[openStatusOpen][id]; !ok {
		return false
	}

	conf := g.GSystemOpenConfig[id]
	if conf.Level > actor.Level {
		return false
	}

	if conf.Fight > actor.Power {
		return false
	}

	return true
}

func onConfigLoadFinish(isGameStart bool) {
	openIds = make(map[int]int)
	systemOpens = make(map[int]map[int]*systemOpenData)
	systemOpens[openStatusOpen] = make(map[int]*systemOpenData)
	systemOpens[openStatusComing] = make(map[int]*systemOpenData)

	for _, conf := range g.GSystemOpenConfig {
		conf.Time = g.ParseTime(conf.TimeType, conf.TimeSubType, conf.Time)
		conf.HTime = g.ParseTime(conf.HTimeType, conf.HTimeSubType, conf.HTime)

		checkOpen(conf.Type, openStatusOpen, false)
		checkOpen(conf.Type, openStatusComing, false)
	}
}

func onSystemTimeChange() {
	for _, conf := range g.GSystemOpenConfig {
		checkOpen(conf.Type, openStatusOpen, true)
		checkOpen(conf.Type, openStatusComing, true)
	}
}

func checkOpen(id int, openStatus int, update bool) {
	delete(systemOpens[openStatus], id)

	timerCheck := fmt.Sprintf("systemopen_%d_%d", openStatus, id)
	timer.StopTimer(nil, timerCheck)

	var (
		timeType    int
		timeSubtype int
		timeData    interface{}
	)

	conf, ok := g.GSystemOpenConfig[id]
	if !ok {
		updateOpenIds(id)
		if update {
			onUpdate(id)
		}
		return
	}

	switch openStatus {
	case openStatusOpen:
		timeType, timeSubtype, timeData = conf.TimeType, conf.TimeSubType, conf.Time
	case openStatusComing:
		timeType, timeSubtype, timeData = conf.HTimeType, conf.HTimeSubType, conf.HTime
	}

	now := time.Now()
	status, start, end := g.CheckTimeStatus(now, timeType, timeSubtype, timeData)
	switch status {
	case c.TimeStatusExpire:
	case c.TimeStatusUnlimit:
		systemOpens[openStatus][id] = &systemOpenData{start: -1, end: -1}
	case c.TimeStatusOutside:
		timer.After(nil, timerCheck, time.Second*time.Duration(start), checkOpen, id, openStatus, true)
	case c.TimeStatusInRange:
		systemOpens[openStatus][id] = &systemOpenData{start: start, end: int(now.Unix()) + end}
		timer.After(nil, timerCheck, time.Second*time.Duration(end), checkOpen, id, openStatus, true)
	}

	updateOpenIds(id)
	if update {
		onUpdate(id)
	}
}

func updateOpenIds(id int) {
	if _, ok := systemOpens[openStatusOpen][id]; ok {
		openIds[id] = openStatusOpen
	} else if _, ok = systemOpens[openStatusComing][id]; ok {
		openIds[id] = openStatusComing
	} else {
		delete(openIds, id)
	}

	writer := pack.AllocPack(proto.Base, proto.BaseSOpenSystemList, int16(len(openIds)))
	for id, openStatus := range openIds {
		data := systemOpens[openStatus][id]
		pack.Write(writer, id, openStatus, data.start, data.end)
	}

	openPackData = pack.EncodeWriter(writer)
}

func onActorLogin(actor *t.Actor, offSec int) {
	actor.ReplyData(openPackData)
}

func onUpdate(id int) {
	var (
		status int
		ok     bool
		start  = -1
		end    = -1
	)
	if status, ok = openIds[id]; ok {
		sysData := systemOpens[status][id]
		start, end = sysData.start, sysData.end
	}

	data.Broadcast(proto.Base, proto.BaseSOpenSystemSync, id, status, start, end)
}
