package attr

import (
	"math"

	c "github.com/sencydai/gameworld/constdefine"
	g "github.com/sencydai/gameworld/gconfig"
	t "github.com/sencydai/gameworld/typedefine"

	"github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"
	"github.com/sencydai/gameworld/service"

	_ "github.com/sencydai/gameworld/log"
)

func init() {
	service.RegConfigLoadFinish(onConfigLoadFinish)
}

var (
	heroBaseLevelExp       float64
	heroBaseLevelMult      float64
	heroBaseScaleRatio     map[int]float64
	heroBaseAttrWorthRatio map[int]float64
)

func onConfigLoadFinish() {
	baseConf := g.GHeroBaseConfig
	heroBaseLevelExp = float64(baseConf.LevelExp) / 10000
	heroBaseLevelMult = float64(baseConf.LevelMult) / 10000

	heroBaseScaleRatio = make(map[int]float64)
	for _, attr := range baseConf.BaseScaleRatio {
		heroBaseScaleRatio[attr.Type] = float64(attr.Value) / 10000
	}

	heroBaseAttrWorthRatio = make(map[int]float64)
	for _, attr := range baseConf.AttrWorthRatio {
		heroBaseAttrWorthRatio[attr.Type] = float64(attr.Value) / 10000
	}
}

//CalcHeroPower 计算英雄战力
func CalcHeroPower(heroId int, attrs map[int]float64) int {
	//英雄战力=【（基础攻击+额外攻击+最小攻击*0.5+最大攻击*0.5）*
	//	{1+暴击率%/2*暴击倍率%+命中率%/2+伤害加深%/2+吸血率%/2+速度/2400} +
	//	（基础护甲+额外护甲+基础生命*0.08+额外生命*0.08）*{1+闪避率%/2+抗暴率%/2+伤害减免%/2+反伤率%/4}】* 英雄职业系数
	attack := attrs[c.AttrAttackBase] + attrs[c.AttrExAttack] + (attrs[c.AttrMinAttackBase]+attrs[c.AttrMaxAttackBase])*0.5
	crit := ((attrs[c.AttrCrit]*attrs[c.AttrCritDamage])/10000+attrs[c.AttrHit]+attrs[c.AttrDamageAddPct]+attrs[c.AttrSuckBlood])/20000 +
		1 + (attrs[c.AttrSpeedBase]+attrs[c.AttrExSpeed])/2400

	defense := attrs[c.AttrDefenseBase] + attrs[c.AttrExDefense] + (attrs[c.AttrHpBase]+attrs[c.AttrExHp])*0.08
	dodge := (attrs[c.AttrDodge]+attrs[c.AttrCritDefense]+attrs[c.AttrDamageSubPct]+attrs[c.AttrThorns]*0.5)/20000 + 1

	conf := g.GHeroConfig[heroId]
	return int((attack*crit + defense*dodge) * float64(conf.PowerRatio) / 10000)
}

//refreshHeroTotalAttr 总属性
func refreshHeroTotalAttr(actor *t.Actor, hero *t.HeroStaticData) {
	//领主属性加成
	refreshHeroLordAttr(actor, hero)

	heroAttr := actor.GetHeroAttr(hero.Guid)
	attrs := heroAttr.Total
	for t := range attrs {
		delete(attrs, t)
	}

	//领主
	for t, v := range heroAttr.Lord {
		attrs[t] = v
	}
	//基础
	for t, v := range heroAttr.Base {
		attrs[t] += v
	}
	//等级
	for t, v := range heroAttr.Level {
		attrs[t] += v
	}
	//等阶
	for t, v := range heroAttr.Stage {
		attrs[t] += v
	}

	CalcEntityFightAttr(attrs)

	if !actor.IsOnline() {
		return
	}

	//power := hero.Power
	hero.Power = CalcHeroPower(hero.Id, attrs)

	writer := pack.AllocPack(proto.Hero, proto.HeroSRefreshPower, hero.Guid, float64(hero.Power), int16(len(attrs)))
	for t, v := range attrs {
		pack.Write(writer, t, int(v))
	}
	actor.ReplyWriter(writer)

	//log.Infof("refreshHeroTotalAttr: actor:%d, hero:%d, power:%d, attrs:%v", actor.ActorId, hero.Guid, hero.Power, attrs)

	CalcLordPower(actor)
}

//RefreshHeroAttr 更新英雄属性
func RefreshHeroAttr(actor *t.Actor, hero *t.HeroStaticData) {
	//初始属性
	refreshHeroBaseAttr(actor, hero)

	//等级属性
	RefreshHeroLevelAttr(actor, hero, false)

	refreshHeroTotalAttr(actor, hero)
}

func refreshHeroLordAttr(actor *t.Actor, hero *t.HeroStaticData) {
	heroAttr := actor.GetHeroAttr(hero.Guid)
	attrs := heroAttr.Lord
	for t := range attrs {
		delete(attrs, t)
	}

	lordAttr := actor.GetLordAttr()
	if len(lordAttr.Total) == 0 {
		return
	}
	lordTotal := lordAttr.Total
	//攻击指挥
	attrs[c.AttrExAttack] += float64(int(lordTotal[c.AttrAttackCom] * (1 + lordTotal[c.AttrAttackComPct]/10000)))
	//防御指挥
	attrs[c.AttrExDefense] += float64(int(lordTotal[c.AttrDefenseCom] * (1 + lordTotal[c.AttrDefenseComPct]/10000)))

	//特性
	heroConf := g.GHeroConfig[hero.Id]
	for _, feasure := range heroConf.Feature {
		switch feasure {
		case c.FeasureNear:
			attrs[c.AttrDamageAddPct] += lordTotal[c.AttrNearDamageAdd]
			attrs[c.AttrDamageSubPct] += lordTotal[c.AttrNearDamageSub]
		case c.FeasureLong:
			attrs[c.AttrDamageAddPct] += lordTotal[c.AttrLongDamageAdd]
			attrs[c.AttrDamageSubPct] += lordTotal[c.AttrLongDamageSub]
		case c.FeasurePhys:
			attrs[c.AttrDamageAddPct] += lordTotal[c.AttrPhysDamageAdd]
			attrs[c.AttrDamageSubPct] += lordTotal[c.AttrPhysDamageSub]
		case c.FeasureMagic:
			attrs[c.AttrDamageAddPct] += lordTotal[c.AttrMagicDamageAdd]
			attrs[c.AttrDamageSubPct] += lordTotal[c.AttrMagicDamageSub]
		}
	}
}

//refreshHeroBaseAttr 基础属性
func refreshHeroBaseAttr(actor *t.Actor, hero *t.HeroStaticData) {
	heroAttr := actor.GetHeroAttr(hero.Guid)
	attrs := heroAttr.Base
	for t := range attrs {
		delete(attrs, t)
	}

	heroConf := g.GHeroConfig[hero.Id]
	rawConf := g.GHeroRawConfig[heroConf.RawHero]
	for _, v := range rawConf.Attrs {
		attrs[v.Type] = float64(v.Value)
	}
}

//RefreshHeroLevelAttr 等级属性
func RefreshHeroLevelAttr(actor *t.Actor, hero *t.HeroStaticData, refreshTotal bool) {
	heroAttr := actor.GetHeroAttr(hero.Guid)
	attrs := heroAttr.Level
	for t := range attrs {
		delete(attrs, t)
	}

	heroConf := g.GHeroConfig[hero.Id]
	rawConf := g.GHeroRawConfig[heroConf.RawHero]
	rarityConf := g.GHeroRarityConfig[rawConf.Rarity]
	level := float64(hero.Level)
	levelExp := math.Pow(level, heroBaseLevelExp)
	levelMult := level * heroBaseLevelMult
	constV := (levelExp + levelMult) * float64(rarityConf.LevelRatio) / 10000
	for _, attr := range rawConf.AttrRatio {
		attrs[attr.Type] = float64(int(constV * heroBaseScaleRatio[attr.Type] * float64(attr.Value) / 10000 / heroBaseAttrWorthRatio[attr.Type]))
	}

	if refreshTotal {
		refreshHeroTotalAttr(actor, hero)
	}
}
