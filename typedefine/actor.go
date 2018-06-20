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
	Bag *ActorBaseBagData //背包
}

type ActorBaseBagData struct {
	Currency    map[int]int         //货币id:数量
	AccCurrency map[int]int         //历史累积货币 id:数量
	Items       map[int]map[int]int //道具 类型:id:数量
}

type ActorExData struct {
	Pf     string //平台
	NewDay int64  //上次newday时间
}

type ActorDynamicData struct {
	Attr *ActorDynamicAttrData
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
	if actor.Account != nil {
		actor.Account.Reply(data)
	}
}
