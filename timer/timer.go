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
	mutex       sync.Mutex
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
		if _, ok := sysTimers[name]; ok {
			log.Errorf("timer %s is existed", name)
			return nil
		}
		t := time.NewTimer(delay)
		sysTimers[name] = t
		return t
	}

	actors, ok := actorTimers[actor]
	if !ok {
		actors = make(map[string]*time.Timer)
		actorTimers[actor] = actors
	} else if _, ok = actors[name]; ok {
		log.Errorf("timer %s is existed", name)
		return nil
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
	mutex.Lock()
	defer mutex.Unlock()

	if actor == nil {
		_, ok := sysTimers[name]
		return !ok
	}
	actors, ok := actorTimers[actor]
	if !ok {
		return !ok
	}
	_, ok = actors[name]
	return !ok
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

func triggerSystemMsg(name string, stop bool, cb reflect.Value, values []reflect.Value) {
	if stop {
		if !StopTimer(nil, name) {
			return
		}
	} else if IsStoped(nil, name) {
		return
	}

	cb.Call(values)
}

func triggerSystemMsgGo(name string, stop bool, cb reflect.Value, values []reflect.Value) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalf("timer %s ==> %v,%s", name, err, string(debug.Stack()))
			}
		}()

		triggerSystemMsg(name, stop, cb, values)
	}()
}

func triggerActorMsg(actor *t.Actor, name string, stop bool, cb reflect.Value, values []reflect.Value) {
	if stop {
		if !StopTimer(actor, name) {
			return
		}
	} else if IsStoped(actor, name) {
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

func triggerActorMsgGo(actor *t.Actor, name string, stop bool, cb reflect.Value, values []reflect.Value) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalf("timer %s ==> %v,%s", name, err, string(debug.Stack()))
			}
		}()

		triggerActorMsg(actor, name, stop, cb, values)
	}()

}

func After(actor *t.Actor, name string, delay time.Duration, cbFunc interface{}, args ...interface{}) {
	t := addTimer(actor, name, delay)
	if t == nil {
		return
	}
	go func() {
		select {
		case <-t.C:
			cb, values := base.ReflectFunc(cbFunc, args)
			if actor == nil {
				dispatch.PushSystemMsg(triggerSystemMsg, name, true, cb, values)
			} else {
				dispatch.PushActorMsg(actor, triggerActorMsg, name, true, cb, values)
			}
		}
	}()
}

func AfterGo(actor *t.Actor, name string, delay time.Duration, cbFunc interface{}, args ...interface{}) {
	t := addTimer(actor, name, delay)
	if t == nil {
		return
	}
	go func() {
		select {
		case <-t.C:
			cb, values := base.ReflectFunc(cbFunc, args)
			if actor == nil {
				dispatch.PushSystemMsg(triggerSystemMsgGo, name, true, cb, values)
			} else {
				dispatch.PushActorMsg(actor, triggerActorMsgGo, name, true, cb, values)
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
	t := addTimer(actor, name, delay)
	if t == nil {
		return
	}
	go func() {
		cb, values := base.ReflectFunc(cbFunc, args)

		var count int
		for !IsStoped(actor, name) {
			select {
			case <-t.C:
				if times > 0 {
					if count < times {
						count++
						if count == times {
							go loopPush(actor, name, true, cb, values)
							return
						}

						t.Reset(interval)
						go loopPush(actor, name, false, cb, values)
					}

					continue
				}

				t.Reset(interval)
				go loopPush(actor, name, false, cb, values)
			}
		}
	}()
}

func loopPush(actor *t.Actor, name string, stop bool, cb reflect.Value, values []reflect.Value) {
	if actor == nil {
		dispatch.PushSystemMsg(triggerSystemMsg, name, stop, cb, values)
	} else {
		dispatch.PushActorMsg(actor, triggerActorMsg, name, stop, cb, values)
	}
}

func LoopDayMoment(name string, last time.Time, hour, min, sec int, cbFunc interface{}, args ...interface{}) {
	if base.CheckMomentHappend(last, hour, min, sec) {
		cb, values := base.ReflectFunc(cbFunc, args)
		cb.Call(values)
	}

	Loop(nil, name, base.GetMomentDelay(hour, min, sec), time.Hour*24, -1, cbFunc, args...)
}
