package mainfuben

import (
	"bytes"

	_ "github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"
	c "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/dispatch"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/service/bag"
	"github.com/sencydai/gameworld/service/fight"
	t "github.com/sencydai/gameworld/typedefine"
)

func init() {
	dispatch.RegActorMsgHandle(proto.Fuben, proto.FubenCLoginMainFuben, onLoginMainFuben)
	service.RegActorLogin(onActorLogin)
	fight.RegFightClear(fight.MainFuben, onFightClear)
	fight.RegFightAward(fight.MainFuben, onGetFightAwards)
}

func onActorLogin(actor *t.Actor, offSec int) {
	onSendBaseInfo(actor)
}

func onSendBaseInfo(actor *t.Actor) {
	fubenId := actor.GetMainFuben() + 1
	if fubenId > len(g.GMainFubenConfig) {
		fubenId = -1
	}
	actor.Reply(proto.Fuben, proto.FubenSMainFuben, fubenId)
}

func onLoginMainFuben(actor *t.Actor, reader *bytes.Reader) {
	fubenId := actor.GetMainFuben() + 1
	fubenConf, ok := g.GMainFubenConfig[fubenId]
	if !ok {
		return
	}

	fight.NewPvE(actor, fight.MainFuben,
		fubenConf.Lord, fubenConf.Monster, "", 0,
		[]interface{}{fubenId})
}

func onFightClear(actor *t.Actor, fightData *t.FightData) {
	if fightData.RealResult != fight.Win {
		return
	}

	fubenConf := g.GMainFubenConfig[actor.GetMainFuben()+1]
	fightData.Awards = bag.GetRandomAwards(actor, c.RTProbability, fubenConf.Awards, true)
}

func onGetFightAwards(actor *t.Actor, fightData *t.FightData) {
	if fightData.RealResult != fight.Win {
		return
	}
	exData := actor.GetExData()
	exData.MainFB++
	onSendBaseInfo(actor)
}
