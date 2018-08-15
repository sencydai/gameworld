package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	//"runtime/pprof"
	"strconv"
	"time"

	"github.com/sencydai/gameworld/log"

	"github.com/sencydai/gameworld/data"
	"github.com/sencydai/gameworld/dispatch"
	"github.com/sencydai/gameworld/engine"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service"

	_ "github.com/sencydai/gameworld/rank"
	"github.com/sencydai/gameworld/service/actormgr"
	_ "github.com/sencydai/gameworld/service/backdoor"
	_ "github.com/sencydai/gameworld/service/bag"
	_ "github.com/sencydai/gameworld/service/building"
	_ "github.com/sencydai/gameworld/service/chat"
	_ "github.com/sencydai/gameworld/service/cross"
	_ "github.com/sencydai/gameworld/service/guard"
	_ "github.com/sencydai/gameworld/service/hero"
	_ "github.com/sencydai/gameworld/service/hero/heroartifact"
	_ "github.com/sencydai/gameworld/service/hero/heroequip"
	_ "github.com/sencydai/gameworld/service/lord"
	_ "github.com/sencydai/gameworld/service/lord/lorddecor"
	_ "github.com/sencydai/gameworld/service/lord/lordequip"
	_ "github.com/sencydai/gameworld/service/lord/lordlevel"
	_ "github.com/sencydai/gameworld/service/lord/lordskill"
	_ "github.com/sencydai/gameworld/service/lord/lordtalent"
	_ "github.com/sencydai/gameworld/service/mainfuben"
	_ "github.com/sencydai/gameworld/service/ranksystem"
	_ "github.com/sencydai/gameworld/service/systemopen"
	_ "github.com/sencydai/gameworld/service/worldboss"
)

func init() {
	service.RegGm("reload", onReLoadConfig)
	service.RegGm("maxaccount", onSetMaxAccount)
	service.RegGm("realaccount", onSetRealAccount)
	service.RegGameStart(onGameStart)
}

func onGameStart() {
	monitorMaxAccountCount()
	monitorSystemTime()
}

func monitorMaxAccountCount() {
	//服务器不拥堵，增加人数上限,否则减少人数上限
	go func() {
		loop := time.Second * 5
		min := time.Microsecond * 200
		chSync := make(chan bool, 1)
		for {
			select {
			case <-time.After(loop):
				tick := time.Now()
				dispatch.PushSystemMsg(func() {
					chSync <- true
				})

				<-chSync

				delay := time.Since(tick)
				if delay < min {
					count := uint(data.GetAccountCount() * 2)
					if count >= g.GetMaxCount() {
						g.AddMaxCount()
					}
				}
			}
		}
	}()
}

func monitorSystemTime() {
	//监控系统时间改变
	go func() {
		loop := time.Second * 10
		for {
			tick := time.Now()
			select {
			case <-time.After(loop):
				delay := time.Now().Unix() - tick.Unix()
				if math.Abs(float64(delay)) >= 20 {
					log.Info("system time changed")
					dispatch.PushSystemMsg(service.OnSystemTimeChange)
				}
			}
		}
	}()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UnixNano())

	if buff, err := ioutil.ReadFile("config.json"); err != nil {
		fmt.Printf("load config file error: %s\n", err.Error())
		return
	} else if err = json.Unmarshal(buff, &g.GameConfig); err != nil {
		fmt.Printf("parse config file error: %s\n", err.Error())
		return
	}

	g.ServerIdML = int64(g.GameConfig.ServerId) << 32

	dispatch.InitData(g.GameConfig.MaxConnection)

	// file, err := os.Create(fmt.Sprintf("%s/server_%d.profile", g.GameConfig.LogPath, g.GameConfig.ServerId))
	// if err != nil {
	// 	fmt.Printf("create profile error: %s\n", err.Error())
	// 	return
	// }
	//pprof.StartCPUProfile(file)
	//defer pprof.StopCPUProfile()

	if err := log.SetFileName(fmt.Sprintf("%s/server_%d", g.GameConfig.LogPath, g.GameConfig.ServerId)); err != nil {
		fmt.Printf("create log file error: %s\n", err.Error())
		return
	}
	log.SetLevel(g.GameConfig.LogLevel)

	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("%v,%s", err, string(debug.Stack()))
		}
		log.Close()
	}()

	startT := time.Now()
	log.Info("server starting...")

	//加载配置表
	log.Info("load config datas...")
	g.LoadConfigs(g.GameConfig.ConfigPath)

	//加载敏感词
	log.Info("load filtertext...")
	g.LoadFilterTexts(g.GameConfig.ConfigPath)

	//加载随机昵称
	log.Info("load random names...")
	g.LoadRandomNames(g.GameConfig.ConfigPath)

	//数据库引擎
	log.Info("start database engine...")
	engine.InitEngine()

	data.OnLoadSystemData()
	actormgr.OnLoadMaxActorId()
	actormgr.OnLoadAllActorNames()

	dispatch.OnRun()

	service.OnConfigReloadFinish(true)
	service.OnGameStart()

	startGateWay()

	log.Info("============================================================")
	log.Infof("server started success, cost:%v", time.Since(startT))
	log.Info("============================================================")

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, os.Interrupt, os.Kill)
	select {
	case <-signalC:
	}

	tick := time.Now()
	log.Info("server closing...")

	//保存所有玩家数据
	waitClose := make(chan bool, 1)
	dispatch.PushSystemMsg(func() {
		defer func() {
			waitClose <- true
			g.CloseGame()
		}()
		actormgr.OnGameClose()
		service.OnGameClose()
		data.OnGameClose()
	})

	<-waitClose

	// select {
	// case <-waitClose:
	// case <-time.After(time.Second * 10):
	// 	log.Warn("close server failed...")
	// 	gconfig.CloseGame()
	// 	actormgr.OnGameClose()
	// 	service.OnGameClose()
	// 	data.OnGameClose()
	// }

	log.Infof("server closed success, cost:%v; running time:%v", time.Since(tick), time.Since(startT))
}

func onReLoadConfig(map[string]string) (int, string) {
	log.Info("=================reload config===================")
	//加载配置表
	log.Info("load config datas...")
	g.LoadConfigs(g.GameConfig.ConfigPath)

	//加载敏感词
	log.Info("load filtertext...")
	g.LoadFilterTexts(g.GameConfig.ConfigPath)

	log.Info("load random names...")
	g.LoadRandomNames(g.GameConfig.ConfigPath)

	service.OnConfigReloadFinish(false)

	log.Info("==============reload config success===============")

	return 0, "reload success"
}

func onSetMaxAccount(values map[string]string) (int, string) {
	count, _ := strconv.Atoi(values["count"])
	g.SetMaxCount(uint(count))

	return 0, fmt.Sprintf("max account %d", g.GetMaxCount())
}

func onSetRealAccount(values map[string]string) (int, string) {
	count, _ := strconv.Atoi(values["count"])
	g.SetRealCount(uint(count))

	return 0, fmt.Sprintf("real account %d", g.GetRealCount())
}
