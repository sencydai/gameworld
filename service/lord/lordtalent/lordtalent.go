package lordtalent

import (
	"bytes"

	c "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/dispatch"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/service/bag"
	t "github.com/sencydai/gameworld/typedefine"
)

func init() {
	service.RegActorLogin(onActorLogin)

	dispatch.RegActorMsgHandle(proto.Lord, proto.LordCTalentLearn, onLearnTalent)
	dispatch.RegActorMsgHandle(proto.Lord, proto.LordCTalentUpgrade, onUpgradeTalent)
}

func onActorLogin(actor *t.Actor, offSec int) {
	talentData := actor.GetLordTalentData()
	writer := pack.AllocPack(proto.Lord, proto.LordSTalentInit,
		talentData.Count, int16(len(talentData.Learn)))
	for id, level := range talentData.Learn {
		pack.Write(writer, id, level)
	}
	actor.ReplyWriter(writer)
}

//学习天赋
func onLearnTalent(actor *t.Actor, reader *bytes.Reader) {
	var talentId int
	pack.Read(reader, &talentId)

	conf, ok := g.GTalentConfig[talentId]
	if !ok {
		return
	}

	talentData := actor.GetLordTalentData()
	if _, ok := talentData.Learn[talentId]; ok {
		return
	}

	//需求职业
	if conf.Job > 0 {
		jobData := actor.GetJobData()
		if _, ok := jobData.Jobs[conf.Job]; !ok {
			return
		}
	}

	//是否开启
	if conf.TotalCost > talentData.Count {
		return
	}

	//前置天赋
	if conf.PerTalent > 0 {
		level, ok := talentData.Learn[conf.PerTalent]
		//未学习
		if !ok {
			return
		}
		levelConfs := g.GTalentLevelConfig[conf.PerTalent]
		//未满级
		if level < len(levelConfs) {
			return
		}
	}

	//学习消耗点数
	if conf.Cost > 0 {
		if ok, _ := bag.DeductItem(actor, c.CTScienceP, conf.Cost, true, "learnTalent"); !ok {
			return
		}
		talentData.Count += conf.Cost
	}

	talentData.Learn[talentId] = 1

	writer := pack.AllocPack(proto.Lord, proto.LordSTalentLearn, talentId, 1, talentData.Count)
	actor.ReplyWriter(writer)
}

//升级天赋
func onUpgradeTalent(actor *t.Actor, reader *bytes.Reader) {
	var talentId int
	pack.Read(reader, &talentId)

	talentData := actor.GetLordTalentData()
	level, ok := talentData.Learn[talentId]
	if !ok {
		return
	}

	levelConfs := g.GTalentLevelConfig[talentId]
	//已满级
	if level >= len(levelConfs) {
		return
	}

	conf := levelConfs[level]
	if ok, _ := bag.DeductItem(actor, c.CTScienceP, conf.Cost, true, "upgradeTalent"); !ok {
		return
	}
	talentData.Count += conf.Cost
	talentData.Learn[talentId]++

	writer := pack.AllocPack(proto.Lord, proto.LordSTalentUpgrade,
		talentId, talentData.Learn[talentId], talentData.Count)
	actor.ReplyWriter(writer)
}
