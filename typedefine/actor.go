package typedefine

import (
	"time"
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

type ActorBaseData struct {
}

type ActorExData struct {
	pf     string
	NewDay int64
}

type ActorDynamicData struct {
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

func (actor *Actor) Reply(data []byte) {
	if actor.Account == nil {
		return
	}

	actor.Account.Reply(data)
}
