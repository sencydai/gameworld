package dispatch

import (
	"bytes"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/sencydai/gameworld/base"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/log"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
	t "github.com/sencydai/gameworld/typedefine"
)

//AccountMsgHandler 客户端信息处理接口定义
type accountMsgHandler func(account *t.Account, reader *bytes.Reader)
type actorMsgHandler func(actor *t.Actor, reader *bytes.Reader)

type crossMsgHandler func(serverId int, reader *bytes.Reader)

type messageType = byte

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
)

var (
	systemId = proto.System
	messages chan *t.Message

	accountMsgs     = make(map[byte]accountMsgHandler)
	actorMsgs       = make(map[int]actorMsgHandler)
	crossMsgHandles = make(map[int]crossMsgHandler)

	TriggerSystemMsg   func(string, *time.Timer, bool, reflect.Value, []reflect.Value)
	TriggerSystemMsgGo func(string, *time.Timer, bool, reflect.Value, []reflect.Value)
	TriggerActorMsg    func(*t.Actor, string, *time.Timer, bool, reflect.Value, []reflect.Value)
	TriggerActorMsgGo  func(*t.Actor, string, *time.Timer, bool, reflect.Value, []reflect.Value)
)

func InitData(maxActorCount uint) {
	messages = make(chan *t.Message, maxActorCount*10)
}

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
			msg := account.Msg
			msg.MtType = mtClientAccount
			msg.CBArgs = cmdId
			writerClientMsg(account, sysId, cmdId)
		}
	} else {
		if account.Actor == nil {
			return
		}
		cmd := (int(sysId) << 16) + int(cmdId)
		if _, ok := actorMsgs[cmd]; ok {
			msg := account.Msg
			msg.MtType = mtClientActor
			msg.CBArgs = cmd
			writerClientMsg(account, sysId, cmdId)
		}
	}
}

const (
	msgRouterTimeout = time.Millisecond * 5
	msgHandleTimeout = time.Second * 3
)

func writerClientMsg(account *t.Account, sysId, cmdId byte) {
	select {
	case messages <- account.Msg:
	case <-time.After(msgRouterTimeout):
		log.Warnf("server is very busy now,discard account %d msg %d %d", account.AccountId, sysId, cmdId)
		return
	}

	select {
	case <-account.GetCmdCh():
	case <-time.After(msgHandleTimeout):
		account.Close()
		log.Warnf("server is very busy now,disconnect account %d", account.AccountId)
		g.ReduceRealCount()
	}
}

func PushActorMsg(actor *t.Actor, handler interface{}, args ...interface{}) {
	cb, values := base.ReflectFunc(handler, args)
	messages <- &t.Message{MtType: mtActor, CBFunc: cb, CBArgs: values, Actor: actor}
}

func PushSystemMsg(handler interface{}, args ...interface{}) {
	cb, values := base.ReflectFunc(handler, args)
	messages <- &t.Message{MtType: mtSystem, CBFunc: cb, CBArgs: values}
}

func PushSystemAsynMsg(cbFunc interface{}, asynFunc interface{}, asynArgs ...interface{}) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalf("%v,%s", err, string(debug.Stack()))
			}
		}()
		asynCb, values := base.ReflectFunc(asynFunc, asynArgs)
		messages <- &t.Message{MtType: mtSystemAsynCB, CBFunc: reflect.ValueOf(cbFunc), CBArgs: asynCb.Call(values)}
	}()
}

func PushCrossMsg(serverId, msgId int, reader *bytes.Reader) {
	if _, ok := crossMsgHandles[msgId]; !ok {
		return
	}
	messages <- &t.Message{MtType: mtCrossMsg, CBArgs: []interface{}{serverId, msgId, reader}}
}

func PushTimerSysMsg(args ...interface{}) {
	messages <- &t.Message{MtType: mtTimerSystem, CBArgs: args}
}

func PushTimerSysGoMsg(args ...interface{}) {
	messages <- &t.Message{MtType: mtTimerSystemGo, CBArgs: args}
}

func PushTimerActorMsg(actor *t.Actor, args ...interface{}) {
	messages <- &t.Message{MtType: mtTimerActor, CBArgs: args, Actor: actor}
}

func PushTimerActorGoMsg(actor *t.Actor, args ...interface{}) {
	messages <- &t.Message{MtType: mtTimerActorGo, CBArgs: args, Actor: actor}
}

func dispatch(msg *t.Message) {
	var msgCh chan bool
	defer func() {
		if msgCh != nil {
			msgCh <- true
		}

		if err := recover(); err != nil {
			if err == pack.ReadEOF {
				log.Fatalf("%v ==> %s", err, base.FileLine(5))
			} else {
				log.Fatalf("%v,%s", err, string(debug.Stack()))
			}
		}
	}()

	switch msg.MtType {
	case mtClientAccount:
		account := msg.Account
		msgCh = account.GetCmdCh()
		if account.IsClose() {
			return
		}
		accountMsgs[msg.CBArgs.(byte)](account, msg.Reader)
	case mtClientActor:
		account := msg.Account
		msgCh = account.GetCmdCh()
		if account.IsClose() {
			return
		}
		actorMsgs[msg.CBArgs.(int)](account.Actor, msg.Reader)
	case mtActor:
		actor := msg.Actor
		account := actor.Account
		if account == nil || account.IsClose() {
			return
		}
		cb := msg.CBFunc.(reflect.Value)
		args := msg.CBArgs.([]reflect.Value)
		v := reflect.ValueOf(actor)
		if len(args) == 0 {
			args = []reflect.Value{v}
		} else {
			args = append([]reflect.Value{v}, args...)
		}
		cb.Call(args)
	case mtSystem:
		cb := msg.CBFunc.(reflect.Value)
		args := msg.CBArgs.([]reflect.Value)
		cb.Call(args)
	case mtSystemAsynCB:
		cb := msg.CBFunc.(reflect.Value)
		values := msg.CBArgs.([]reflect.Value)
		cb.Call(values)
	case mtCrossMsg:
		args := msg.CBArgs.([]interface{})
		serverId := args[0].(int)
		msgId := args[1].(int)
		reader := args[2].(*bytes.Reader)
		crossMsgHandles[msgId](serverId, reader)
	case mtTimerSystem:
		args := msg.CBArgs.([]interface{})
		name := args[0].(string)
		t := args[1].(*time.Timer)
		stop := args[2].(bool)
		cb := args[3].(reflect.Value)
		values := args[4].([]reflect.Value)
		TriggerSystemMsg(name, t, stop, cb, values)
	case mtTimerSystemGo:
		args := msg.CBArgs.([]interface{})
		name := args[0].(string)
		t := args[1].(*time.Timer)
		stop := args[2].(bool)
		cb := args[3].(reflect.Value)
		values := args[4].([]reflect.Value)
		TriggerSystemMsgGo(name, t, stop, cb, values)
	case mtTimerActor:
		actor := msg.Actor
		account := actor.Account
		if account == nil || account.IsClose() {
			return
		}
		args := msg.CBArgs.([]interface{})
		name := args[0].(string)
		t := args[1].(*time.Timer)
		stop := args[2].(bool)
		cb := args[3].(reflect.Value)
		values := args[4].([]reflect.Value)
		TriggerActorMsg(actor, name, t, stop, cb, values)
	case mtTimerActorGo:
		actor := msg.Actor
		account := actor.Account
		if account == nil || account.IsClose() {
			return
		}
		args := msg.CBArgs.([]interface{})
		name := args[0].(string)
		t := args[1].(*time.Timer)
		stop := args[2].(bool)
		cb := args[3].(reflect.Value)
		values := args[4].([]reflect.Value)
		TriggerActorMsgGo(actor, name, t, stop, cb, values)
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
}
