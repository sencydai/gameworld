package data

import (
	"time"

	"github.com/sencydai/gameworld/engine"
	. "github.com/sencydai/gameworld/typedefine"
	"github.com/sencydai/utils/log"
)

var (
	onlineActors = make(map[int64]*Actor)
	actorCaches  = make(map[int64]*ActorCache)
	cacheTimeout = time.Hour * 12
)

func AddOnlineActor(actor *Actor) {
	actor.LoginTime = time.Now()
	onlineActors[actor.ActorId] = actor
	log.Infof("actor(%d,%d,%s) login", actor.AccountId, actor.ActorId, actor.ActorName)
	delete(actorCaches, actor.ActorId)
}

func RemoveOnlineActor(actor *Actor, flush chan bool) {
	if actor, ok := onlineActors[actor.ActorId]; ok {
		actor.LogoutTime = time.Now()
		delete(onlineActors, actor.ActorId)
		engine.UpdateActor(actor, flush)
		log.Infof("actor(%d,%d,%s) logout", actor.AccountId, actor.ActorId, actor.ActorName)
		actor.DynamicData = nil
		actor.ExData = nil
		actor.Account = nil
		actorCaches[actor.ActorId] = &ActorCache{Actor: actor, Refresh: time.Now()}
	}
}

type LoopActorsHandle func(actor *Actor) bool

func LoopActors(handle LoopActorsHandle) {
	for _, actor := range onlineActors {
		if ok := handle(actor); !ok {
			break
		}
	}
}

func GetOnlineActor(actorId int64) *Actor {
	return onlineActors[actorId]
}

func GetActor(actorId int64) *Actor {
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
	actorCaches[actorId] = &ActorCache{Actor: actor, Refresh: time.Now()}
	return actor
}

func clearTimeoutActorCache() {
	now := time.Now()
	for actorId, cache := range actorCaches {
		if now.Sub(cache.Refresh) > cacheTimeout {
			delete(actorCaches, actorId)
		}
	}
}
