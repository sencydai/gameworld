package attr

import (
	c "github.com/sencydai/gameworld/constdefine"
	g "github.com/sencydai/gameworld/gconfig"
	t "github.com/sencydai/gameworld/typedefine"

	"github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"

	_ "github.com/sencydai/gameworld/log"
)

//CalcLordPower 领主战力
func CalcLordPower(actor *t.Actor) {
	if !actor.IsOnline() {
		return
	}

	//领主属性未计算
	lordAttr := actor.GetLordAttr()
	if len(lordAttr.Total) == 0 {
		return
	}

	power := lordAttr.SkillPower
	for _, guid := range actor.GetFightHeros() {
		heroAttr := actor.GetHeroAttr(guid)
		//英雄属性未计算
		if len(heroAttr.Total) == 0 {
			return
		}
		hero := actor.GetHeroStaticData(guid)
		power += hero.Power
	}

	lordTotal := lordAttr.Total
	power += int(lordTotal[c.AttrLordDamage]+lordTotal[c.AttrLordDamageSub]) * 4

	oldPower := actor.Power
	actor.Power = power

	writer := pack.AllocPack(proto.Lord, proto.LordSRefreshPower, float64(power), int16(len(lordTotal)))
	for t, v := range lordTotal {
		pack.Write(writer, t, int(v))
	}
	actor.ReplyWriter(writer)

	if oldPower != actor.Power {
		rankData := t.GetRank(c.RankPower)
		rankData.Insert(actor.ActorId, int64(actor.Power))
	}

	//log.Infof("CalcLordPower: actor:%d, power:%d", actor.ActorId, power)
}

//更新领主属性
func refreshLordAttr(actor *t.Actor) {
	//技能战力
	RefreshLordSkillPower(actor, false)

	//等级属性
	RefreshLordLevelAttr(actor, false)

	//装备属性
	refreshLordEquipsAttr(actor)

	//领主天赋
	refreshLordTalentsAttr(actor)

	//总属性
	refreshLordTotalAttr(actor)
}

//更新领主总属性
func refreshLordTotalAttr(actor *t.Actor) {
	lordAttr := actor.GetLordAttr()
	attrs := lordAttr.Total
	for t := range attrs {
		delete(attrs, t)
	}

	//等级
	for t, v := range lordAttr.Level {
		attrs[t] += v
	}

	//装备
	for _, v := range lordAttr.Equip {
		for t, vv := range v {
			attrs[t] += vv
		}
	}

	//天赋
	for _, v := range lordAttr.Talent {
		for t, vv := range v {
			attrs[t] += vv
		}
	}

	//更新英雄的领主属性
	heros := actor.GetFightHeros()
	for _, guid := range heros {
		heroAttr := actor.GetHeroAttr(guid)
		//英雄属性没有初始化
		if len(heroAttr.Total) == 0 {
			return
		}
	}

	for _, guid := range heros {
		refreshHeroTotalAttr(actor, actor.GetHeroStaticData(guid))
	}
}

//RefreshLordSkillPower 领主技能战力
func RefreshLordSkillPower(actor *t.Actor, refreshTotal bool) {
	lordAttr := actor.GetLordAttr()
	lordAttr.SkillPower = 0
	for _, skill := range actor.GetLordSkillData() {
		conf := g.GLordSkillLevelConfig[skill.Id][skill.Index][skill.Level]
		lordAttr.SkillPower += conf.Fight
	}

	if refreshTotal {
		CalcLordPower(actor)
	}
}

//RefreshLordLevelAttr 领主等级属性
func RefreshLordLevelAttr(actor *t.Actor, refreshTotal bool) {
	lordAttr := actor.GetLordAttr()
	attrs := lordAttr.Level
	for t := range attrs {
		delete(attrs, t)
	}

	conf := g.GLordLevelConfig[actor.Level]
	for _, attr := range conf.Attrs {
		attrs[attr.Type] = float64(attr.Value)
	}

	if refreshTotal {
		refreshLordTotalAttr(actor)
	}
}

//refreshLordEquipsAttr 领主装备
func refreshLordEquipsAttr(actor *t.Actor) {
	lordAttr := actor.GetLordAttr()
	attrs := lordAttr.Equip
	for t := range attrs {
		delete(attrs, t)
	}

	equipData := actor.GetLordEquipData()
	for pos, equip := range equipData.Equips {
		posAttrs := make(map[int]float64)
		attrs[pos] = posAttrs
		//装备属性
		if equip.Id > 0 {
			equipConf := g.GLordEquipConfig[equip.Id]
			for _, attr := range equipConf.Attrs {
				posAttrs[attr.Type] = float64(attr.Value)
			}
		}
		//强化属性
		strengConf := g.GLordEquipStrengConfig[equip.Stage][equip.Level]
		for _, attr := range strengConf.Attrs[pos] {
			posAttrs[attr.Type] += float64(attr.Value)
		}
	}
}

//RefreshLordEquipAttr 更新装备属性
func RefreshLordEquipAttr(actor *t.Actor, pos int) {
	lordAttr := actor.GetLordAttr()
	attrs := lordAttr.Equip
	delete(attrs, pos)

	equipData := actor.GetLordEquipData()
	if equip, ok := equipData.Equips[pos]; ok {
		posAttrs := make(map[int]float64)
		attrs[pos] = posAttrs
		//装备属性
		if equip.Id > 0 {
			equipConf := g.GLordEquipConfig[equip.Id]
			for _, attr := range equipConf.Attrs {
				posAttrs[attr.Type] = float64(attr.Value)
			}
		}
		//强化属性
		strengConf := g.GLordEquipStrengConfig[equip.Stage][equip.Level]
		for _, attr := range strengConf.Attrs[pos] {
			posAttrs[attr.Type] += float64(attr.Value)
		}
	}

	refreshLordTotalAttr(actor)
}

//refreshLordTalentsAttr 领主天赋
func refreshLordTalentsAttr(actor *t.Actor) {
	lordAttr := actor.GetLordAttr()
	attrs := lordAttr.Talent
	for t := range attrs {
		delete(attrs, t)
	}

	talentData := actor.GetLordTalentData()
	for id, level := range talentData.Learn {
		levelConf := g.GTalentLevelConfig[id][level]
		skillConf := g.GSkillConfig[levelConf.SkillId]
		if len(skillConf.Attr) > 0 {
			attr := make(map[int]float64)
			for _, v := range skillConf.Attr {
				attr[v.Type] = float64(v.Value)
			}
			attrs[id] = attr
		}
	}
}

//refreshLordTalentAttr 领主天赋
func refreshLordTalentAttr(actor *t.Actor, id int) {
	lordAttr := actor.GetLordAttr()
	attrs := lordAttr.Talent
	delete(attrs, id)

	talentData := actor.GetLordTalentData()
	if level, ok := talentData.Learn[id]; ok {
		levelConf := g.GTalentLevelConfig[id][level]
		skillConf := g.GSkillConfig[levelConf.SkillId]
		if len(skillConf.Attr) > 0 {
			attr := make(map[int]float64)
			for _, v := range skillConf.Attr {
				attr[v.Type] = float64(v.Value)
			}
			attrs[id] = attr
		}
	}

	refreshLordTotalAttr(actor)
}
