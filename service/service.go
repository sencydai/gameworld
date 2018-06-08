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

type GmHandle func(url.Values) (int, string)

type ActorCreateHandle func(*Actor)
type ActorBeforeLoginHandle func(actor *Actor, offTime time.Duration)
type ActorLoginHandle func(actor *Actor, offTime time.Duration)
type ActorLogoutHandle func(*Actor)
type ActorNewDayHandle func(*Actor)

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

func OnGameStart() {
	for _, handle := range gameStartHandles {
		handle()
	}
}

func RegGameClose(handle GameCloseHandle) {
	gameCloseHandles = append(gameCloseHandles, handle)
}

func OnGameClose() {
	for _, handle := range gameCloseHandles {
		handle()
	}
}

func RegConfigLoadFinish(handle ConfigLoadFinishHandle) {
	configLoadHandles = append(configLoadHandles, handle)
}

func OnConfigLoadFinish() {
	for _, handle := range configLoadHandles {
		handle()
	}
}

func RegSystemNewDay(handle SystemNewDayHandle) {
	systemNewDayHandles = append(systemNewDayHandles, handle)
}

func OnSystemNewDay() {
	for _, handle := range systemNewDayHandles {
		handle()
	}
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

func OnActorCreate(actor *Actor) {
	for _, handle := range actorCreateHandles {
		handle(actor)
	}
}

func RegActorBeforeLogin(handle ActorBeforeLoginHandle) {
	actorBeforeLoginHandles = append(actorBeforeLoginHandles, handle)
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

func RegActorLogin(handle ActorLoginHandle) {
	actorLoginHandles = append(actorLoginHandles, handle)
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

func RegActorLogout(handle ActorLogoutHandle) {
	actorLogoutHandles = append(actorLogoutHandles, handle)
}

func OnActorLogout(actor *Actor) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("actor(%d) logout error: %v,%s", actor.ActorId, err, string(debug.Stack()))
		}
	}()

	for _, handle := range actorLogoutHandles {
		handle(actor)
	}
}

func RegactorNewDay(handle ActorNewDayHandle) {
	actorNewDayHandles = append(actorNewDayHandles, handle)
}

func OnActorNewDay(actor *Actor) (ok bool) {
	ok = true

	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("actor(%d) newday error: %v,%s", actor.ActorId, err, string(debug.Stack()))
		}
		exData := actor.GetExData()
		exData.NewDay = time.Now().Unix()
	}()

	for _, handle := range actorNewDayHandles {
		handle(actor)
	}

	return
}
