package lordskill

import (
	"bytes"

	"github.com/sencydai/gameworld/dispatch"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/service/attr"
	"github.com/sencydai/gameworld/service/bag"
	t "github.com/sencydai/gameworld/typedefine"
)

func init() {
	service.RegActorLogin(onActorLogin)
	service.RegActorUpgrade(onActorUpgrade)

	dispatch.RegActorMsgHandle(proto.Lord, proto.LordCSkillStage, onSkillIndex)
	dispatch.RegActorMsgHandle(proto.Lord, proto.LordCSkillUpgrade, onSkillLevel)
	dispatch.RegActorMsgHandle(proto.Lord, proto.LordCSkillExchangePos, onExchangePos)
}

func onSendBaseInfo(actor *t.Actor) {
	skillData := actor.GetLordSkillData()
	writer := pack.AllocPack(proto.Lord, proto.LordSSkillInit, int16(len(skillData)))
	for pos, data := range skillData {
		pack.Write(writer, pos, data.Id, data.Index, data.Level)
	}
	actor.ReplyWriter(writer)
}

func onActorLogin(actor *t.Actor, offSec int) {
	onSendBaseInfo(actor)
}

//玩家升级
func onActorUpgrade(actor *t.Actor, oldLevel int) {
	skillData := actor.GetLordSkillData()
	skillCount := len(g.GLordSkillConfig)
	//所有技能已学
	if len(skillData) == skillCount {
		return
	}

	maxSkillId := 0
	for _, data := range skillData {
		if data.Id > maxSkillId {
			maxSkillId = data.Id
		}
	}

	pos := 0
	for i := maxSkillId + 1; i <= skillCount; i++ {
		conf := g.GLordSkillConfig[i][1]
		if conf.LordLevel > actor.Level {
			break
		}
		for i := pos + 1; i <= skillCount; i++ {
			if _, ok := skillData[i]; !ok {
				pos = i
				skillData[pos] = &t.ActorBaseLordSkillData{Id: conf.Id, Index: 1, Level: 1}
				if oldLevel != 0 {
					actor.Reply(proto.Lord, proto.LordSNewSkill, pos, conf.Id, 1, 1)
					attr.RefreshLordSkillPower(actor, true)
				}

				break
			}
		}
	}
}

//技能进阶
func onSkillIndex(actor *t.Actor, reader *bytes.Reader) {
	var (
		pos   int
		index int
	)
	pack.Read(reader, &pos, &index)
	skillData := actor.GetLordSkillData()
	skill, ok := skillData[pos]
	//技能未开启
	if !ok {
		return
	}
	if skill.Index != 1 || skill.Index == index {
		return
	}
	confs := g.GLordSkillConfig[skill.Id]
	if _, ok := confs[index]; !ok {
		return
	}
	conf := confs[1]
	//进阶需求技能等级不足
	if conf.Level > skill.Level {
		return
	}
	skill.Index = index
	actor.Reply(proto.Lord, proto.LordSSkillStage, pos, index)

	attr.RefreshLordSkillPower(actor, true)
}

//技能升级
func onSkillLevel(actor *t.Actor, reader *bytes.Reader) {
	var count int16
	pack.Read(reader, &count)

	skillData := actor.GetLordSkillData()
	updates := make(map[int]int)
	for i := int16(0); i < count; i++ {
		var (
			pos   int
			level int
		)
		pack.Read(reader, &pos, &level)
		//技能等级不能大于领主等级
		if level > actor.Level || level > len(g.GLordSkillCostConfig) {
			continue
		}
		skill, ok := skillData[pos]
		if !ok || skill.Level >= level {
			continue
		}

		cost := make(map[int]int)
		for i := skill.Level; i < level; i++ {
			conf := g.GLordSkillCostConfig[i]
			for _, item := range conf.Consume {
				cost[item.Id] += item.Count
			}
		}
		if !bag.DeductItems2(actor, cost, true, "skillupgrade") {
			break
		}

		skill.Level = level
		updates[pos] = level
	}

	writer := pack.AllocPack(proto.Lord, proto.LordSSkillUpgrade, int16(len(updates)))
	for pos, level := range updates {
		pack.Write(writer, pos, level)
	}
	actor.ReplyWriter(writer)

	if len(updates) > 0 {
		attr.RefreshLordSkillPower(actor, true)
	}
}

//更改技能位置
func onExchangePos(actor *t.Actor, reader *bytes.Reader) {
	var (
		pos1 int
		pos2 int
	)
	pack.Read(reader, &pos1, &pos2)
	if pos1 == pos2 || pos1 < 1 || pos2 < 2 || pos1 > len(g.GLordSkillCostConfig) || pos2 > len(g.GLordSkillCostConfig) {
		return
	}
	skillData := actor.GetLordSkillData()
	skill1, skill2 := skillData[pos1], skillData[pos2]
	if skill1 == nil && skill2 == nil {
		return
	}

	//两个位置都不为空
	if skill1 != nil && skill2 != nil {
		skillData[pos1], skillData[pos2] = skill2, skill1
	} else if skill1 == nil { //位置1为空
		skillData[pos1] = skill2
		delete(skillData, pos2)
	} else { //位置2为空
		skillData[pos2] = skill2
		delete(skillData, pos1)
	}

	onSendBaseInfo(actor)
}
