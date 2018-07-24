package dispatch

import (
	"bytes"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"
	"github.com/sencydai/gameworld/base"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/log"
	t "github.com/sencydai/gameworld/typedefine"
	"github.com/sencydai/utils"
)

//AccountMsgHandler 客户端信息处理接口定义
type accountMsgHandler func(account *t.Account, reader *bytes.Reader)
type actorMsgHandler func(actor *t.Actor, reader *bytes.Reader)

type crossMsgHandler func(serverId int, reader *bytes.Reader)

type messageType byte

const (
	mtClientAccount messageType = 1  //客户端账号消息
	mtClientActor   messageType = 2  //客户端角色消息
	mtActor         messageType = 3  //玩家消息
	mtSystem        messageType = 4  //系统消息
	mtSystemAsynCB  messageType = 5  //系统异步回调
	mtCrossMsg      messageType = 6  //跨服消息
	mtTimerSystem   messageType = 7  //系统定时器消息
	mtTimerSystemGo messageType = 8  //系统定时器go消息
	mtTimerActor    messageType = 9  //玩家定时器消息
	mtTimerActorGo  messageType = 10 //玩家定时器go消息

	chanMessageLen = 20000
	clientTimeout  = time.Second * 5
)

type message struct {
	mtType messageType
	cbFunc interface{}
	cbArgs interface{}
	actor  *t.Actor
}

var (
	systemId        = proto.System
	messages        = make(chan *message, chanMessageLen)
	clientMessages  = make(chan *message, chanMessageLen/2)
	accountMsgs     = make(map[byte]accountMsgHandler)
	actorMsgs       = make(map[int]actorMsgHandler)
	crossMsgHandles = make(map[int]crossMsgHandler)

	TriggerSystemMsg   func(string, bool, reflect.Value, []reflect.Value)
	TriggerSystemMsgGo func(string, bool, reflect.Value, []reflect.Value)
	TriggerActorMsg    func(*t.Actor, string, bool, reflect.Value, []reflect.Value)
	TriggerActorMsgGo  func(*t.Actor, string, bool, reflect.Value, []reflect.Value)
)

//RegAccountMsgHandle 注册客户端消息处理函数
func RegAccountMsgHandle(cmdId byte, handle func(*t.Account, *bytes.Reader)) {
	accountMsgs[cmdId] = handle
}

func RegActorMsgHandle(sysId, cmdId byte, handle func(*t.Actor, *bytes.Reader)) {
	cmd := (int(sysId) << 16) + int(cmdId)
	actorMsgs[cmd] = handle
}

func RegCrossMsg(msgId int, handle func(int, *bytes.Reader)) {
	crossMsgHandles[msgId] = handle
}

func PushClientMsg(account *t.Account, sysId, cmdId byte, reader *bytes.Reader) {
	if sysId == systemId {
		if _, ok := accountMsgs[cmdId]; ok && account.Actor == nil {
			clientMessages <- &message{mtType: mtClientAccount, cbArgs: []interface{}{account, cmdId, reader}}
		}
	} else {
		actor := account.Actor
		if actor == nil {
			return
		}
		cmd := (int(sysId) << 16) + int(cmdId)
		if _, ok := actorMsgs[cmd]; ok {
			clientMessages <- &message{mtType: mtClientActor, cbArgs: []interface{}{cmd, reader}, actor: actor}
		}
	}
}

func PushActorMsg(actor *t.Actor, handler interface{}, args ...interface{}) {
	cb, values := base.ReflectFunc(handler, args)
	messages <- &message{mtType: mtActor, cbFunc: cb, cbArgs: values, actor: actor}
}

func PushSystemMsg(handler interface{}, args ...interface{}) {
	cb, values := base.ReflectFunc(handler, args)
	messages <- &message{mtType: mtSystem, cbFunc: cb, cbArgs: values}
}

func PushSystemAsynMsg(cbFunc interface{}, asynFunc interface{}, asynArgs ...interface{}) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalf("%v,%s", err, string(debug.Stack()))
			}
		}()
		asynCb, values := base.ReflectFunc(asynFunc, asynArgs)
		messages <- &message{mtType: mtSystemAsynCB, cbFunc: reflect.ValueOf(cbFunc), cbArgs: asynCb.Call(values)}
	}()
}

func PushCrossMsg(serverId, msgId int, reader *bytes.Reader) {
	if _, ok := crossMsgHandles[msgId]; !ok {
		return
	}
	messages <- &message{mtType: mtCrossMsg, cbArgs: []interface{}{serverId, msgId, reader}}
}

func PushTimerSysMsg(args ...interface{}) {
	messages <- &message{mtType: mtTimerSystem, cbArgs: args}
}

func PushTimerSysGoMsg(args ...interface{}) {
	messages <- &message{mtType: mtTimerSystemGo, cbArgs: args}
}

func PushTimerActorMsg(actor *t.Actor, args ...interface{}) {
	messages <- &message{mtType: mtTimerActor, cbArgs: args, actor: actor}
}

func PushTimerActorGoMsg(actor *t.Actor, args ...interface{}) {
	messages <- &message{mtType: mtTimerActorGo, cbArgs: args, actor: actor}
}

func dispatch(msg *message) {
	defer func() {
		if err := recover(); err != nil {
			if err == pack.ReadEOF {
				log.Errorf("%v ==> %s", err, utils.FileLine(5))
			} else {
				log.Fatalf("%v,%s", err, string(debug.Stack()))
			}
		}
	}()

	switch msg.mtType {
	case mtClientAccount:
		args := msg.cbArgs.([]interface{})
		account := args[0].(*t.Account)
		if account.IsClose() {
			return
		}
		cmdId := args[1].(byte)
		reader := args[2].(*bytes.Reader)
		tick := args[3].(time.Time)
		if time.Since(tick) >= clientTimeout {
			account.Close()
			return
		}

		accountMsgs[cmdId](account, reader)
	case mtClientActor:
		actor := msg.actor
		account := actor.Account
		if account == nil || account.IsClose() {
			return
		}
		args := msg.cbArgs.([]interface{})
		cmd := args[0].(int)
		reader := args[1].(*bytes.Reader)
		tick := args[2].(time.Time)
		if time.Since(tick) >= clientTimeout {
			account.Close()
			return
		}
		actorMsgs[cmd](actor, reader)
	case mtActor:
		actor := msg.actor
		account := actor.Account
		if account == nil || account.IsClose() {
			return
		}
		cb := msg.cbFunc.(reflect.Value)
		args := msg.cbArgs.([]reflect.Value)
		v := reflect.ValueOf(actor)
		if len(args) == 0 {
			args = []reflect.Value{v}
		} else {
			args = append([]reflect.Value{v}, args...)
		}
		cb.Call(args)
	case mtSystem:
		cb := msg.cbFunc.(reflect.Value)
		args := msg.cbArgs.([]reflect.Value)
		cb.Call(args)
	case mtSystemAsynCB:
		cb := msg.cbFunc.(reflect.Value)
		values := msg.cbArgs.([]reflect.Value)
		cb.Call(values)
	case mtCrossMsg:
		args := msg.cbArgs.([]interface{})
		serverId := args[0].(int)
		msgId := args[1].(int)
		reader := args[2].(*bytes.Reader)
		crossMsgHandles[msgId](serverId, reader)
	case mtTimerSystem:
		args := msg.cbArgs.([]interface{})
		name := args[0].(string)
		stop := args[1].(bool)
		cb := args[2].(reflect.Value)
		values := args[3].([]reflect.Value)
		TriggerSystemMsg(name, stop, cb, values)
	case mtTimerSystemGo:
		args := msg.cbArgs.([]interface{})
		name := args[0].(string)
		stop := args[1].(bool)
		cb := args[2].(reflect.Value)
		values := args[3].([]reflect.Value)
		TriggerSystemMsgGo(name, stop, cb, values)
	case mtTimerActor:
		actor := msg.actor
		account := actor.Account
		if account == nil || account.IsClose() {
			return
		}
		args := msg.cbArgs.([]interface{})
		name := args[0].(string)
		stop := args[1].(bool)
		cb := args[2].(reflect.Value)
		values := args[3].([]reflect.Value)
		TriggerActorMsg(actor, name, stop, cb, values)
	case mtTimerActorGo:
		actor := msg.actor
		account := actor.Account
		if account == nil || account.IsClose() {
			return
		}
		args := msg.cbArgs.([]interface{})
		name := args[0].(string)
		stop := args[1].(bool)
		cb := args[2].(reflect.Value)
		values := args[3].([]reflect.Value)
		TriggerActorMsgGo(actor, name, stop, cb, values)
	}
}

func OnRun() {
	go func() {
		for msg := range messages {
			if g.IsGameClose() {
				break
			}
			dispatch(msg)
		}
	}()

	go func() {
		waitTime := time.Millisecond
		for msg := range clientMessages {
			if g.IsGameClose() {
				break
			}
			args := msg.cbArgs.([]interface{})
			if msg.actor == nil {
				account := args[0].(*t.Account)
				if account.IsClose() {
					continue
				}
			} else {
				account := msg.actor.Account
				if account == nil || account.IsClose() {
					continue
				}
			}

			msg.cbArgs = append(args, time.Now())
			messages <- msg
			time.Sleep(waitTime)
		}
	}()
}
