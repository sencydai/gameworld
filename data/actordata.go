package data

import (
	"bytes"
	"time"

	"github.com/sencydai/gameworld/engine"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/log"
	"github.com/sencydai/gameworld/proto/pack"
	"github.com/sencydai/gameworld/service"
	t "github.com/sencydai/gameworld/typedefine"
)

var (
	onlineActors = make(map[int64]*t.Actor)
	actorCaches  = make(map[int64]*t.ActorCache)
	cacheTimeout = time.Hour * 2
)

func init() {
	service.RegActorDataLoad(onActorDataLoad)
}

func onActorDataLoad(actor *t.Actor) {
	//领主模型
	lordConf := g.GLordConfig[actor.Camp][actor.Sex]
	dynamicData := actor.GetDynamicData()
	dynamicData.LordModel = lordConf.Model
}

func AddOnlineActor(actor *t.Actor) {
	actor.LoginTime = time.Now()
	onlineActors[actor.ActorId] = actor
	log.Infof("actor(%d,%d,%s) login", actor.AccountId, actor.ActorId, actor.ActorName)
	delete(actorCaches, actor.ActorId)

	service.OnActorDataLoad(actor)
}

func RemoveOnlineActor(actor *t.Actor) {
	if actor, ok := onlineActors[actor.ActorId]; ok {
		delete(onlineActors, actor.ActorId)
		actor.Account = nil
		actor.LogoutTime = time.Now()
		engine.UpdateActor(actor)
		log.Infof("actor(%d,%d,%s) logout", actor.AccountId, actor.ActorId, actor.ActorName)
		actor.DynamicData = nil
		actor.ExData = nil

		actorCaches[actor.ActorId] = &t.ActorCache{Actor: actor, Refresh: time.Now()}

		service.OnActorDataLoad(actor)
	}
}

func LoopActors(handle func(actor *t.Actor) bool) {
	for _, actor := range onlineActors {
		if ok := handle(actor); !ok {
			break
		}
	}
}

func Broadcast(sysId, cmdId byte, data ...interface{}) {
	BroadcastWriter(pack.AllocPack(sysId, cmdId, data...))
}

func BroadcastData(data []byte) {
	for _, actor := range onlineActors {
		actor.ReplyData(data)
	}
}

func BroadcastWriter(writer *bytes.Buffer) {
	data := pack.EncodeWriter(writer)
	for _, actor := range onlineActors {
		actor.ReplyData(data)
	}
}

func GetOnlineActor(actorId int64) *t.Actor {
	return onlineActors[actorId]
}

func GetActor(actorId int64) *t.Actor {
	actor := onlineActors[actorId]
	if actor != nil {
		return actor
	}
	actorCache := actorCaches[actorId]
	if actorCache != nil {
		actorCache.Refresh = time.Now()
		return actorCache.Actor
	}
	actor, err := engine.QueryActorCache(actorId)
	if err != nil {
		return nil
	}

	service.OnActorDataLoad(actor)

	actorCaches[actorId] = &t.ActorCache{Actor: actor, Refresh: time.Now()}
	return actor
}

func clearTimeoutActorCache() {
	now := time.Now()
	for actorId, cache := range actorCaches {
		if now.Sub(cache.Refresh) >= cacheTimeout {
			delete(actorCaches, actorId)
		}
	}
}

func GetOnlineCount() int {
	return len(onlineActors)
}

func GetCacheCount() int {
	return len(actorCaches)
}
