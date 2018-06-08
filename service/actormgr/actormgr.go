package actormgr

import (
	"bytes"
	"database/sql"
	"math"
	"time"

	"github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"
	"github.com/sencydai/gameworld/base"
	. "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/data"
	"github.com/sencydai/gameworld/dispatch"
	"github.com/sencydai/gameworld/engine"
	"github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/timer"
	. "github.com/sencydai/gameworld/typedefine"
	"github.com/sencydai/utils/log"
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
		maxActorId = gconfig.ServerIdML
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
		gconfig.UseRandomName(name)
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
}

func onAccountLogin(account *Account, reader *bytes.Reader) {
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
	if serverId != gconfig.GameConfig.ServerId {
		account.Close()
		return
	}

	dispatch.PushSystemAsynMsg(func(accountId int, pass string, gmLevel byte, err error) {
		if err != nil || pass != password {
			account.Close()
			return
		}
		if account.AccountId != 0 {
			account.Close()
			return
		}

		if data.GetAccount(accountId) != nil {
			account.Close()
			return
		}

		account.AccountId = accountId
		account.GmLevel = gmLevel
		account.Reply(pack.EncodeData(proto.System, proto.SystemSLogin, byte(0)))
		data.AppendAccount(account)

	}, engine.GetAccountInfo, accountName)
}

func onGetActorList(account *Account, reader *bytes.Reader) {
	if account.AccountId == 0 || account.Actor != nil {
		return
	}

	dispatch.PushSystemAsynMsg(func(actors []*AccountActor, err error) {
		if err != nil {
			account.Reply(pack.EncodeData(proto.System, proto.SystemSActorLists, account.AccountId, -1))
			return
		}
		writer := pack.AllocPack(proto.System, proto.SystemSActorLists, account.AccountId, len(actors))
		for _, actor := range actors {
			conf := gconfig.GLordConfig[actor.Camp][actor.Sex]
			pack.Write(writer, actor.ActorId, actor.ActorName, conf.Head, actor.Sex, actor.Level, actor.Camp, account.AccountId)
		}

		account.Reply(pack.EncodeWriter(writer))

	}, engine.GetAccountActors, account.AccountId)
}

func onGetRandName(account *Account, reader *bytes.Reader) {
	if account.AccountId == 0 {
		return
	}
	name := gconfig.GetRandomName()
	writer := pack.AllocPack(proto.System, proto.SystemSRandomName)
	if len(name) == 0 {
		pack.Write(writer, -9, 1, "")
	} else {
		pack.Write(writer, 0, 1, name)
	}
	account.Reply(pack.EncodeWriter(writer))
}

func onCreateActor(account *Account, reader *bytes.Reader) {
	if account.AccountId == 0 || account.Actor != nil {
		return
	}

	var (
		errCode int
		actorId int64
	)

	defer func() {
		account.Reply(pack.EncodeData(proto.System, proto.SystemSCreateActor, float64(actorId), errCode))
	}()

	var (
		name string
		camp int
		sex  int
		icon int
		pf   string
	)
	pack.Read(reader, &name, &camp, &sex, &icon, &pf)
	confs, ok := gconfig.GLordConfig[camp]
	if !ok {
		errCode = -11
		return
	}

	_, ok = confs[sex]
	if !ok {
		errCode = -8
		return
	}
	if IsActorNameExist(name) {
		errCode = -6
		return
	}
	rName := []rune(name)
	if len(rName) == 0 || len(rName) > MaxActorNameLen || gconfig.QueryName(name) {
		errCode = -12
		return
	}
	count, err := engine.GetActorCount(account.AccountId)
	if err != nil {
		log.Errorf("GetActorCount %d error: %s", account.AccountId, err.Error())
		errCode = -1
		return
	}
	if count > 0 {
		errCode = -15
		return
	}
	actorId = newActorId()
	nowT := time.Now()
	actor := &Actor{
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
		BaseData:   &ActorBaseData{},
		ExData:     &ActorExData{Pf: pf},
	}
	service.OnActorCreate(actor)
	if err = engine.InsertActor(actor); err != nil {
		log.Errorf("create actor error: %s", err.Error())
		errCode = -1
		return
	}
	AppendActorName(name, actorId)
	gconfig.UseRandomName(name)
	log.Infof("account(%d) create actor(%d) success", account.AccountId, actorId)
}

func onActorLogin(account *Account, reader *bytes.Reader) {
	if account.AccountId == 0 || account.Actor != nil {
		return
	}

	var (
		aId float64
		pf  string
	)
	pack.Read(reader, &aId, &pf)
	actor, err := engine.QueryActor(int64(aId))
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
	data.AddOnlineActor(actor)
	offTime := time.Duration(math.Max(float64(actor.LoginTime.Sub(actor.LogoutTime)), 0))

	service.OnActorBeforeLogin(actor, offTime)

	if !base.IsSameDay(time.Now(), base.Unix(exData.NewDay)) {
		service.OnActorNewDay(actor)
	}
	actor.Account = account
	service.OnActorLogin(actor, offTime)

	log.Infof("actor(%d) init success", actor.ActorId)
	var ch chan bool
	timer.Loop(actor, "actorSaveData", time.Minute*30, time.Minute*30, -1, engine.UpdateActor, ch)
}

func actorLogout(actor *Actor, flush chan bool) {
	timer.StopActorTimers(actor)
	service.OnActorLogout(actor)
	data.RemoveOnlineActor(actor, flush)
}

func OnAccountLogout(account *Account) {
	if account.Actor != nil {
		flush := make(chan bool, 1)
		actorLogout(account.Actor, flush)
		<-flush
	}
	if account.AccountId != 0 {
		data.RemoveAccount(account.AccountId)
	}
}

func OnGameClose() {
	chs := make(map[chan bool]bool)
	data.LoopActors(func(actor *Actor) bool {
		flush := make(chan bool, 1)
		chs[flush] = false
		actorLogout(actor, flush)
		return true
	})
	for ch := range chs {
		<-ch
	}
}
