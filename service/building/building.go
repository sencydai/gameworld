package building

import (
	"bytes"
	"fmt"

	"github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"
	c "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/dispatch"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/service/bag"
	t "github.com/sencydai/gameworld/typedefine"
)

const (
	outputMinute = 1
	outputOnce   = 2

	levelTypeMainCity = 1
	levelTypeLord     = 2
)

var (
	minOutputBuildings map[int]bool
)

func init() {
	service.RegConfigLoadFinish(onConfigLoadFinish)
	service.RegActorLogin(onActorLogin)
	service.RegActorMinTimer(onActorMinTimer)
	service.RegActorUpgrade(onActorUpgrade)

	dispatch.RegActorMsgHandle(proto.Building, proto.BuildingCUpgrade, onUpgrade)
}

func onConfigLoadFinish() {
	minOutputBuildings = make(map[int]bool)
	for _, conf := range g.GBuildingConfig {
		if conf.OutputType == outputMinute {
			minOutputBuildings[conf.Id] = true
		}
	}
}

func onActorLogin(actor *t.Actor, offSec int) {
	buildings := actor.GetBuildings()
	writer := pack.AllocPack(proto.Building, proto.BuildingSInit, int16(len(buildings)))
	for id, level := range buildings {
		pack.Write(writer, id, level)
	}
	actor.ReplyWriter(writer)
}

func onActorMinTimer(actor *t.Actor, times int) {
	buildings := actor.GetBuildings()
	for id := range minOutputBuildings {
		if level, ok := buildings[id]; ok {
			if times > 1 {
				levelConf := g.GBuildingLevelConfig[id][level]
				bag.PutItems2BagRatio(actor, levelConf.Output, float64(times), c.ASHookOutput, fmt.Sprintf("building_%d_%d", id, level))
			} else {
				calcOutput(actor, id, level)
			}
		}
	}
}

func onActorUpgrade(actor *t.Actor, oldLevel int) {
	buildings := actor.GetBuildings()
	if len(buildings) >= len(g.GBuildingConfig) {
		return
	}

	for _, conf := range g.GBuildingConfig {
		if _, ok := buildings[conf.Id]; ok || conf.Level > actor.Level {
			continue
		}
		buildings[conf.Id] = 1
		if conf.OutputType == outputOnce {
			calcOutput(actor, conf.Id, 1)
		}
		if oldLevel > 0 {
			onSendUpdate(actor, conf.Id, 1)
		}
	}
}

func onSendUpdate(actor *t.Actor, id, level int) {
	actor.Reply(proto.Building, proto.BuildingSUpdate, id, level)
}

func calcOutput(actor *t.Actor, id, level int) {
	levelConf := g.GBuildingLevelConfig[id][level]
	bag.PutItems2Bag(actor, levelConf.Output, c.ASHookOutput, fmt.Sprintf("building_%d_%d", id, level))
}

func onUpgrade(actor *t.Actor, reader *bytes.Reader) {
	var id int
	pack.Read(reader, &id)

	buildings := actor.GetBuildings()
	level, ok := buildings[id]
	if !ok {
		return
	}
	levelConfs := g.GBuildingLevelConfig[id]
	if level >= len(levelConfs) {
		return
	}

	levelConf := levelConfs[level]
	switch levelConf.LevelType {
	case levelTypeMainCity:
		if actor.GetBuildingLevel(c.BuildingMainCity) < levelConf.NeedLevel {
			return
		}
	case levelTypeLord:
		if actor.Level < levelConf.NeedLevel {
			return
		}
	}

	if !bag.DeductItems(actor, levelConf.Consume, true, "upgradeBuilding") {
		return
	}

	level++
	buildings[id] = level

	onSendUpdate(actor, id, level)

	buildConf := g.GBuildingConfig[id]
	if buildConf.OutputType == outputOnce {
		calcOutput(actor, id, level)
	}
}
