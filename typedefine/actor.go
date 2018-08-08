package typedefine

import (
	"bytes"
	"time"

	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
)

type Actor struct {
	ActorId    int64
	ActorName  string
	AccountId  int
	Camp       int
	Sex        int
	Level      int
	Power      int
	CreateTime time.Time
	LoginTime  time.Time
	LogoutTime time.Time

	BaseData *ActorBaseData
	ExData   *ActorExData

	DynamicData *ActorDynamicData
	Account     *Account
}

type ActorCache struct {
	Actor   *Actor
	Refresh time.Time
}

func (actor *Actor) GetBaseData() *ActorBaseData {
	return actor.BaseData
}

func (actor *Actor) GetExData() *ActorExData {
	if actor.ExData == nil {
		actor.ExData = &ActorExData{}
	}

	return actor.ExData
}

func (actor *Actor) GetDynamicData() *ActorDynamicData {
	if actor.DynamicData == nil {
		actor.DynamicData = &ActorDynamicData{}
	}

	return actor.DynamicData
}

func (actor *Actor) Reply(sysId, cmdId byte, data ...interface{}) {
	actor.ReplyWriter(pack.AllocPack(sysId, cmdId, data...))
}

func (actor *Actor) ReplyData(data []byte) {
	if account := actor.Account; account != nil {
		account.Reply(data)
	}
}

func (actor *Actor) ReplyWriter(writer *bytes.Buffer) {
	if account := actor.Account; account != nil {
		account.Reply(pack.EncodeWriter(writer))
	}
}

func (actor *Actor) SendTips(msg string) {
	actor.Reply(proto.Chat, proto.ChatSTips, 0, msg)
}

func (actor *Actor) IsOnline() bool {
	return actor.Account != nil
}
