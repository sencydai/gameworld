package dispatch

import (
	"bytes"
	"reflect"
	"runtime/debug"

	proto "github.com/sencydai/gamecommon/protocol"
	"github.com/sencydai/gameworld/base"
	. "github.com/sencydai/gameworld/typedefine"
	"github.com/sencydai/utils/log"
)

//AccountMsgHandler 客户端信息处理接口定义
type AccountMsgHandler func(account *Account, reader *bytes.Reader)
type ActorMsgHandler func(actor *Actor, reader *bytes.Reader)

type CrossMsgHandler func(serverId int, reader *bytes.Reader)

type messageType byte

const (
	mtClient       messageType = 1 //客户端消息
	mtActor        messageType = 2 //玩家消息
	mtSystem       messageType = 3 //系统消息
	mtSystemAsynCB messageType = 4 //系统异步回调
	mtCrossMsg     messageType = 5 //跨服消息
)

type message struct {
	mtType messageType
	cbFunc interface{}
	cbArgs interface{}
	actor  *Actor
}

var (
	systemId        = proto.System
	messages        = make(chan *message, 5000)
	accountMsgs     = make(map[byte]AccountMsgHandler)
	actorMsgs       = make(map[int]ActorMsgHandler)
	crossMsgHandles = make(map[int]CrossMsgHandler)
)

//RegAccountMsgHandle 注册客户端消息处理函数
func RegAccountMsgHandle(cmdId byte, handle AccountMsgHandler) {
	accountMsgs[cmdId] = handle
}

func RegActorMsgHandle(sysId, cmdId byte, handle ActorMsgHandler) {
	cmd := (int(sysId) << 16) + int(cmdId)
	actorMsgs[cmd] = handle
}

func RegCrossMsg(msgId int, handle CrossMsgHandler) {
	crossMsgHandles[msgId] = handle
}

func PushClientMsg(cbArgs ...interface{}) {
	messages <- &message{mtType: mtClient, cbFunc: handleClientMsg, cbArgs: cbArgs}
}

func PushActorMsg(actor *Actor, handler interface{}, args ...interface{}) {
	messages <- &message{mtType: mtActor, cbFunc: handler, cbArgs: args, actor: actor}
}

func PushSystemMsg(handler interface{}, args ...interface{}) {
	messages <- &message{mtType: mtSystem, cbFunc: handler, cbArgs: args}
}

func PushSystemAsynMsg(cbFunc interface{}, asynFunc interface{}, asynArgs ...interface{}) {
	asynCb, values := base.ReflectFunc(asynFunc, asynArgs)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalf("PushSystemAsynMsg error: %v,%s", err, string(debug.Stack()))
			}
		}()
		values = asynCb.Call(values)
		messages <- &message{mtType: mtSystemAsynCB, cbFunc: cbFunc, cbArgs: values}
	}()
}

func PushCrossMsg(cbArgs ...interface{}) {
	messages <- &message{mtType: mtCrossMsg, cbFunc: handleCrossMsg, cbArgs: cbArgs}
}

func dispatch(msg *message) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("dispath msg error: %v,%s", err, string(debug.Stack()))
		}
	}()

	switch msg.mtType {
	case mtClient:
		args := msg.cbArgs.([]interface{})
		account := args[0].(*Account)
		sysId := args[1].(byte)
		cmdId := args[2].(byte)
		reader := args[3].(*bytes.Reader)
		handleClientMsg(account, sysId, cmdId, reader)
	case mtActor:
		actor := msg.actor
		if actor.Account != nil {
			args := msg.cbArgs.([]interface{})
			if len(args) == 0 {
				args = []interface{}{actor}
			} else {
				args = append([]interface{}{actor}, args...)
			}
			cb, values := base.ReflectFunc(msg.cbFunc, args)
			cb.Call(values)
		}
	case mtSystem:
		cb, values := base.ReflectFunc(msg.cbFunc, msg.cbArgs.([]interface{}))
		cb.Call(values)
	case mtSystemAsynCB:
		values := msg.cbArgs.([]reflect.Value)
		cb := reflect.ValueOf(msg.cbFunc)
		cb.Call(values)
	case mtCrossMsg:
		args := msg.cbArgs.([]interface{})
		msgId := args[0].(int)
		reader := args[1].(*bytes.Reader)
		handleCrossMsg(msgId, reader)
	}
}

func handleClientMsg(account *Account, sysId, cmdId byte, reader *bytes.Reader) {
	if sysId == systemId {
		if handle, ok := accountMsgs[cmdId]; ok {
			handle(account, reader)
		}
		return
	}
	if account.Actor == nil {
		return
	}
	cmd := (int(sysId) << 16) + int(cmdId)
	if handle, ok := actorMsgs[cmd]; ok {
		handle(account.Actor, reader)
	}
}

func handleCrossMsg(msgId int, reader *bytes.Reader) {
	if handle, ok := crossMsgHandles[msgId]; ok {
		handle(msgId, reader)
	}
}

func OnRun() {
	go func() {
		for msg := range messages {
			dispatch(msg)
		}
	}()
}
