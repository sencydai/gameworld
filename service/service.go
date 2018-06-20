package service

import (
	"net/url"
	"runtime/debug"
	"time"

	. "github.com/sencydai/gameworld/typedefine"
	"github.com/sencydai/utils/log"
)

type GameStartHandle func()
type GameCloseHandle func()
type ConfigLoadFinishHandle func()
type SystemNewDayHandle func()

type GmHandle func(values url.Values) (int, string)

type ActorCreateHandle func(actor *Actor)
type ActorBeforeLoginHandle func(actor *Actor, offTime time.Duration)
type ActorLoginHandle func(actor *Actor, offTime time.Duration)
type ActorLogoutHandle func(actor *Actor)
type ActorNewDayHandle func(actor *Actor)

var (
	gameStartHandles    = make([]GameStartHandle, 0)
	gameCloseHandles    = make([]GameCloseHandle, 0)
	configLoadHandles   = make([]ConfigLoadFinishHandle, 0)
	systemNewDayHandles = make([]SystemNewDayHandle, 0)

	gmHandles = make(map[string]GmHandle)

	actorCreateHandles      = make([]ActorCreateHandle, 0)
	actorBeforeLoginHandles = make([]ActorBeforeLoginHandle, 0)
	actorLoginHandles       = make([]ActorLoginHandle, 0)
	actorLogoutHandles      = make([]ActorLogoutHandle, 0)
	actorNewDayHandles      = make([]ActorNewDayHandle, 0)
)

func RegGameStart(handle GameStartHandle) {
	gameStartHandles = append(gameStartHandles, handle)
}

func RegGameClose(handle GameCloseHandle) {
	gameCloseHandles = append(gameCloseHandles, handle)
}

func RegConfigLoadFinish(handle ConfigLoadFinishHandle) {
	configLoadHandles = append(configLoadHandles, handle)
}

func RegSystemNewDay(handle SystemNewDayHandle) {
	systemNewDayHandles = append(systemNewDayHandles, handle)
}

func RegGm(cmd string, handle GmHandle) {
	gmHandles[cmd] = handle
}

func GetGmHandle(cmd string) GmHandle {
	return gmHandles[cmd]
}

func RegActorCreate(handle ActorCreateHandle) {
	actorCreateHandles = append(actorCreateHandles, handle)
}

func RegActorBeforeLogin(handle ActorBeforeLoginHandle) {
	actorBeforeLoginHandles = append(actorBeforeLoginHandles, handle)
}

func RegActorLogin(handle ActorLoginHandle) {
	actorLoginHandles = append(actorLoginHandles, handle)
}

func RegActorLogout(handle ActorLogoutHandle) {
	actorLogoutHandles = append(actorLogoutHandles, handle)
}

func RegActorNewDay(handle ActorNewDayHandle) {
	actorNewDayHandles = append(actorNewDayHandles, handle)
}

func OnGameStart() {
	for _, handle := range gameStartHandles {
		handle()
	}
}

func OnGameClose() {
	for _, handle := range gameCloseHandles {
		handle()
	}
}

func OnConfigReloadFinish() {
	for _, handle := range configLoadHandles {
		handle()
	}
}

func OnSystemNewDay() {
	for _, handle := range systemNewDayHandles {
		handle()
	}
}

func OnActorCreate(actor *Actor) {
	for _, handle := range actorCreateHandles {
		handle(actor)
	}
}

func OnActorBeforeLogin(actor *Actor, offTime time.Duration) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("actor(%d) before login error: %v,%s", actor.ActorId, err, string(debug.Stack()))
		}
	}()
	for _, handle := range actorBeforeLoginHandles {
		handle(actor, offTime)
	}
}

func OnActorLogin(actor *Actor, offTime time.Duration) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("actor(%d) login error: %v,%s", actor.ActorId, err, string(debug.Stack()))
		}
	}()
	for _, handle := range actorLoginHandles {
		handle(actor, offTime)
	}
}

func OnActorLogout(actor *Actor) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("actor(%d) logout error: %v,%s", actor.ActorId, err, string(debug.Stack()))
		}
	}()
	actor.Account = nil
	for _, handle := range actorLogoutHandles {
		handle(actor)
	}
}

func OnActorNewDay(actor *Actor) (ok bool) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("actor(%d) newday error: %v,%s", actor.ActorId, err, string(debug.Stack()))
		}
		exData := actor.GetExData()
		exData.NewDay = time.Now().Unix()
	}()

	ok = true

	for _, handle := range actorNewDayHandles {
		handle(actor)
	}

	return
}
