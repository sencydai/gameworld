package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sencydai/gamecommon/pack"
	"github.com/sencydai/utils/log"

	"github.com/sencydai/gameworld/dispatch"
	"github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service/actormgr"
	. "github.com/sencydai/gameworld/typedefine"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024 * 10,
		CheckOrigin: func(*http.Request) bool {
			return true
		},
	}

	connCount   = 0
	connCountMu sync.Mutex
)

func addConnCount() bool {
	connCountMu.Lock()
	defer connCountMu.Unlock()
	if connCount >= gconfig.GameConfig.MaxConnection {
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
	errChan := make(chan bool, 2)
	account := NewAccount(conn)
	//读
	go func(errChan chan bool) {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalf("%v : %s", err, string(debug.Stack()))
			}
			errChan <- true
		}()
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if gconfig.IsGameClose() {
				break
			}
			if len(data) < pack.HEAD_SIZE {
				break
			}
			reader := bytes.NewReader(data)
			var tag int
			pack.Read(reader, &tag)
			if tag != pack.DEFAULT_TAG {
				break
			}

			var dataLen int
			pack.Read(reader, &dataLen)
			if dataLen < 2 {
				break
			}
			data = data[pack.HEAD_SIZE:]
			if dataLen != len(data) {
				break
			}
			reader.Reset(data)
			var (
				pid   uint32
				sysId byte
				cmdId byte
			)
			pack.Read(reader, &pid, &sysId, &cmdId)
			dispatch.PushClientMsg(account, sysId, cmdId, reader)
			time.Sleep(time.Millisecond * 150)
		}
	}(errChan)

	//写
	go func(errChan chan bool) {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalf("%v : %s", err, string(debug.Stack()))
			}
			errChan <- true
		}()

		for data := range account.GetData() {
			if gconfig.IsGameClose() {
				break
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				break
			}
		}
	}(errChan)

	<-errChan

	subConnCount()
	account.Close()
	if gconfig.IsGameClose() {
		return
	}
	dispatch.PushSystemMsg(actormgr.OnAccountLogout, account)
}

func startGateWay() {
	server := http.NewServeMux()
	server.HandleFunc("/", handleConnection)
	go http.ListenAndServe(fmt.Sprintf(":%d", gconfig.GameConfig.Port), server)

	log.Info("gateway started...")
}
