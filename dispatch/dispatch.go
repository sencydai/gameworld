package dispatch

import (
	"bytes"
	proto "github.com/sencydai/gamecommon/protocol"
	"github.com/sencydai/gameworld/base"
	. "github.com/sencydai/gameworld/typedefine"
	"github.com/sencydai/utils/log"
	"reflect"
	"runtime/debug"
)

type AccountMsgHandler func(*Account, *bytes.Reader)
type ActorMsgHandler func(*Actor, *bytes.Reader)

var (
	systemId    = proto.System
	messages    = make(chan *Message, 5000)
	accountMsgs = make(map[byte]AccountMsgHandler)
	actorMsgs   = make(map[int]ActorMsgHandler)
)

func RegAccountMsgHandle(cmdId byte, handle AccountMsgHandler) {
	accountMsgs[cmdId] = handle
}

func RegActorMsgHandle(sysId, cmdId byte, handle ActorMsgHandler) {
	cmd := (int(sysId) << 16) + int(cmdId)
	actorMsgs[cmd] = handle
}

func PushClientMsg(cbArgs ...interface{}) {
	messages <- &Message{MTType: MTClient, CBFunc: handleClientMsg, CBArgs: cbArgs}
}

func PushSystemMsg(handler interface{}, args ...interface{}) {
	messages <- &Message{MTType: MTSystem, CBFunc: handler, CBArgs: args}
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
		messages <- &Message{MTType: MTSystemAsynCB, CBFunc: cbFunc, CBArgs: values}
	}()
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

func dispatch(msg *Message) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("dispatch msg error: %v,%s", err, string(debug.Stack()))
		}
	}()

	switch msg.MTType {
	case MTClient:
		args := msg.CBArgs.([]interface{})
		account := args[0].(*Account)
		sysId := args[1].(byte)
		cmdId := args[2].(byte)
		reader := args[3].(*bytes.Reader)
		handleClientMsg(account, sysId, cmdId, reader)
	case MTSystem:
		cb, values := base.ReflectFunc(msg.CBFunc, msg.CBArgs.([]interface{}))
		cb.Call(values)
	case MTSystemAsynCB:
		values := msg.CBArgs.([]reflect.Value)
		cb := reflect.ValueOf(msg.CBFunc)
		cb.Call(values)
	}
}

func OnRun() {
	go func() {
		for msg := range messages {
			dispatch(msg)
		}
	}()
}
