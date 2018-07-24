package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"
	"github.com/sencydai/gameworld/log"

	"github.com/sencydai/gameworld/dispatch"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service/actormgr"
	t "github.com/sencydai/gameworld/typedefine"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024 * 10,
		CheckOrigin: func(*http.Request) bool {
			return true
		},
	}

	systemId = proto.System

	connCount   = 0
	connCountMu sync.Mutex
)

func addConnCount() bool {
	connCountMu.Lock()
	defer connCountMu.Unlock()
	if connCount >= g.GameConfig.MaxConnection {
		return false
	}
	connCount++
	return true
}

func subConnCount() {
	connCountMu.Lock()
	defer connCountMu.Unlock()
	connCount--
}

func readSelfSalt(conn *websocket.Conn) error {
	_, data, err := conn.ReadMessage()
	if err != nil {
		return err
	}

	var selfSalt uint32
	reader := bytes.NewReader(data)
	pack.Read(reader, &selfSalt)

	return conn.WriteMessage(websocket.BinaryMessage, pack.GetBytes(uint32(rand.Int31())))
}

func readCheckKey(conn *websocket.Conn) error {
	_, data, err := conn.ReadMessage()
	if err != nil {
		return err
	}

	var checkKey int16
	reader := bytes.NewReader(data)
	pack.Read(reader, &checkKey)

	return nil
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if !addConnCount() || readSelfSalt(conn) != nil || readCheckKey(conn) != nil {
		conn.Close()
		return
	}

	account := t.NewAccount(conn)

	defer func() {
		if err := recover(); err != nil {
		}

		subConnCount()
		if g.IsGameClose() {
			return
		}

		account.Close()
		dispatch.PushSystemMsg(actormgr.OnAccountLogout, account)
	}()

	var (
		tag     int
		dataLen int

		pid   uint32
		sysId byte
		cmdId byte
	)
	headSize := pack.HEAD_SIZE
	defTag := pack.DEFAULT_TAG
	for {
		_, data, err := conn.ReadMessage()
		if err != nil || g.IsGameClose() || len(data) < headSize {
			break
		}

		reader := bytes.NewReader(data)
		pack.Read(reader, &tag)
		if tag != defTag {
			break
		}
		pack.Read(reader, &dataLen)
		if dataLen < 2 {
			break
		}
		data = data[headSize:]
		if dataLen != len(data) {
			break
		}
		reader.Reset(data)
		pack.Read(reader, &pid, &sysId, &cmdId)
		dispatch.PushClientMsg(account, sysId, cmdId, reader)
	}
}

func startGateWay() {
	server := http.NewServeMux()
	server.HandleFunc("/", handleConnection)
	go http.ListenAndServe(fmt.Sprintf(":%d", g.GameConfig.Port), server)

	log.Info("gateway started...")
}
