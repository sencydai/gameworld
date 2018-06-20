package data

import (
	"fmt"
	"time"

	"github.com/json-iterator/go"

	"github.com/sencydai/gameworld/base"
	"github.com/sencydai/gameworld/engine"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/timer"
	. "github.com/sencydai/gameworld/typedefine"
	"github.com/sencydai/utils/log"
)

var (
	json         = jsoniter.ConfigCompatibleWithStandardLibrary
	sysDataIndex = []int{SYSTEM_ACTOR_INDEX, SYSTEM_RANK_INDEX, SYSTEM_GUILD_INDEX, SYSTEM_COMMON_INDEX}

	sysData = &SystemData{
		Actors:      make(map[int64]*SystemStaticActorData),
		Rank:        &SystemStaticRankData{},
		Guild:       &SystemStaticGuildData{},
		Data:        &SystemStaticCommonData{},
		OpenServer:  time.Now(),
		DynamicData: &SystemDynamicData{},
	}
)

func init() {
	service.RegGameStart(onGameStart)
}

func loadData(index int, data interface{}, defValue string) {
	if text, err := engine.GetSystemData(index, defValue); err != nil {
		panic(err)
	} else if err = json.Unmarshal([]byte(text), data); err != nil {
		panic(err)
	}
}

func OnLoadSystemData() {
	loadData(SYSTEM_ACTOR_INDEX, &sysData.Actors, "{}")
	loadData(SYSTEM_RANK_INDEX, sysData.Rank, "{}")
	loadData(SYSTEM_GUILD_INDEX, sysData.Guild, "{}")
	loadData(SYSTEM_COMMON_INDEX, sysData.Data, "{}")

	if text, err := engine.GetSystemData(SYSTEM_OPENSERVER_INDEX, base.FormatDateTime(time.Now())); err != nil {
		panic(err)
	} else if sysData.OpenServer, err = base.ParseDateTime(text); err != nil {
		panic(err)
	}
}

func marshalSystemData(index int) ([]byte, error) {
	switch index {
	case SYSTEM_ACTOR_INDEX:
		return json.Marshal(sysData.Actors)
	case SYSTEM_RANK_INDEX:
		return json.Marshal(sysData.Rank)
	case SYSTEM_GUILD_INDEX:
		return json.Marshal(sysData.Guild)
	case SYSTEM_COMMON_INDEX:
		return json.Marshal(sysData.Data)
	default:
		return nil, fmt.Errorf("error system data index")
	}
}

func saveSystemData() {
	for _, index := range sysDataIndex {
		if data, err := marshalSystemData(index); err != nil {
			log.Errorf("json marshal system data %d error: %s", index, err.Error())
		} else {
			engine.UpdateSystemData(index, string(data), nil)
		}
	}
}

func onGameStart() {
	timer.Loop(nil, "savesystemdata", time.Hour, time.Hour, -1, saveSystemData)
	timer.Loop(nil, "clearactorcache", time.Hour*12, time.Hour*12, -1, clearTimeoutActorCache)
	data := GetSysData()
	timer.LoopDayMoment("newday", base.Unix(data.NewDay), 0, 0, 10, onNewDay)
}

func onNewDay() {
	log.Info("==============newday============")
	service.OnSystemNewDay()
	data := GetSysData()
	data.NewDay = time.Now().Unix()

	LoopActors(service.OnActorNewDay)
}

func OnGameClose() {
	chs := make(map[chan bool]bool)
	for _, index := range sysDataIndex {
		if data, err := marshalSystemData(index); err != nil {
			log.Errorf("json marshal system data %d error: %s", index, err.Error())
		} else {
			ch := make(chan bool, 1)
			chs[ch] = true
			engine.UpdateSystemData(index, string(data), ch)
		}
	}

	for ch := range chs {
		<-ch
	}

	log.Info("save system data success")
}

func GetSysActorData(actorId int64) *SystemStaticActorData {
	data, ok := sysData.Actors[actorId]
	if !ok {
		data = &SystemStaticActorData{}
		sysData.Actors[actorId] = data
	}
	return data
}

func GetSysRankData() *SystemStaticRankData {
	return sysData.Rank
}

func GetSysGuildData() *SystemStaticGuildData {
	return sysData.Guild
}

func GetSysData() *SystemStaticCommonData {
	return sysData.Data
}

func GetSysOpenServerTime() time.Time {
	return sysData.OpenServer
}

func GetSysDynamicData() *SystemDynamicData {
	return sysData.DynamicData
}
