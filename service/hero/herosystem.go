package hero

import (
	"github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"
	c "github.com/sencydai/gameworld/constdefine"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/service/bag"
	t "github.com/sencydai/gameworld/typedefine"
)

func init() {
	service.RegActorCreate(onActorCreate)
	service.RegActorLogin(onActorLogin)
}

func onActorCreate(actor *t.Actor) {
	fightHeros := actor.GetFightHeros()
	conf := g.GLordConfig[actor.Camp][actor.Sex]
	for pos, id := range conf.Heros {
		hero := bag.NewHero(actor, id)
		hero.PosType = c.HPTFight
		hero.Pos = pos
		fightHeros[pos] = hero.Guid
	}
}

func onActorLogin(actor *t.Actor, offSec int) {
	onSendBaseInfo(actor)
}

func onSendBaseInfo(actor *t.Actor) {
	writer := pack.AllocPack(proto.Hero, proto.HeroSArmyInit)
	fightHeros := actor.GetFightHeros()
	pack.Write(writer, int16(len(fightHeros)))
	for pos, guid := range fightHeros {
		pack.Write(writer, pos, guid)
	}

	assistHeros := actor.GetAssistHeros()
	pack.Write(writer, int16(len(assistHeros)))
	for pos, guid := range assistHeros {
		pack.Write(writer, pos, guid)
	}
	actor.ReplyWriter(writer)
}
