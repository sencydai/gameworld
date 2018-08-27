package gm

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/json-iterator/go"
	"github.com/sencydai/gameworld/dispatch"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/log"
	"github.com/sencydai/gameworld/service"
)

type gmCmdReturnCode struct {
	Code int
	Data string
	Cost string
}

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
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

	values := make(map[string]string)
	for k, v := range r.Form {
		if len(v) == 0 {
			continue
		}
		values[strings.ToLower(k)] = v[0]
	}

	cmd, ok := values["cmd"]
	if !ok {
		w.WriteHeader(400)
		return
	}

	handle := service.GetGmHandle(cmd)
	if handle == nil {
		w.WriteHeader(400)
		return
	}

	delete(values, "cmd")

	now := time.Now()
	ch := make(chan []byte, 1)
	dispatch.PushSystemMsg(func(cmd string, values map[string]string) {
		gmCmdCode := &gmCmdReturnCode{}

		defer func() {
			if err := recover(); err != nil {
				gmCmdCode.Code = -1
				gmCmdCode.Data = fmt.Sprint(err)
			}
			gmCmdCode.Cost = fmt.Sprint(time.Since(now))
			data, _ := json.Marshal(gmCmdCode)
			cmdData, _ := json.Marshal(values)
			log.Infof("handle backdoor gmcmd [%s : %s] result: %s", cmd, string(cmdData), string(data))
			ch <- data
		}()

		gmCmdCode.Code, gmCmdCode.Data = handle(values)
	}, cmd, values)

	w.Write(<-ch)
}

func onGameStart() {
	server := http.NewServeMux()
	server.HandleFunc("/backdoor/gmcmd", handleGmCmd)
	if len(g.GameConfig.CertFile) == 0 || len(g.GameConfig.KeyFile) == 0 {
		go http.ListenAndServe(fmt.Sprintf(":%d", g.GameConfig.Port+1), server)
	} else {
		go http.ListenAndServeTLS(fmt.Sprintf(":%d", g.GameConfig.Port+1),
			g.GameConfig.CertFile, g.GameConfig.KeyFile, server)
	}

	log.Info("start backdoor service")
}
