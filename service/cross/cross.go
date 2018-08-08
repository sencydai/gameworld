package cross

import (
	"bytes"
	"fmt"
	"time"

	"github.com/nats-io/go-nats"
	"github.com/sencydai/gameworld/dispatch"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service"

	"github.com/sencydai/gameworld/log"
	"github.com/sencydai/gameworld/proto/pack"
)

var (
	conn *nats.Conn
)

func init() {
	service.RegGameStart(onGameStart)
}

func onDisconnect(*nats.Conn) {
	log.Info("======= crossmsg disconnect =======")
}

func onReconnect(*nats.Conn) {
	log.Info("======= crossmsg reconnect =======")
}

func onGameStart() {
	var err error
	conn, err = nats.Connect(
		g.GameConfig.CrossUrl,
		nats.ReconnectWait(time.Second*1),
		nats.MaxReconnects(-1),
		nats.DisconnectHandler(onDisconnect),
		nats.ReconnectHandler(onReconnect))
	if err != nil {
		panic(err)
	}

	_, err = conn.Subscribe("crossMsg_all", onRecvMsg)
	if err != nil {
		panic(err)
	}

	_, err = conn.Subscribe(fmt.Sprintf("crossMsg_%d", g.GameConfig.ServerId), onRecvMsg)
	if err != nil {
		panic(err)
	}

	log.Info("start crossmsg service")
}

func onRecvMsg(msg *nats.Msg) {
	reader := bytes.NewReader(msg.Data)
	var (
		serverId int
		msgId    int
	)
	pack.Read(reader, &serverId, &msgId)

	dispatch.PushCrossMsg(serverId, msgId, reader)
}

func PublishAllServerMsg(msgId int, data ...interface{}) {
	publishMsg("crossMsg_all", msgId, data...)
}

func PublishSpecServerMsg(serverId, msgId int, data ...interface{}) {
	publishMsg(fmt.Sprintf("crossMsg_%d", serverId), msgId, data...)
}

func publishMsg(sub string, msgId int, data ...interface{}) {
	if conn.IsReconnecting() {
		return
	}
	writer := pack.NewWriter(g.GameConfig.ServerId, msgId)
	pack.Write(writer, data...)

	conn.Publish(sub, writer.Bytes())
}
