package lordlevel

import (
	"github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"
	c "github.com/sencydai/gameworld/constdefine"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/log"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/service/attr"
	"github.com/sencydai/gameworld/service/bag"
	t "github.com/sencydai/gameworld/typedefine"
)

func init() {
	bag.RegObtainLordExp(onObtainLordExp)
}

func onObtainLordExp(actor *t.Actor, count int) {
	exData := actor.GetExData()
	exData.Exp += count

	curLevel := actor.Level

	recursionUpgrade(actor)

	actor.ReplyWriter(pack.AllocPack(proto.Lord, proto.LordSUpgrade, actor.Level, exData.Exp))
	if curLevel != actor.Level {
		log.Infof("actor(%d) upgrade: %d %d", actor.ActorId, curLevel, actor.Level)
		attr.RefreshLordLevelAttr(actor, true)
		rankData := t.GetRank(c.RankLevel)
		rankData.Insert(actor.ActorId, int64(actor.Level))
		service.OnActorUpgrade(actor, curLevel)
	}
}

func recursionUpgrade(actor *t.Actor) {
	if actor.Level >= len(g.GLordLevelConfig) {
		return
	}
	levelConf := g.GLordLevelConfig[actor.Level]
	exData := actor.GetExData()
	if exData.Exp < levelConf.NeedExp {
		return
	}
	actor.Level++
	exData.Exp -= levelConf.NeedExp

	recursionUpgrade(actor)
}
