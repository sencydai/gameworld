package heroequip

import (
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
	"github.com/sencydai/gameworld/service"
	t "github.com/sencydai/gameworld/typedefine"
)

func init() {
	service.RegActorLogin(onActorLogin)
}

func onActorLogin(actor *t.Actor, offSec int) {
	equipData := actor.GetEquipDatas()
	writer := pack.AllocPack(proto.Hero, proto.HeroSEquipInit, int16(len(equipData)))
	for heroPos, equips := range equipData {
		pack.Write(writer, heroPos, int16(len(equips)))
		for pos, guid := range equips {
			pack.Write(writer, pos, guid)
		}
	}

	actor.ReplyWriter(writer)
}
