package data

import (
	"fmt"
	"time"

	"github.com/json-iterator/go"

	"github.com/sencydai/gameworld/base"
	"github.com/sencydai/gameworld/engine"
	"github.com/sencydai/gameworld/log"
	proto "github.com/sencydai/gameworld/proto/protocol"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/timer"
	t "github.com/sencydai/gameworld/typedefine"
)

var (
	json         = jsoniter.ConfigCompatibleWithStandardLibrary
	sysDataIndex = []int{t.SYSTEM_ACTOR_INDEX, t.SYSTEM_RANK_INDEX, t.SYSTEM_GUILD_INDEX, t.SYSTEM_COMMON_INDEX}
)

func init() {
	service.RegGameStart(onGameStart)
	service.RegSystemTimeChange(onSystemTimeChange)
}

func loadData(index int, data interface{}, defValue string) {
	if text, err := engine.GetSystemData(index, defValue); err != nil {
		panic(err)
	} else if err = json.Unmarshal([]byte(text), data); err != nil {
		panic(err)
	}
}

func OnLoadSystemData() {
	sysData := t.GetSysData()
	loadData(t.SYSTEM_ACTOR_INDEX, &sysData.Actors, "{}")
	loadData(t.SYSTEM_RANK_INDEX, &sysData.Rank, "{}")
	loadData(t.SYSTEM_GUILD_INDEX, sysData.Guild, "{}")
	loadData(t.SYSTEM_COMMON_INDEX, sysData.Data, "{}")

	if text, err := engine.GetSystemData(t.SYSTEM_OPENSERVER_INDEX, base.FormatDateTime(time.Now())); err != nil {
		panic(err)
	} else if sysData.OpenServer, err = base.ParseDateTime(text); err != nil {
		panic(err)
	}

	for _, rankData := range sysData.Rank {
		rankData.OnDataLoaded()
	}
}

func marshalSystemData(index int) ([]byte, error) {
	sysData := t.GetSysData()
	switch index {
	case t.SYSTEM_ACTOR_INDEX:
		return json.MarshalIndent(sysData.Actors, "", " ")
	case t.SYSTEM_RANK_INDEX:
		return json.MarshalIndent(sysData.Rank, "", " ")
	case t.SYSTEM_GUILD_INDEX:
		return json.MarshalIndent(sysData.Guild, "", " ")
	case t.SYSTEM_COMMON_INDEX:
		return json.MarshalIndent(sysData.Data, "", " ")
	default:
		return nil, fmt.Errorf("error system data index")
	}
}

func PackSystemDatas() map[int]string {
	values := make(map[int]string)
	for _, index := range sysDataIndex {
		if data, err := marshalSystemData(index); err != nil {
			log.Errorf("json marshal system data %d error: %s", index, err.Error())
		} else {
			values[index] = string(data)
		}
	}

	return values
}

func onGameStart() {
	timer.Loop(nil, "savesystemdata", time.Hour, time.Hour, -1, func() {
		go func(values map[int]string) {
			engine.FlushSystemData(values)
		}(PackSystemDatas())
	})

	timer.Loop(nil, "clearactorcache", cacheTimeout, cacheTimeout, -1, clearTimeoutActorCache)
	data := t.GetSysCommonData()
	timer.LoopDayMoment("newday", base.Unix(data.NewDay), 0, 0, 0, onNewDay)

	service.RegGm("flush", func(map[string]string) (int, string) {
		saveData()
		return 0, "success"
	})

	service.RegGm("newday", func(map[string]string) (int, string) {
		onNewDay()
		return 0, "success"
	})

	_, min, sec := time.Now().Clock()
	min = min % 10
	timer.After(nil, "statOnlineActors", time.Second*time.Duration(600-min*60-sec), onStatOnlineActors)
}

func onStatOnlineActors() {
	log.Optf("online actors: %d", len(onlineActors))

	_, min, sec := time.Now().Clock()
	min = min % 10
	timer.After(nil, "statOnlineActors", time.Second*time.Duration(600-min*60-sec), onStatOnlineActors)
}

func onSystemTimeChange() {
	data := t.GetSysCommonData()
	timer.LoopDayMoment("newday", base.Unix(data.NewDay), 0, 0, 0, onNewDay)
}

func onNewDay() {
	log.Info("==============newday============")
	service.OnSystemNewDay()

	//newday
	Broadcast(proto.Base, proto.BaseSNewDay)

	now := time.Now()
	nowS := now.Unix()
	open := t.GetSysOpenServerTime()
	openS := open.Unix()
	deltaOpen := base.GetDeltaDays(now, open)

	data := t.GetSysCommonData()
	data.NewDay = nowS

	//同步服务器时间
	Broadcast(proto.Base, proto.BaseSSyncTime, int(nowS), openS, deltaOpen)

	LoopActors(func(actor *t.Actor) bool {
		service.OnActorNewDay(actor)
		return true
	})
}

func saveData() {
	tick := time.Now()
	engine.FlushSystemData(PackSystemDatas())
	log.Infof("save system data: %v", time.Since(tick))
}

func OnGameClose() {
	saveData()
}
