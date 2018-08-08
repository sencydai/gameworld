package chat

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"

	//"github.com/sencydai/gameworld/data"
	c "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/data"
	"github.com/sencydai/gameworld/dispatch"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/log"
	"github.com/sencydai/gameworld/service"
	t "github.com/sencydai/gameworld/typedefine"
)

const (
	maxChatLen = 255
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func init() {
	dispatch.RegActorMsgHandle(proto.Chat, proto.ChatCSendChatMsg, onSendChat)
}

func onSendChat(actor *t.Actor, reader *bytes.Reader) {
	//tick := time.Now()
	var (
		channel   byte
		content   string
		superLink string
	)
	pack.Read(reader, &channel, &content, &superLink)

	if len(content) == 0 && len(superLink) == 0 {
		return
	}

	if strings.HasPrefix(content, "@") && actor.Account.GmLevel > 0 {
		text := strings.TrimLeft(content, "@")
		text = strings.TrimSpace(text)

		r := []rune(text)
		var index int
		for ; index < len(r); index++ {
			if r[index] == ' ' {
				break
			}
		}

		cmd := string(r[0:index])
		cmd = strings.ToLower(cmd)
		handle := service.GetGmHandle(cmd)
		if handle != nil {
			if index < len(text) {
				text = string(r[index+1:])
			} else {
				text = ""
			}
			values := make(map[string]string)
			values["actor"] = strconv.FormatInt(actor.ActorId, 10)
			for _, s := range strings.Split(text, "&") {
				ss := strings.Split(s, "=")
				if len(ss) == 2 {
					k := strings.ToLower(strings.TrimSpace(ss[0]))
					v := strings.TrimSpace(ss[1])
					values[k] = v
				}
			}

			code, data := handle(values)
			valueData, _ := json.Marshal(values)
			log.Infof("handle client gmcmd [%s : %s] code=%d,data=%s", cmd, string(valueData), code, data)
			actor.SendTips(fmt.Sprintf("code=%d,data=%s", code, data))
			return
		}
	}

	if len([]rune(content)) > maxChatLen {
		return
	}
	content = g.FilterText(content)

	decorData := actor.GetDecorData()
	head := decorData[c.LDTHead]
	frame := decorData[c.LDTFrame]
	chat := decorData[c.LDTChat]

	writer := pack.AllocPack(
		proto.Chat,
		proto.ChatSSendChatMsg,
		channel,
		float64(actor.ActorId),
		actor.ActorName,
		content,
		superLink,
		head.Id,
		frame.Id,
		actor.Level,
		chat.Id,
		1,
		g.GameConfig.ServerId,
	)

	data.BroadcastWriter(writer)
	//log.Infof("actor(%d),cost(%v) chat: %s", actor.ActorId, time.Since(tick), content)
}

func BroadcastMsg(noticeType byte, channel string, roll byte, msg string) {
	data.Broadcast(proto.Chat, proto.ChatSSysMsg, noticeType, channel, roll, msg)
}
