package heroartifact

import (
	"github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"
	"github.com/sencydai/gameworld/service"
	t "github.com/sencydai/gameworld/typedefine"
)

func init() {
	service.RegActorLogin(onActorLogin)
}

func onActorLogin(actor *t.Actor, offSec int) {
	onSendBaseInfo(actor)
}

func onSendBaseInfo(actor *t.Actor) {
	artis := actor.GetArtifactDatas()
	writer := pack.AllocPack(proto.Hero, proto.HeroSArtifactInit, int16(len(artis)))
	for pos, guid := range artis {
		pack.Write(writer, pos, guid)
	}
	actor.ReplyWriter(writer)
}
