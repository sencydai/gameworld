package lord

import (
	"bytes"
	"time"

	"github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"
	"github.com/sencydai/gameworld/base"
	c "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/data"
	"github.com/sencydai/gameworld/dispatch"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/service/actormgr"
	"github.com/sencydai/gameworld/service/attr"
	"github.com/sencydai/gameworld/service/cross"
	"github.com/sencydai/gameworld/timer"
	t "github.com/sencydai/gameworld/typedefine"
)

func init() {
	service.RegGameStart(onGameStart)
	service.RegActorCreate(onActorCreate)
	service.RegActorLogin(onActorLogin)
	service.RegActorBeforeLogin(onActorBeforeLogin)

	dispatch.RegActorMsgHandle(proto.Lord, proto.LordCChangeJob, onChangeJob)
	dispatch.RegActorMsgHandle(proto.Base, proto.BaseCUpdateClientData, onUpdateClientData)
	dispatch.RegActorMsgHandle(proto.Lord, proto.LordCLookupLord, onLookupLord)

	dispatch.RegCrossMsg(proto.CrossLookLordReq, onCrossLookLordReq)
	dispatch.RegCrossMsg(proto.CrossLookLordRes, onCrossLookLordRes)
}

func onGameStart() {
	t.NewRank(c.RankLevel, c.RankLevelCount)
	t.NewRank(c.RankPower, c.RankPowerCount)
}

func onActorCreate(actor *t.Actor) {
	jobData := actor.GetJobData()
	conf := g.GLordConfig[actor.Camp][actor.Sex]
	jobData.Jobs[conf.Job] = 1
	jobConf := g.GChangeJobConfig[conf.Job]
	jobData.Level = jobConf.Level
}

func onActorBeforeLogin(actor *t.Actor, offSec int) {
	//每分钟定时器
	times := offSec / 60
	if times > 0 {
		service.OnActorMinTimer(actor, times)
	}

	timer.Loop(actor, "mintimer", time.Minute, time.Minute, -1, service.OnActorMinTimer, 1)
}

func onSendBaseInfo(actor *t.Actor) {
	exData := actor.GetExData()
	jobData := actor.GetJobData()
	writer := pack.AllocPack(
		proto.Lord,
		proto.LordSBaseInfo,
		actor.Camp,
		actor.Sex,
		actor.Level,
		exData.Exp,
		jobData.Level,
		int16(len(jobData.Jobs)),
	)
	for job := range jobData.Jobs {
		pack.Write(writer, job)
	}
	pack.Write(writer, float64(actor.Power))
	actor.ReplyWriter(writer)
}

func onSyncTime(actor *t.Actor) {
	now := time.Now()
	nowS := now.Unix()
	open := t.GetSysOpenServerTime()
	openS := open.Unix()
	deltaOpen := base.GetDeltaDays(now, open)
	actor.Reply(proto.Base, proto.BaseSSyncTime, int(nowS), openS, deltaOpen)
}

func onSyncClientData(actor *t.Actor) {
	clientData := actor.GetClientData()
	writer := pack.AllocPack(proto.Base, proto.BaseSClientData, int16(len(clientData)))
	for k, v := range clientData {
		pack.Write(writer, k, v)
	}
	actor.ReplyWriter(writer)
}

func onActorLogin(actor *t.Actor, offSec int) {
	attr.RefreshAttr(actor)

	onSendBaseInfo(actor)
	onSyncTime(actor)
	onSyncClientData(actor)
}

//转职
func onChangeJob(actor *t.Actor, reader *bytes.Reader) {
	var jobId int
	pack.Read(reader, &jobId)

	conf, ok := g.GChangeJobConfig[jobId]
	if !ok {
		return
	}

	//铁匠铺等级

	jobData := actor.GetJobData()
	if conf.Level-jobData.Level != 1 {
		return
	}
	jobData.Level++
	jobData.Jobs[jobId] = 1

	writer := pack.AllocPack(proto.Lord, proto.LordSChangeJob, jobData.Level, int16(len(jobData.Jobs)))
	for id := range jobData.Jobs {
		pack.Write(writer, id)
	}
	actor.ReplyWriter(writer)
}

//更新前端数据
func onUpdateClientData(actor *t.Actor, reader *bytes.Reader) {
	var (
		key   int
		value string
	)
	pack.Read(reader, &key, &value)

	clientData := actor.GetClientData()
	clientData[key] = value
}

func packActorData(writer *bytes.Buffer, aid int64, aName string) {
	actor := data.GetActor(aid)
	if actor == nil {
		var ok bool
		aid, ok = actormgr.GetActorId(aName)
		if ok {
			actor = data.GetActor(aid)
		}
	}
	pack.Write(writer, float64(aid))
	if actor == nil {
		pack.Write(writer, "")
		return
	}
	pack.Write(writer,
		actor.ActorName,
		actor.GetLordModel(),
		actor.GetLordHead(),
		actor.GetLordFrame(),
		actor.Level,
		0,
		float64(actor.Power),
		int16(c.LEPMax),
	)

	equipData := actor.GetLordEquipData()
	for i := 1; i <= c.LEPMax; i++ {
		equip := equipData.Equips[i]
		pack.Write(writer, equip.Id)
	}

	heros := actor.GetFightHeros()
	pack.Write(writer, int16(len(heros)))
	for _, guid := range heros {
		hero := actor.GetHeroStaticData(guid)
		pack.Write(writer, hero.Pos, hero.Guid, hero.Id, hero.Level, hero.Stage)
	}
	pack.Write(writer, "")
}

//查看玩家
func onLookupLord(actor *t.Actor, reader *bytes.Reader) {
	var (
		lookType  int
		lookParam string
		serverId  int
		aid       float64
		aName     string
	)
	pack.Read(reader, &lookType, &lookParam, &serverId, &aid, &aName)

	//跨服查看
	if serverId != -1 && serverId != g.GameConfig.ServerId {
		cross.PublishSpecServerMsg(serverId, proto.CrossLookLordReq, actor.ActorId, lookType, lookParam, int64(aid), aName)
		return
	}

	writer := pack.AllocPack(proto.Lord, proto.LordSLookupLord, lookType, lookParam, g.GameConfig.ServerId)
	packActorData(writer, int64(aid), aName)
	actor.ReplyWriter(writer)
}

func onCrossLookLordReq(serverId int, reader *bytes.Reader) {
	var (
		actorId   int64
		lookType  int
		lookParam string
		aid       int64
		aName     string
	)
	pack.Read(reader, &actorId, &lookType, &lookParam, &aid, &aName)

	writer := pack.NewWriter(actorId, lookType, lookParam, g.GameConfig.ServerId)
	packActorData(writer, aid, aName)
	cross.PublishSpecServerMsg(serverId, proto.CrossLookLordRes, writer.Bytes())
}

func onCrossLookLordRes(serverId int, reader *bytes.Reader) {
	var actorId int64
	pack.Read(reader, &actorId)
	actor := data.GetOnlineActor(actorId)
	if actor == nil {
		return
	}

	data := make([]byte, reader.Len())
	reader.Read(data)

	actor.Reply(proto.Lord, proto.LordSLookupLord, data)
}
