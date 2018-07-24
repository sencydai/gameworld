package fight

import (
	"fmt"
	"time"

	"github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"
	c "github.com/sencydai/gameworld/constdefine"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/service/attr"
	"github.com/sencydai/gameworld/timer"
	t "github.com/sencydai/gameworld/typedefine"
)

type actorFightClear func(actor *t.Actor, fightData *t.FightData)
type actorFightAward func(actor *t.Actor, fightData *t.FightData)

var (
	heroTemplates     map[int]*t.FightHeroTemplate //英雄模板
	monsterTemplates  map[int]*t.FightHeroTemplate //怪物模板
	guardModels       map[int][]int                //亲卫模型
	raceRatios        map[int]float64              //种族系数
	rawHeroRaceRatios map[int]float64              //原始英雄种族系数
	buffGainTypes     map[int][]*t.BuffConfig
	buffTriggers      map[int]bool
	skillRepeatCounts map[int]int

	maxLordSkillCount int
	fightIndex        int64

	actorFightClearHandles = make(map[int]actorFightClear)
	actorFightAwardHandles = make(map[int]actorFightAward)
)

func RegFightClear(fightType int, handle func(*t.Actor, *t.FightData)) {
	actorFightClearHandles[fightType] = handle
}

func RegFightAward(fightType int, handle func(*t.Actor, *t.FightData)) {
	actorFightAwardHandles[fightType] = handle
}

func newFightIndex() int64 {
	fightIndex++
	return fightIndex
}

func init() {
	service.RegConfigLoadFinish(onConfigLoadFinish)
}

func onConfigLoadFinish() {
	heroTemplates = make(map[int]*t.FightHeroTemplate)
	monsterTemplates = make(map[int]*t.FightHeroTemplate)
	guardModels = make(map[int][]int)
	rawHeroRaceRatios = make(map[int]float64)

	raceRatios = make(map[int]float64)
	for _, conf := range g.GRaceConfig {
		raceRatios[conf.Race] = float64(conf.Ratio)/10000 + 1
	}

	gainTypes := make(map[int][]*t.BuffConfig)
	triggers := make(map[int]bool)
	for _, conf := range g.GBuffConfig {
		if _, ok := gainTypes[conf.GainType]; !ok {
			gainTypes[conf.GainType] = make([]*t.BuffConfig, 0)
		}
		gainTypes[conf.GainType] = append(gainTypes[conf.GainType], conf)

		triggers[conf.Effect.Trigger] = true
	}
	buffGainTypes = gainTypes
	buffTriggers = triggers

	repeat := make(map[int]int)
	for _, conf := range g.GSkillConfig {
		if conf.RepeatCount > 0 {
			repeat[conf.Skill] = conf.RepeatCount
		}
	}
	skillRepeatCounts = repeat

	maxLordSkillCount = len(g.GLordSkillConfig)
}

func getGuardModel(model int) []int {
	models, ok := guardModels[model]
	if ok {
		return models
	}

	modelConf := g.GGuardModelConfig[model]
	models = []int{modelConf.Model, modelConf.Model}
	guardModels[model] = models
	return models
}

func getRawHeroRaceRatio(rawId int) float64 {
	ratio, ok := rawHeroRaceRatios[rawId]
	if ok {
		return ratio
	}

	rawConf := g.GHeroRawConfig[rawId]
	ratio = raceRatios[rawConf.Race]
	rawHeroRaceRatios[rawId] = ratio
	return ratio
}

func parseHeroSkill(skills map[int]int) map[int][]int {
	var items map[int][]int
	if len(skills) > 0 {
		items = make(map[int][]int)
		for i := 1; i <= len(skills); i++ {
			skillId := skills[i]
			skillConf := g.GSkillConfig[skillId]
			if skillConf.Trigger > 0 {
				if _, ok := items[skillConf.Trigger]; !ok {
					items[skillConf.Trigger] = make([]int, 0)
				}
				trigger := items[skillConf.Trigger]
				items[skillConf.Trigger] = append(trigger, skillId)
			}
		}
	}

	return items
}

func parseHeroFeature(features map[int]int) map[int]bool {
	items := make(map[int]bool)
	for _, feature := range features {
		items[feature] = true
	}

	return items
}

func getHeroTemplate(heroId int) *t.FightHeroTemplate {
	hero, ok := heroTemplates[heroId]
	if ok {
		return hero
	}

	heroConf := g.GHeroConfig[heroId]

	hero = &t.FightHeroTemplate{
		Model:     heroConf.Model,
		CommSkill: heroConf.CommSkill,
		Skills:    parseHeroSkill(heroConf.Skills),
		RaceRatio: getRawHeroRaceRatio(heroConf.RawHero),
		Feature:   parseHeroFeature(heroConf.Feature),
	}

	heroTemplates[heroId] = hero

	return hero
}

func getMonsterTemplate(monsterId int) *t.FightHeroTemplate {
	hero, ok := monsterTemplates[monsterId]
	if ok {
		return hero
	}

	monsterConf := g.GMonsterConfig[monsterId]
	if monsterConf.Hero > 0 {
		hero = getHeroTemplate(monsterConf.Hero)
		monsterTemplates[monsterId] = hero
		return hero
	}

	hero = &t.FightHeroTemplate{
		Model:     monsterConf.Model,
		CommSkill: monsterConf.CommSkill,
		Skills:    parseHeroSkill(monsterConf.Skills),
		RaceRatio: raceRatios[monsterConf.Race],
		Feature:   parseHeroFeature(monsterConf.Feature),
	}

	monsterTemplates[monsterId] = hero

	return hero
}

func NewFightData(actor *t.Actor, fightType int, cbArgs []interface{}) *t.FightData {
	if actor.GetFightData() != nil {
		return nil
	}

	fightData := &t.FightData{
		Guid:        newFightIndex(),
		Type:        fightType,
		CbArgs:      cbArgs,
		Data:        make([]*t.FightLord, 2),
		Order:       []int{0, 1},
		Entities:    make(map[int]*t.FightEntity),
		RawEntities: make(map[int]*t.FightEntity),
		Logs:        make([]*t.FightLog, 0),
		StartTime:   time.Now(),
	}

	for i := 0; i < 2; i++ {
		fightData.Data[i] = &t.FightLord{
			Pos:   i * posRate,
			Heros: make(map[int]*t.FightHeroTemplate),
		}
	}

	dynamicData := actor.GetDynamicData()
	dynamicData.FightData = fightData
	return fightData
}

func Fighting(actor *t.Actor, fightData *t.FightData, packArgs []interface{}) {
	timer.NextGo(actor, fmt.Sprintf("Fighting_%d", fightData.Type), func(actor *t.Actor) {
		writer := pack.AllocPack(proto.Fight, proto.FightSInit, fightData.Type, float64(fightData.Guid))
		pack.Write(writer, int16(len(fightData.Data)+len(fightData.Entities)))
		for _, lord := range fightData.Data {
			pack.Write(writer,
				"",
				lord.Name,
				lord.Pos,
				lord.Model,
				0,
				0,
			)
			pack.Write(writer, int8(len(lord.Gmodel)), lord.Gmodel[0], lord.Gmodel[1])
			pack.Write(writer, int8(len(lord.Equips)))
			for pos, id := range lord.Equips {
				pack.Write(writer, pos, id)
			}
		}

		for _, entity := range fightData.Entities {
			heroTemp := fightData.Data[entity.LordIndex].Heros[entity.HeroPos]
			pack.Write(writer, "", "", entity.Pos, heroTemp.Model, int(entity.RawAttrs[c.AttrHp]), int(entity.Attrs[c.AttrHp]))
		}

		pack.Write(writer, packArgs...)

		actor.ReplyWriter(writer)

		startFighting(actor, fightData)
	})
}

//NewPvE pve
func NewPvE(actor *t.Actor, fightType int,
	monsterLord t.MonsterLord, monsterHeros map[int]t.MonsterHero,
	monsterName string, monsterLevel int,
	packArgs []interface{}, cbArgs ...interface{}) *t.FightData {
	fightData := NewFightData(actor, fightType, cbArgs)
	if fightData == nil {
		return nil
	}

	InitActorFightData(fightData, fightData.Data[0], actor, 0)
	InitMonsterFightData(fightData, fightData.Data[1], monsterLord, monsterHeros, monsterName, monsterLevel)

	Fighting(actor, fightData, packArgs)

	//log.Infof("NewPvE: actor:%d,type:%d,guid:%d", actor.ActorId, fightType, fightData.Guid)
	return fightData
}

//InitActorFightData 玩家战斗数据初始化
func InitActorFightData(data *t.FightData, lord *t.FightLord, actor *t.Actor, maxHeroPos int) {
	lordAttr := actor.GetLordAttr()
	if len(lordAttr.Total) == 0 {
		attr.RefreshAttr(actor)
	}

	lord.Model = actor.GetLordModel()
	lord.Name = actor.ActorName
	guard := actor.GetGuardData()
	lord.Gmodel = getGuardModel(guard.Model)
	lord.Power = actor.Power

	equipData := actor.GetLordEquipData()
	if len(equipData.Equips) > 0 {
		lord.Equips = make(map[int]int)
		for pos, equip := range equipData.Equips {
			if equip.Id > 0 {
				lord.Equips[pos] = equip.Id
			}
		}
	}

	talentData := actor.GetLordTalentData()
	if len(talentData.Learn) > 0 {
		lord.PassSkills = make([]int, len(talentData.Learn))
		index := 0
		for id, level := range talentData.Learn {
			levelConf := g.GTalentLevelConfig[id][level]
			lord.PassSkills[index] = levelConf.SkillId
			index++
		}
	}

	skillData := actor.GetLordSkillData()
	if len(skillData) > 0 {
		lord.ActiveSkills = make(map[int]int)
		lord.SkillEffectParams = make(map[int]map[int]int)
		lord.SkillEffectExParams = make(map[int]map[int]int)
		for pos, skill := range skillData {
			skillConf := g.GLordSkillConfig[skill.Id][skill.Index]
			skillId := skillConf.SkillId
			lord.ActiveSkills[pos] = skillId
			levelConf := g.GLordSkillLevelConfig[skill.Id][skill.Index][skill.Level]
			lord.SkillEffectParams[skillId] = levelConf.EffectParam
			lord.SkillEffectExParams[skillId] = levelConf.EffectExParam
		}
	}

	lord.Entity = &t.FightEntity{
		Pos:         lord.Pos,
		LordIndex:   lord.Pos / posRate,
		RawAttrs:    make(map[int]float64),
		Attrs:       make(map[int]float64),
		SkillCount:  make(map[int]int),
		Buffs:       make(map[int]*t.FightBuff),
		BuffEffects: make(map[int]bool),
		Effect: &t.FightSkillEffect{
			Effect:       make(map[int]int),
			TargetEffect: make(map[int]map[int]int),
		},
		WholeEffect:       make(map[int]int),
		WholeTargetEffect: make(map[int]map[int]int),
	}
	attrs := lord.Entity.Attrs
	rawAttrs := lord.Entity.RawAttrs
	for t, v := range lordAttr.Total {
		attrs[t] = v
		rawAttrs[t] = v
	}
	lord.AttrSum = int(attrs[c.AttrAttackCom] + attrs[c.AttrDefenseCom] + attrs[c.AttrLordDamage] + attrs[c.AttrLordDamageSub])

	for pos, guid := range actor.GetFightHeros() {
		if maxHeroPos > 0 && pos > maxHeroPos {
			continue
		}
		hero := actor.GetHeroStaticData(guid)
		heroTemp := getHeroTemplate(hero.Id)
		lord.Heros[pos] = heroTemp

		entity := &t.FightEntity{
			Pos:         lord.Pos + pos,
			LordIndex:   lord.Entity.LordIndex,
			HeroPos:     pos,
			RaceRatio:   heroTemp.RaceRatio,
			RawAttrs:    make(map[int]float64),
			Attrs:       make(map[int]float64),
			Feature:     heroTemp.Feature,
			Buffs:       make(map[int]*t.FightBuff),
			BuffEffects: make(map[int]bool),
			ImmuneBuff:  make(map[int]bool),
			Effect: &t.FightSkillEffect{
				Effect:       make(map[int]int),
				TargetEffect: make(map[int]map[int]int),
			},
			WholeEffect:       make(map[int]int),
			WholeTargetEffect: make(map[int]map[int]int),
		}
		attrs := entity.Attrs
		rawAttrs := entity.RawAttrs
		heroAttr := actor.GetHeroAttr(hero.Guid)
		for v, t := range heroAttr.Total {
			attrs[v] = t
			rawAttrs[v] = t
		}
		data.Entities[entity.Pos] = entity
		data.RawEntities[entity.Pos] = entity
		//		log.Infof("pos:%d,attrs:%v", entity.Pos, entity.Attrs)
	}
}

//InitMonsterFightData 怪物战斗数据
func InitMonsterFightData(data *t.FightData, lord *t.FightLord,
	monsterLord t.MonsterLord, monsterHeros map[int]t.MonsterHero,
	monsterName string, monsterLevel int) {

	lordConf := g.GMonsterLordConfig[monsterLord.Id]
	lord.Model = lordConf.Model
	if len(monsterName) > 0 {
		lord.Name = monsterName
	} else {
		lord.Name = lordConf.Name
	}
	guard := lordConf.GuardModel
	lord.Gmodel = []int{guard[1], guard[2]}

	lord.Equips = make(map[int]int)
	for _, equipId := range lordConf.Equips {
		equipConf := g.GLordEquipConfig[equipId]
		lord.Equips[equipConf.Pos] = equipId
	}

	if len(lordConf.LordPassive) > 0 {
		lord.PassSkills = make([]int, len(lordConf.LordPassive))
		for index, id := range lordConf.LordPassive {
			lord.PassSkills[index-1] = id
		}
	}

	lord.ActiveSkills = make(map[int]int)
	for pos, id := range lordConf.LordActive {
		lord.ActiveSkills[pos] = id
	}

	lord.Entity = &t.FightEntity{
		Pos:         lord.Pos,
		LordIndex:   lord.Pos / posRate,
		RawAttrs:    make(map[int]float64),
		Attrs:       make(map[int]float64),
		SkillCount:  make(map[int]int),
		Buffs:       make(map[int]*t.FightBuff),
		BuffEffects: make(map[int]bool),
		Effect: &t.FightSkillEffect{
			Effect:       make(map[int]int),
			TargetEffect: make(map[int]map[int]int),
		},
		WholeEffect:       make(map[int]int),
		WholeTargetEffect: make(map[int]map[int]int),
	}
	var level int
	if monsterLevel > 0 {
		level = monsterLevel
	} else {
		level = monsterLord.Level
	}

	attrs := lord.Entity.Attrs
	rawAttrs := lord.Entity.RawAttrs
	lordTemplateConf := g.GMonsterLordTemplateConfig[monsterLord.Template][level]
	for _, attr := range lordTemplateConf.Attr {
		attrs[attr.Type] = float64(attr.Value)
		rawAttrs[attr.Type] = float64(attr.Value)
	}

	lord.AttrSum = int(attrs[c.AttrAttackCom] + attrs[c.AttrDefenseCom] + attrs[c.AttrLordDamage] + attrs[c.AttrLordDamageSub])

	for _, monsterHero := range monsterHeros {
		heroTemp := getHeroTemplate(monsterHero.Id)
		lord.Heros[monsterHero.Pos] = heroTemp

		var level int
		if monsterLevel > 0 {
			level = monsterLevel
		} else {
			level = monsterHero.Level
		}

		entity := &t.FightEntity{
			Pos:         lord.Pos + monsterHero.Pos,
			LordIndex:   lord.Entity.LordIndex,
			HeroPos:     monsterHero.Pos,
			RaceRatio:   heroTemp.RaceRatio,
			RawAttrs:    calcMonsterAttr(monsterHero.Id, monsterHero.Template, level, nil),
			Attrs:       make(map[int]float64),
			Feature:     heroTemp.Feature,
			Buffs:       make(map[int]*t.FightBuff),
			BuffEffects: make(map[int]bool),
			ImmuneBuff:  make(map[int]bool),
			Effect: &t.FightSkillEffect{
				Effect:       make(map[int]int),
				TargetEffect: make(map[int]map[int]int),
			},
			WholeEffect:       make(map[int]int),
			WholeTargetEffect: make(map[int]map[int]int),
		}

		attrs := entity.Attrs
		for t, v := range entity.RawAttrs {
			attrs[t] = v
		}

		data.Entities[entity.Pos] = entity
		data.RawEntities[entity.Pos] = entity

		//log.Infof("pos:%d,attrs:%v", entity.Pos, entity.Attrs)
	}
}

func calcMonsterAttr(id, template, level int, exAttrs map[int]t.Attr) map[int]float64 {
	monsterConf := g.GMonsterConfig[id]
	templateConf := g.GMonsterTemplateConfig[template][level]

	attrs := make(map[int]float64)
	for _, attr := range templateConf.Attr {
		attrs[attr.Type] = float64(attr.Value)
	}

	for _, attr := range monsterConf.AttrRate {
		attrs[attr.Type] = float64(int(attrs[attr.Type] * float64(attr.Value) / 10000))
	}

	for _, attr := range monsterConf.AttrCom {
		attrs[attr.Type] += float64(attr.Value)
	}

	for _, attr := range exAttrs {
		attrs[attr.Type] += float64(attr.Value)
	}

	attr.CalcEntityFightAttr(attrs)

	return attrs
}
