package lorddecor

import (
	"bytes"

	c "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/dispatch"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/service/bag"
	t "github.com/sencydai/gameworld/typedefine"

	"github.com/sencydai/gameworld/log"
)

func init() {
	service.RegActorCreate(onActorCreate)
	service.RegActorLogin(onActorLogin)
	bag.RegObtainLordDecor(onObtainDecor)

	dispatch.RegActorMsgHandle(proto.Lord, proto.LordCDecorChange, onChangeDecor)
}

func onActorCreate(actor *t.Actor) {
	decorData := actor.GetDecorData()
	conf := g.GLordConfig[actor.Camp][actor.Sex]

	//头像
	decor := &t.ActorBaseDecorData{Id: conf.Head, Unlock: make(map[int]byte)}
	decor.Unlock[decor.Id] = 1
	decorData[c.LDTHead] = decor

	//边框
	decor = &t.ActorBaseDecorData{Id: conf.Frame, Unlock: make(map[int]byte)}
	decor.Unlock[decor.Id] = 1
	decorData[c.LDTFrame] = decor

	//聊天框
	decor = &t.ActorBaseDecorData{Id: conf.Chat, Unlock: make(map[int]byte)}
	decor.Unlock[decor.Id] = 1
	decorData[c.LDTChat] = decor
}

func onObtainDecor(actor *t.Actor, ldt c.LordDecorType, id int) {
	switch ldt {
	case c.LDTHead:
		if _, ok := g.GLordHeadConfig[id]; !ok {
			log.Errorf("not find lord head %d", id)
			return
		}
	case c.LDTFrame:
		if _, ok := g.GLordFrameConfig[id]; !ok {
			log.Errorf("not find lord frame %d", id)
			return
		}
	case c.LDTChat:
		if _, ok := g.GLordChatConfig[id]; !ok {
			log.Errorf("not find lord chat %d", id)
			return
		}
	}

	decorData := actor.GetDecorData()
	decor := decorData[ldt]
	//已解锁
	if _, ok := decor.Unlock[id]; ok {
		return
	}
	decor.Unlock[id] = 1

	writer := pack.AllocPack(proto.Lord, proto.LordSDecorUnlock, ldt, id)
	actor.ReplyWriter(writer)
}

func onActorLogin(actor *t.Actor, offSec int) {
	decorData := actor.GetDecorData()
	writer := pack.AllocPack(
		proto.Lord,
		proto.LordSDecorInit,
		int16(len(decorData)),
	)
	for t, decor := range decorData {
		pack.Write(writer, t, decor.Id, int16(len(decor.Unlock)))
		for id := range decor.Unlock {
			pack.Write(writer, id)
		}
	}
	actor.ReplyWriter(writer)
}

//更换装饰
func onChangeDecor(actor *t.Actor, reader *bytes.Reader) {
	var (
		dt int
		id int
	)
	pack.Read(reader, &dt, &id)

	decorData := actor.GetDecorData()
	decor, ok := decorData[dt]
	if !ok || decor.Id == id {
		return
	}
	//装饰未解锁
	if _, ok := decor.Unlock[id]; !ok {
		return
	}
	decor.Id = id
	writer := pack.AllocPack(proto.Lord, proto.LordSDecorChange, dt, id)
	actor.ReplyWriter(writer)
}
