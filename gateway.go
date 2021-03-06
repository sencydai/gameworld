package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sencydai/gameworld/log"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"

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

	connCount   uint
	connCountMu sync.Mutex
)

func addConnCount() bool {
	connCountMu.Lock()
	defer connCountMu.Unlock()
	if connCount >= g.GetRealCount() {
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
	if !addConnCount() {
		conn.Close()
		return
	}
	if readSelfSalt(conn) != nil || readCheckKey(conn) != nil {
		subConnCount()
		conn.Close()
		return
	}

	var (
		tag     int
		dataLen int

		pid   uint32
		sysId byte
		cmdId byte
	)
	headSize := pack.HEAD_SIZE
	defTag := pack.DEFAULT_TAG
	buff := make([]byte, 0)
	reader := bytes.NewReader(buff)
	account := t.NewAccount(conn, reader)

	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}

		subConnCount()
		if g.IsGameClose() {
			return
		}

		account.Close()
		dispatch.PushSystemMsg(actormgr.OnAccountLogout, account)
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil || g.IsGameClose() {
			break
		}
		buff = append(buff, data...)
		if len(buff) < headSize {
			continue
		}

		reader.Reset(buff)
		pack.Read(reader, &tag)
		if tag != defTag {
			break
		}
		pack.Read(reader, &dataLen)
		if dataLen < 2 {
			break
		}
		size := headSize + dataLen
		if len(buff) < size {
			continue
		}
		data = buff[headSize:size]
		buff = buff[size:]
		reader.Reset(data)
		pack.Read(reader, &pid, &sysId, &cmdId)
		dispatch.PushClientMsg(account, sysId, cmdId, reader)
	}
}

func startGateWay() {
	server := http.NewServeMux()
	server.HandleFunc("/", handleConnection)

	if len(g.GameConfig.CertFile) == 0 || len(g.GameConfig.KeyFile) == 0 {
		go http.ListenAndServe(fmt.Sprintf(":%d", g.GameConfig.Port), server)
	} else {
		go http.ListenAndServeTLS(fmt.Sprintf(":%d", g.GameConfig.Port),
			g.GameConfig.CertFile, g.GameConfig.KeyFile, server)
	}

	log.Info("gateway started...")
}
