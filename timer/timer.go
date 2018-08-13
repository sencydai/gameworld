package timer

import (
	"reflect"
	"runtime/debug"
	"sync"
	"time"

	"github.com/sencydai/gameworld/base"
	"github.com/sencydai/gameworld/dispatch"
	t "github.com/sencydai/gameworld/typedefine"

	"github.com/sencydai/gameworld/log"
)

var (
	sysTimers   = make(map[string]*time.Timer)
	actorTimers = make(map[*t.Actor]map[string]*time.Timer)
	mutex       sync.RWMutex
)

func init() {
	dispatch.TriggerSystemMsg = triggerSystemMsg
	dispatch.TriggerSystemMsgGo = triggerSystemMsgGo
	dispatch.TriggerActorMsg = triggerActorMsg
	dispatch.TriggerActorMsgGo = triggerActorMsgGo
}

func addTimer(actor *t.Actor, name string, delay time.Duration) *time.Timer {
	mutex.Lock()
	defer mutex.Unlock()

	if actor == nil {
		if t, ok := sysTimers[name]; ok {
			t.Stop()
		}
		t := time.NewTimer(delay)
		sysTimers[name] = t
		return t
	}

	actors, ok := actorTimers[actor]
	if !ok {
		actors = make(map[string]*time.Timer)
		actorTimers[actor] = actors
	} else if t, ok := actors[name]; ok {
		t.Stop()
	}

	t := time.NewTimer(delay)
	actors[name] = t
	return t
}

func StopTimer(actor *t.Actor, name string) bool {
	mutex.Lock()
	defer mutex.Unlock()

	if actor == nil {
		timer, ok := sysTimers[name]
		if ok {
			timer.Stop()
			delete(sysTimers, name)
		}
		return ok
	}
	actors, ok := actorTimers[actor]
	if !ok {
		return ok
	}
	timer, ok := actors[name]
	if ok {
		timer.Stop()
		delete(actors, name)
	}

	return ok
}

func IsStoped(actor *t.Actor, name string) bool {
	mutex.RLock()
	defer mutex.RUnlock()

	if actor == nil {
		_, ok := sysTimers[name]
		return !ok
	}
	actors, ok := actorTimers[actor]
	if !ok {
		return true
	}
	_, ok = actors[name]
	return !ok
}

func stopTimer(actor *t.Actor, name string, t *time.Timer) bool {
	mutex.Lock()
	defer mutex.Unlock()

	if actor == nil {
		timer, ok := sysTimers[name]
		if ok && timer == t {
			timer.Stop()
			delete(sysTimers, name)
			return true
		}
		return false
	}
	actors, ok := actorTimers[actor]
	if !ok {
		return ok
	}
	timer, ok := actors[name]
	if ok {
		timer.Stop()
		delete(actors, name)
	}

	return ok
}

func isStoped(actor *t.Actor, name string, t *time.Timer) bool {
	mutex.RLock()
	defer mutex.RUnlock()

	if actor == nil {
		timer, ok := sysTimers[name]
		return !ok || timer != t
	}
	actors, ok := actorTimers[actor]
	if !ok {
		return true
	}
	timer, ok := actors[name]
	return !ok || timer != t
}

func StopActorTimers(actor *t.Actor) {
	mutex.Lock()
	defer mutex.Unlock()

	if actors, ok := actorTimers[actor]; ok {
		for _, timer := range actors {
			timer.Stop()
		}
		delete(actorTimers, actor)
	}
}

func triggerSystemMsg(name string, t *time.Timer, stop bool, cb reflect.Value, values []reflect.Value) {
	if stop {
		if !stopTimer(nil, name, t) {
			return
		}
	} else if isStoped(nil, name, t) {
		return
	}

	cb.Call(values)
}

func triggerSystemMsgGo(name string, t *time.Timer, stop bool, cb reflect.Value, values []reflect.Value) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalf("timer %s ==> %v,%s", name, err, string(debug.Stack()))
			}
		}()

		triggerSystemMsg(name, t, stop, cb, values)
	}()
}

func triggerActorMsg(actor *t.Actor, name string, t *time.Timer, stop bool, cb reflect.Value, values []reflect.Value) {
	if stop {
		if !stopTimer(actor, name, t) {
			return
		}
	} else if isStoped(actor, name, t) {
		return
	}

	v := reflect.ValueOf(actor)
	if len(values) == 0 {
		values = []reflect.Value{v}
	} else {
		values = append([]reflect.Value{v}, values...)
	}

	cb.Call(values)
}

func triggerActorMsgGo(actor *t.Actor, name string, t *time.Timer, stop bool, cb reflect.Value, values []reflect.Value) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalf("timer %s ==> %v,%s", name, err, string(debug.Stack()))
			}
		}()

		triggerActorMsg(actor, name, t, stop, cb, values)
	}()

}

func After(actor *t.Actor, name string, delay time.Duration, cbFunc interface{}, args ...interface{}) {
	go func() {
		t := addTimer(actor, name, delay)
		select {
		case <-t.C:
			cb, values := base.ReflectFunc(cbFunc, args)
			if actor == nil {
				dispatch.PushSystemMsg(triggerSystemMsg, name, t, true, cb, values)
			} else {
				dispatch.PushActorMsg(actor, triggerActorMsg, name, t, true, cb, values)
			}
		}
	}()
}

func AfterGo(actor *t.Actor, name string, delay time.Duration, cbFunc interface{}, args ...interface{}) {
	go func() {
		t := addTimer(actor, name, delay)
		select {
		case <-t.C:
			cb, values := base.ReflectFunc(cbFunc, args)
			if actor == nil {
				dispatch.PushSystemMsg(triggerSystemMsgGo, name, t, true, cb, values)
			} else {
				dispatch.PushActorMsg(actor, triggerActorMsgGo, name, t, true, cb, values)
			}
		}
	}()
}

func Next(actor *t.Actor, name string, cbFunc interface{}, args ...interface{}) {
	After(actor, name, 0, cbFunc, args...)
}

func NextGo(actor *t.Actor, name string, cbFunc interface{}, args ...interface{}) {
	AfterGo(actor, name, 0, cbFunc, args...)
}

func Loop(actor *t.Actor, name string, delay, interval time.Duration, times int, cbFunc interface{}, args ...interface{}) {
	go func() {
		t := addTimer(actor, name, delay)
		cb, values := base.ReflectFunc(cbFunc, args)

		var count int
		for !IsStoped(actor, name) {
			select {
			case <-t.C:
				if times > 0 {
					if count < times {
						count++
						if count == times {
							go loopPush(actor, name, t, true, cb, values)
							return
						}

						t.Reset(interval)
						go loopPush(actor, name, t, false, cb, values)
					}

					continue
				}

				t.Reset(interval)
				go loopPush(actor, name, t, false, cb, values)
			}
		}
	}()
}

func loopPush(actor *t.Actor, name string, t *time.Timer, stop bool, cb reflect.Value, values []reflect.Value) {
	if actor == nil {
		dispatch.PushSystemMsg(triggerSystemMsg, name, t, stop, cb, values)
	} else {
		dispatch.PushActorMsg(actor, triggerActorMsg, name, t, stop, cb, values)
	}
}

func LoopDayMoment(name string, last time.Time, hour, min, sec int, cbFunc interface{}, args ...interface{}) {
	cb, values := base.ReflectFunc(cbFunc, args)
	if base.CheckMomentHappend(last, hour, min, sec) {
		cb.Call(values)
	}
	After(nil, name, base.GetMomentDelay(hour, min, sec), loopDayMoment, name, hour, min, sec, cb, values)

	//Loop(nil, name, base.GetMomentDelay(hour, min, sec), time.Hour*24, -1, cbFunc, args...)
}

func loopDayMoment(name string, hour, min, sec int, cb reflect.Value, values []reflect.Value) {
	After(nil, name, base.GetMomentDelay(hour, min, sec), loopDayMoment, name, hour, min, sec, cb, values)
	cb.Call(values)
}
