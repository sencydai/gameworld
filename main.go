package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/sencydai/utils/log"

	"github.com/sencydai/gameworld/data"
	"github.com/sencydai/gameworld/dispatch"
	"github.com/sencydai/gameworld/engine"
	"github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service"

	"github.com/sencydai/gameworld/service/actormgr"
	_ "github.com/sencydai/gameworld/service/backdoor"
	_ "github.com/sencydai/gameworld/service/bag"
	_ "github.com/sencydai/gameworld/service/cross"
)

func init() {
	service.RegGm("reload", onReLoadConfig)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().Unix())

	if buff, err := ioutil.ReadFile("config.json"); err != nil {
		fmt.Printf("load config file error: %s\n", err.Error())
		return
	} else if err = json.Unmarshal(buff, &gconfig.GameConfig); err != nil {
		fmt.Printf("parse config file error: %s\n", err.Error())
		return
	}

	gconfig.ServerIdML = int64(gconfig.GameConfig.ServerId) << 32

	if err := log.SetFileName(fmt.Sprintf("%s/server_%d_%s.log", gconfig.GameConfig.LogPath, gconfig.GameConfig.ServerId, time.Now().Format("20060102"))); err != nil {
		fmt.Printf("create log file error: %s\n", err.Error())
		return
	}
	log.SetLevel(gconfig.GameConfig.LogLevel)

	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("%v: %s", err, string(debug.Stack()))
		}
		log.Close()
	}()

	tick := time.Now()
	log.Info("server starting...")

	//加载配置表
	log.Info("load config datas...")
	gconfig.LoadConfigs(gconfig.GameConfig.ConfigPath)

	//加载敏感词
	log.Info("load filtertext...")
	gconfig.LoadFilterTexts(gconfig.GameConfig.ConfigPath)

	//加载随机昵称
	log.Info("load random names...")
	gconfig.LoadRandomNames(gconfig.GameConfig.ConfigPath)

	service.OnConfigReloadFinish()

	//数据库引擎
	log.Info("start database engine...")
	engine.InitEngine()

	data.OnLoadSystemData()
	actormgr.OnLoadMaxActorId()
	actormgr.OnLoadAllActorNames()

	dispatch.OnRun()

	service.OnGameStart()

	startGateWay()

	log.Info("============================================================")
	log.Infof("server started success, cost:%v", time.Since(tick))
	log.Info("============================================================")

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, os.Interrupt, os.Kill)
	select {
	case <-signalC:
	}

	tick = time.Now()
	log.Info("server closing...")
	gconfig.CloseGame()

	//保存所有玩家数据
	waitClose := make(chan bool, 1)
	dispatch.PushSystemMsg(func() {
		defer func() {
			waitClose <- true
		}()
		actormgr.OnGameClose()
		data.OnGameClose()
	})

	<-waitClose

	log.Infof("server closed success, cost:%v", time.Since(tick))
}

func onReLoadConfig(values url.Values) (int, string) {
	log.Info("=================reload config===================")
	//加载配置表
	log.Info("load config datas...")
	gconfig.LoadConfigs(gconfig.GameConfig.ConfigPath)

	//加载敏感词
	log.Info("load filtertext...")
	gconfig.LoadFilterTexts(gconfig.GameConfig.ConfigPath)

	log.Info("load random names...")
	gconfig.LoadRandomNames(gconfig.GameConfig.ConfigPath)

	service.OnConfigReloadFinish()

	log.Info("==============reload config success===============")

	return 0, "reload success"
}
