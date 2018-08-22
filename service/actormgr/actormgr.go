package actormgr

import (
	"bytes"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/sencydai/gameworld/base"
	c "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/data"
	"github.com/sencydai/gameworld/dispatch"
	"github.com/sencydai/gameworld/engine"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/log"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/timer"
	t "github.com/sencydai/gameworld/typedefine"
)

var (
	maxActorId int64
	actorNames map[string]int64
)

func OnLoadMaxActorId() {
	var err error
	maxActorId, err = engine.GetMaxActorId()
	if err != nil {
		if err != sql.ErrNoRows {
			panic(err)
		}
	}

	if maxActorId == 0 {
		maxActorId = g.ServerIdML
	}
}

func newActorId() int64 {
	maxActorId++
	return maxActorId
}

func OnLoadAllActorNames() {
	var err error
	actorNames, err = engine.GetAllActorNames()
	if err != nil {
		panic(err)
	}

	for name := range actorNames {
		g.UseRandomName(name)
	}
}

func IsActorNameExist(name string) bool {
	_, ok := actorNames[name]
	return ok
}

func GetActorId(name string) (int64, bool) {
	actorId, ok := actorNames[name]
	return actorId, ok
}

func AppendActorName(name string, actorId int64) {
	actorNames[name] = actorId
}

func RemoveActorName(name string) {
	delete(actorNames, name)
}

func init() {
	dispatch.RegAccountMsgHandle(proto.SystemCLogin, onAccountLogin)
	dispatch.RegAccountMsgHandle(proto.SystemCActorList, onGetActorList)
	dispatch.RegAccountMsgHandle(proto.SystemCRandomName, onGetRandName)
	dispatch.RegAccountMsgHandle(proto.SystemCCreateActor, onCreateActor)
	dispatch.RegAccountMsgHandle(proto.SystemCLoginGame, onActorLogin)

	service.RegGameStart(onGameStart)
	service.RegGm("stat", onGetActorCount)
}

//账号登录
func onAccountLogin(account *t.Account, reader *bytes.Reader) {
	if account.AccountId != 0 {
		account.Close()
		return
	}
	var (
		serverId    int
		accountName string
		password    string
	)
	pack.Read(reader, &serverId, &accountName, &password)
	if serverId != g.GameConfig.ServerId {
		account.Close()
		return
	}
	dispatch.PushSystemAsynMsg(func(accountId int, pass string, gmlevel byte, err error) {
		if err != nil || pass != password {
			account.Close()
			return
		}

		//账号已登录
		if account.AccountId != 0 {
			account.Close()
			return
		}

		//账号已登陆
		if data.GetAccount(accountId) != nil {
			log.Errorf("acccount(%d) is connected", accountId)
			account.Close()
			return
		}
		account.AccountId = accountId
		account.GmLevel = gmlevel
		data.AppendAccount(account)

		account.Reply(pack.EncodeData(proto.System, proto.SystemSLogin, byte(0)))

	}, engine.GetAccountInfo, accountName)
}

func onGetActorList(account *t.Account, reader *bytes.Reader) {
	if account.AccountId == 0 || account.Actor != nil {
		return
	}

	dispatch.PushSystemAsynMsg(func(actors []*t.AccountActor, err error) {
		if err != nil {
			account.Reply(pack.EncodeData(proto.System, proto.SystemSActorLists, account.AccountId, -1))
			return
		}
		account.ActorCount = len(actors)
		writer := pack.AllocPack(proto.System, proto.SystemSActorLists, account.AccountId, len(actors))
		for _, actor := range actors {
			conf := g.GLordConfig[actor.Camp][actor.Sex]
			pack.Write(writer, actor.ActorId, actor.ActorName, conf.Head, actor.Sex, actor.Level, actor.Camp, account.AccountId)
		}
		account.Reply(pack.EncodeWriter(writer))
	}, engine.GetAccountActors, account.AccountId)
}

func onGetRandName(account *t.Account, reader *bytes.Reader) {
	if account.AccountId == 0 {
		return
	}
	//name := g.GetRandomName()
	name := "unknown"
	writer := pack.AllocPack(proto.System, proto.SystemSRandomName)
	if len(name) == 0 {
		pack.Write(writer, -9, 1, "")
	} else {
		pack.Write(writer, 0, 1, name)
	}
	account.Reply(pack.EncodeWriter(writer))
}

func onCreateActor(account *t.Account, reader *bytes.Reader) {
	if account.AccountId == 0 || account.Actor != nil {
		return
	}

	var errCode int
	var actorId int64

	defer func() {
		account.Reply(pack.EncodeData(proto.System, proto.SystemSCreateActor, float64(actorId), errCode))
	}()

	if account.ActorCount != 0 {
		errCode = -15
		return
	}

	var (
		name string
		camp int
		sex  int
		icon int
		pf   string
	)

	pack.Read(reader, &name, &camp, &sex, &icon, &pf)
	confs, ok := g.GLordConfig[camp]
	//阵营错误
	if !ok {
		errCode = -11
		return
	}
	//性别错误
	_, ok = confs[sex]
	if !ok {
		errCode = -8
		return
	}
	name = g.GetRandomName()
	//角色名已存在
	if IsActorNameExist(name) {
		errCode = -6
		return
	}
	rName := []rune(name)
	//名称不合法
	if len(rName) == 0 || len(rName) > c.MaxActorNameLen || g.QueryName(name) {
		errCode = -12
		return
	}

	actorId = newActorId()
	nowT := time.Now()
	actor := &t.Actor{
		ActorId:    actorId,
		ActorName:  name,
		AccountId:  account.AccountId,
		Camp:       camp,
		Sex:        sex,
		Level:      1,
		Power:      1,
		CreateTime: nowT,
		LoginTime:  nowT,
		LogoutTime: nowT,
		BaseData:   &t.ActorBaseData{},
		ExData:     &t.ActorExData{Pf: pf},
	}
	service.OnActorCreate(actor)
	service.OnActorUpgrade(actor, 0)

	if err := engine.InsertActor(actor); err != nil {
		log.Errorf("create actor error: %s", err.Error())
		errCode = -1
		return
	}

	account.ActorCount++
	AppendActorName(name, actorId)
	g.UseRandomName(name)

	log.Infof("account(%d) create actor(%d) success", account.AccountId, actorId)
}

func onActorLogin(account *t.Account, reader *bytes.Reader) {
	if account.AccountId == 0 || account.Actor != nil {
		return
	}
	tick := time.Now()
	var (
		aId float64
		pf  string
	)
	pack.Read(reader, &aId, &pf)

	dispatch.PushSystemAsynMsg(func(actor *t.Actor, err error) {
		if account.AccountId == 0 || account.Actor != nil {
			return
		}
		if err != nil {
			log.Errorf("login actor error: %s", err.Error())
			account.Reply(pack.EncodeData(proto.System, proto.SystemSLoginGame, -1))
			return
		}

		if actor.AccountId != account.AccountId {
			account.Reply(pack.EncodeData(proto.System, proto.SystemSLoginGame, -2))
			return
		}
		account.Reply(pack.EncodeData(proto.System, proto.SystemSLoginGame, 0))

		exData := actor.GetExData()
		exData.Pf = pf
		account.Actor = actor

		account.Reply(pack.EncodeData(proto.Lord, proto.LordSCreateActor,
			float64(actor.ActorId), float64(actor.ActorId),
			g.GameConfig.ServerId, actor.ActorName))

		data.AddOnlineActor(actor)

		offSec := int(time.Duration(math.Max(float64(actor.LoginTime.Sub(actor.LogoutTime)), 0)) / time.Second)

		service.OnActorBeforeLogin(actor, offSec)

		if !base.IsSameDay(time.Now(), base.Unix(exData.NewDay)) {
			service.OnActorNewDay(actor)
		}

		actor.Account = account
		service.OnActorLogin(actor, offSec)

		log.Infof("actor(%d) init success, cost %v", actor.ActorId, time.Since(tick))
		timer.Loop(actor, "actorSaveData", time.Minute*30, time.Minute*30, -1, engine.UpdateActor)
	}, engine.QueryActor, int64(aId))
}

func actorLogout(actor *t.Actor) {
	timer.StopActorTimers(actor)
	service.OnActorLogout(actor)
	data.RemoveOnlineActor(actor)
}

func OnAccountLogout(account *t.Account) {
	if account.Actor != nil {
		actorLogout(account.Actor)
		account.Actor = nil
	}
	if account.AccountId != 0 {
		data.RemoveAccount(account.AccountId)
	}
}

func onGameStart() {
	timer.Loop(nil, "flushactors", time.Minute*5, time.Minute*5, -1, func() {
		go func() {
			engine.FlushActorBuffers()
		}()
	})
}

func OnGameClose() {
	tick := time.Now()
	data.LoopActors(func(actor *t.Actor) bool {
		service.OnActorLogout(actor)
		actor.LogoutTime = time.Now()
		engine.UpdateActor(actor)
		return true
	})
	engine.FlushActorBuffers()
	log.Infof("save actors data: %v", time.Since(tick))
}

func onGetActorCount(map[string]string) (int, string) {
	return 0, fmt.Sprintf(
		"max: %d, real: %d, account: %d, online: %d, offline: %d, engineBuff: %d",
		g.GetMaxCount(),
		g.GetRealCount(),
		data.GetAccountCount(),
		data.GetOnlineCount(),
		data.GetCacheCount(),
		engine.GetCacheCount(),
	)
}
