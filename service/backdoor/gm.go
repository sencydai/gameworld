package backdoor

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/json-iterator/go"
	"github.com/sencydai/gameworld/dispatch"
	"github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/utils/log"
)

type gmCmdReturnCode struct {
	Code int
	Data string
}

var (
	json      = jsoniter.ConfigCompatibleWithStandardLibrary
	gmCmdCode = &gmCmdReturnCode{}
)

func init() {
	service.RegGameStart(onGameStart)
}

func handleGmCmd(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(403)
		return
	}
	r.ParseForm()
	cmd := r.Form.Get("cmd")
	if cmd == "" {
		w.WriteHeader(400)
		return
	}

	handle := service.GetGmHandle(cmd)
	if handle == nil {
		w.WriteHeader(400)
		return
	}

	ch := make(chan []byte, 1)
	dispatch.PushSystemMsg(func(values url.Values) {
		defer func() {
			if err := recover(); err != nil {
				gmCmdCode.Code = -1
				gmCmdCode.Data = fmt.Sprint(err)
			}
			data, _ := json.Marshal(gmCmdCode)
			log.Infof("handle cmd: %v, result: %s", values, string(data))
			ch <- data
		}()
		gmCmdCode.Code, gmCmdCode.Data = handle(values)
	})

	w.Write(<-ch)
}

func onGameStart() {
	server := http.NewServeMux()
	server.HandleFunc("/backdoor/gmcmd", handleGmCmd)

	go http.ListenAndServe(fmt.Sprintf(":%d", gconfig.GameConfig.Port+1), server)
}
