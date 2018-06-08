package timer

import (
	"sync"
	"time"

	"github.com/sencydai/gameworld/base"
	"github.com/sencydai/gameworld/dispatch"
	. "github.com/sencydai/gameworld/typedefine"
	"github.com/sencydai/utils/log"
)

var (
	sysTimers   = make(map[string]*time.Timer)
	actorTimers = make(map[int64]map[string]*time.Timer)
	mutex       sync.Mutex
)

func addTimer(actor *Actor, name string, delay time.Duration) *time.Timer {
	mutex.Lock()
	defer mutex.Unlock()

	if actor == nil {
		if _, ok := sysTimers[name]; ok {
			log.Errorf("timer %s is existed", name)
			return nil
		}

		t := time.NewTimer(delay)
		sysTimers[name] = t
		return t
	}

	actors, ok := actorTimers[actor.ActorId]
	if !ok {
		actors := make(map[string]*time.Timer)
		actorTimers[actor.ActorId] = actors
	} else if _, ok := actors[name]; ok {
		log.Errorf("timer %s is existed", name)
		return nil
	}

	t := time.NewTimer(delay)
	actors[name] = t
	return t
}

func StopTimer(actor *Actor, name string) {
	mutex.Lock()
	defer mutex.Unlock()

	if actor == nil {
		if timer, ok := sysTimers[name]; ok {
			timer.Stop()
			delete(sysTimers, name)
		}

		return
	}

	actors, ok := actorTimers[actor.ActorId]
	if !ok {
		return
	}
	if timer, ok := actors[name]; ok {
		timer.Stop()
		delete(actors, name)
	}
}

func StopActorTimers(actor *Actor) {
	mutex.Lock()
	defer mutex.Unlock()

	if actors, ok := actorTimers[actor.ActorId]; ok {
		for _, timer := range actors {
			timer.Stop()
		}
		delete(actorTimers, actor.ActorId)
	}
}

func After(actor *Actor, name string, delay time.Duration, cbFunc interface{}, args ...interface{}) {
	t := addTimer(actor, name, delay)
	if t == nil {
		return
	}

	go func() {
		select {
		case <-t.C:
			if actor == nil {
				dispatch.PushSystemMsg(cbFunc, args...)
			} else if actor.Account != nil {
				if len(args) == 0 {
					dispatch.PushSystemMsg(cbFunc, actor)
				} else {
					dispatch.PushSystemMsg(cbFunc, append([]interface{}{actor}, args...)...)
				}
			}
		}
		StopTimer(actor, name)
	}()
}

func Loop(actor *Actor, name string, delay, interval time.Duration, times int, cbFunc interface{}, args ...interface{}) {
	t := addTimer(actor, name, delay)
	if t == nil {
		return
	}

	go func() {
		var count int
	TAG_STOP_FOR:
		for {
			select {
			case <-t.C:
				if times > 0 {
					if count < times {
						count++
						t.Reset(interval)
					} else {
						break TAG_STOP_FOR
					}
				} else {
					t.Reset(interval)
				}

				go func() {
					if actor == nil {
						dispatch.PushSystemMsg(cbFunc, args...)
					} else if actor.Account != nil {
						if len(args) == 0 {
							dispatch.PushSystemMsg(cbFunc, actor)
						} else {
							dispatch.PushSystemMsg(cbFunc, append([]interface{}{actor}, args...)...)
						}
					}
				}()
			}
		}
		StopTimer(actor, name)
	}()
}

func LoopDayMoment(name string, last time.Time, hour, min, sec int, cbFunc interface{}, args ...interface{}) {
	if base.CheckMomentHappend(last, hour, min, sec) {
		dispatch.PushSystemMsg(cbFunc, args...)
	}

	Loop(nil, name, base.GetMomentDelay(hour, min, sec), time.Hour*24, -1, cbFunc, args...)
}
