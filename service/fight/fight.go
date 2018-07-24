package fight

import (
	"fmt"
	"math"

	"github.com/sencydai/gameworld/base"
	c "github.com/sencydai/gameworld/constdefine"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service/attr"
	"github.com/sencydai/gameworld/timer"
	t "github.com/sencydai/gameworld/typedefine"

	_ "github.com/sencydai/gameworld/log"
)

func startFighting(actor *t.Actor, fightData *t.FightData) {
	if fightData.Data[0].AttrSum < fightData.Data[1].AttrSum {
		fightData.Order[0], fightData.Order[1] = fightData.Order[1], fightData.Order[0]
	}

	//英雄战斗开始时技能
	for _, entity := range fightData.Entities {
		entityFighting(triggerBegin, fightData, entity, nil)
	}

	//双方领主释放被动技能
	for _, index := range fightData.Order {
		lord := fightData.Data[index]
		for _, skillId := range lord.PassSkills {
			useSkill(triggerBegin, fightData, lord.Entity, nil, skillId)
		}
	}

	//战斗主循环
	loopRoundFighting(actor, fightData)
}

func loopRoundFighting(actor *t.Actor, fightData *t.FightData) {
	onSendFightLogs(actor, fightData)

	//战斗结束
	if fightData.FightResult != 0 {
		timer.Next(actor, fmt.Sprintf("onFightClear_%d", fightData.Type), onFightClear, fightData)
		return
	}
	roundFighting(actor, fightData)
	//timer.NextGo(actor, fmt.Sprintf("roundFighting_%d", fightData.Type), roundFighting, fightData)
}

func roundFighting(actor *t.Actor, fightData *t.FightData) {
	data := actor.GetFightData()
	if data == nil || data.Guid != fightData.Guid || fightData.FightResult != 0 {
		return
	}

	//tick := time.Now()

	fightData.Round++

	//领主释放主动技能
	for _, i := range fightData.Order {
		lord := fightData.Data[i]
		skillPos := int(fightData.Round) % maxLordSkillCount
		if skillPos == 0 {
			skillPos = maxLordSkillCount
		}
		if skillId, ok := lord.ActiveSkills[skillPos]; ok {
			useSkill(triggerAction, fightData, lord.Entity, nil, skillId)
			if fightData.FightResult != 0 {
				break
			}
		}
	}

	actionEntitis := make(map[int]bool)
	for fightData.FightResult == 0 && len(actionEntitis) < len(fightData.Entities) {
		var selEntity *t.FightEntity
		for _, entity := range fightData.Entities {
			if _, ok := actionEntitis[entity.Pos]; !ok {
				if selEntity == nil || selEntity.Attrs[c.AttrSpeed] < entity.Attrs[c.AttrSpeed] {
					selEntity = entity
				}
			}
		}

		if selEntity == nil {
			break
		}
		actionEntitis[selEntity.Pos] = true

		selEntity.ReAction = true
		for selEntity.ReAction {
			selEntity.ReAction = false
			selEntity.IsAction = true
			for _, point := range actionFlow {
				entityFighting(point, fightData, selEntity, nil)
				if selEntity.IsDead || fightData.FightResult != 0 {
					break
				}
			}

			if fightData.FightResult != 0 {
				break
			}
			selEntity.IsAction = false
			//清除技能效果
			for _, entity := range fightData.Entities {
				resetEntityEffect(entity)
			}

			if !selEntity.IsDead {
				newFightLogActionFinish(fightData, selEntity.Pos)
			}
		}
	}

	loopRoundFighting(actor, fightData)

	//	log.Infof("fight:%d,round:%d,cost:%v", fightData.Type, fightData.Round, time.Since(tick))
}

func entityFighting(point int, fightData *t.FightData, entity *t.FightEntity, targetEntity *t.FightEntity) {
	//查找此触发点的所有buff
	if buffTriggers[point] {
		var logEffect *t.FightLogEffect
		buffs := entity.Buffs
		for _, buff := range buffs {
			if buff.Point != point {
				continue
			}
			if logEffect == nil {
				logEffect = newFightLogBuff(fightData, entity.Pos)
			}
			//buff是否已触发
			if !buff.IsTrigger {
				buff.IsTrigger = true
				triggerBuff(fightData, entity, buff, logEffect, true)
			}
			buff.Round--
			newFightLogBuffRound(logEffect, buff.Guid, buff.BuffId, buff.Round)
			//buff消失
			if buff.Round < 0 {
				triggerBuff(fightData, entity, buff, logEffect, false)
			} else if !buffChangeHp(fightData, entity, buff, logEffect) {
				return
			}
		}
	}

	buffEffects := entity.BuffEffects
	//是否眩晕及变羊,不能释放普通攻击和技能
	if buffEffects[buffDizzy] || buffEffects[buffDizzyAndDefenseSub] {
		return
	}
	hero := fightData.Data[entity.LordIndex].Heros[entity.HeroPos]
	skills := hero.Skills[point]
	//如果没有技能，且为攻击时，触发普通攻击
	//是否沉默，沉默只能释放普通攻击，且普通攻击只在攻击时触发
	if len(skills) == 0 || buffEffects[buffSilence] {
		if point == triggerAttack {
			useSkill(point, fightData, entity, targetEntity, hero.CommSkill)
		}
		return
	}
	var isTrigger bool
	for _, skillId := range skills {
		skillConf := g.GSkillConfig[skillId]
		if base.Rand(1, 10000) > skillConf.SkillRandom {
			continue
		}
		if skillConf.SkillType == activeSkill {
			isTrigger = true
		}
		useSkill(point, fightData, entity, targetEntity, skillId)
	}

	//如果没有触发任何技能，且为攻击时
	if !isTrigger && point == triggerAttack {
		useSkill(point, fightData, entity, targetEntity, hero.CommSkill)
	}
}

func useSkill(point int, fightData *t.FightData, entity *t.FightEntity, targetEntity *t.FightEntity, skillId int) {
	if count, ok := skillRepeatCounts[skillId]; ok {
		if entity.SkillCount[skillId] >= count {
			return
		}
		entity.SkillCount[skillId]++
	}

	var targets []*t.FightEntity
	notSelfDeadPoint := point != triggerSelfDead

	effectConfs := g.GSkillEffectConfig[skillId]
	for i := 1; i <= len(effectConfs); i++ {
		effectConf := effectConfs[i]
		if base.Rand(1, 10000) > effectConf.Random {
			continue
		}

		//更新目标
		if effectConf.Target > 0 {
			targets = selectTarget(fightData, entity, targetEntity,
				effectConf.Target, effectConf.Effect, effectConf.TargetParam, effectConf.TargetSpec)
		} else {
			//非复活效果
			if effectConf.Effect != sEffectReliveTarget {
				for j := len(targets) - 1; j >= 0; j-- {
					if targets[j].IsDead {
						targets = append(targets[0:j], targets[j+1:]...)
					}
				}
			} else {
				for j := len(targets) - 1; j >= 0; j-- {
					if !targets[j].IsDead {
						targets = append(targets[0:j], targets[j+1:]...)
					}
				}
			}
		}

		//目标前置条件
		var tmpTargets []*t.FightEntity
		if effectConf.PerCond > 0 {
			//当前置条件不满足时，是否也将不满足的目标保留到下一个效果
			if effectConf.ReservePre > 0 {
				tmpTargets = make([]*t.FightEntity, len(targets))
				copy(tmpTargets, targets)
			}

			targets = filterTargetByPreCondition(targets, fightData, effectConf.PerCond, effectConf.PerCondParam)
		}

		//选中目标
		if len(targets) > 0 {
			logSkill := newFightLogSkillSelTarget(fightData, entity.Pos, skillId, i)
			effects := logSkill.Effects.([]int)
			for _, t := range targets {
				effects = append(effects, t.Pos)
			}
			logSkill.Effects = effects

			//执行技能效果
			execSkillEffect(point, fightData, entity, targets, effectConfs, effectConf)

			if fightData.FightResult != 0 || (entity.IsDead && notSelfDeadPoint) {
				break
			}
		}
	}

	//行动结束
	if !entity.IsAction && !entity.IsDead && fightData.FightResult != 0 {
		newFightLogActionFinish(fightData, entity.Pos)
	}
}

func execSkillEffect(point int, fightData *t.FightData, entity *t.FightEntity, targets []*t.FightEntity,
	effectConfs map[int]*t.SkillEffectConfig, effectConf *t.SkillEffectConfig) {
	switch effectConf.Effect {
	//技能伤害
	case sEffectDamage:
		fallthrough
	case sEffectMaxHpDamage:
		fallthrough
	case sEffectHpDamage:
		fallthrough
	case sEffectRealDamage:
		execSkillDamageEffect(fightData, entity, targets, effectConf)
	//攻击多次
	case sEffectMultlAttack:
		effectParam := getLordSkillEffectParam(fightData, entity, effectConf)
		notSelfDeadPoint := point != triggerSelfDead
		for j := 0; j < int(effectParam); j++ {
			for k := 1; k < effectConf.Index; k++ {
				conf := effectConfs[k]
				//复活目标
				if conf.Effect == sEffectReliveTarget {
					for i := len(targets) - 1; i >= 0; i-- {
						if !targets[i].IsDead {
							targets = append(targets[0:i], targets[i+1:]...)
						}
					}
				} else {
					if entity.IsDead {
						return
					}
					for i := len(targets) - 1; i >= 0; i-- {
						if targets[i].IsDead {
							targets = append(targets[0:i], targets[i+1:]...)
						}
					}
				}

				if len(targets) == 0 {
					return
				}

				execSkillEffect(point, fightData, entity, targets, effectConfs, conf)
				if fightData.FightResult != 0 || (entity.IsDead && notSelfDeadPoint) {
					return
				}
			}
		}
	default:
		execSkillUnDamageEffect(point, fightData, entity, targets, effectConf)
	}
}

func execSkillUnDamageEffect(point int, fightData *t.FightData, entity *t.FightEntity, targets []*t.FightEntity, effectConf *t.SkillEffectConfig) {
	effect := effectConf.Effect
	effectParam := getLordSkillEffectParam(fightData, entity, effectConf)
	effectExParam := getLordSkillEffectExParam(fightData, entity, effectConf)
	logSkill := newFightLogSkillAction(fightData, entity.Pos, effectConf.Skill, effectConf.Index)
	switch effect {
	//目标每拥有一个减益BUFF伤害提高
	//目标每拥有一个增益BUFF伤害提高
	case sEffectDebuffImprove:
		fallthrough
	case sEffectGainImprove:
		wholeTargetEffect := entity.WholeTargetEffect
		for _, target := range targets {
			effects, ok := wholeTargetEffect[target.Pos]
			if !ok {
				effects = make(map[int]int)
				wholeTargetEffect[target.Pos] = effects
			}
			effects[effect] = int(effectParam)
		}
		newFightLogEffectActionResult(logSkill, entity.Pos, feedbackDamageImprove, 0)
	//目标每拥有一个增益BUFF防御提高
	case sEffectGainDefenseImprove:
		entity.Effect.Effect[effect] += int(effectParam) * len(targets)
		newFightLogEffectActionResult(logSkill, entity.Pos, feedbackDefenseRise, 0)
	//每拥有一个敌人，防御上升百分比
	//每拥有一个友军，防御上升百分比
	case sEffectEnemyDefenseImp:
		fallthrough
	case sEffectPartnerDefenseImp:
		for _, target := range targets {
			target.Effect.Effect[effect] += int(effectParam)
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackDefenseRise, 0)
		}
	//反伤(全程有效)
	case sEffectThorns:
		for _, target := range targets {
			target.WholeEffect[effect] += int(effectParam)
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackThorns, 0)
		}
	//吸血
	case sEffectSuckBlood:
		for _, target := range targets {
			target.Effect.Effect[effect] += int(effectParam)
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackSuckBlood, 0)
		}
	//全程吸血
	case sEffectWholeSuckBlood:
		for _, target := range targets {
			target.WholeEffect[effect] += int(effectParam)
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackSuckBlood, 0)
		}
	//生命恢复（值）
	case sEffectHpRecover:
		for _, target := range targets {
			//无法恢复
			if checkBuff(target.Buffs, buffUnRecover) {
				newFightLogEffectActionResult(logSkill, target.Pos, feedbackUnRecoverAction, 0)
			} else {
				oldValue := target.Attrs[c.AttrHp]
				target.Attrs[c.AttrHp] = math.Min(target.RawAttrs[c.AttrHp], oldValue+effectParam)
				newFightLogEffectActionResult(logSkill, target.Pos, feedbackHpRecover, int(target.Attrs[c.AttrHp]-oldValue))
			}
		}
	//生命恢复（%）
	case sEffectHpRecoverPct:
		effectParam /= 10000
		for _, target := range targets {
			//无法恢复
			if checkBuff(target.Buffs, buffUnRecover) {
				newFightLogEffectActionResult(logSkill, target.Pos, feedbackUnRecoverAction, 0)
			} else {
				oldValue := target.Attrs[c.AttrHp]
				rawHp := target.RawAttrs[c.AttrHp]
				target.Attrs[c.AttrHp] = math.Min(rawHp, oldValue+float64(int(rawHp*effectParam)))
				newFightLogEffectActionResult(logSkill, target.Pos, feedbackHpRecover, int(target.Attrs[c.AttrHp]-oldValue))
			}
		}
	//反击(全程有效)
	case sEffectHitBack:
		for _, target := range targets {
			target.WholeEffect[effect] = 1
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackHitBack, 0)
		}
	//减少受到伤害（值）
	//减少受到伤害（%
	case sEffectReduceDamage:
		fallthrough
	case sEffectReduceDamagePct:
		for _, target := range targets {
			target.Effect.Effect[effect] += int(effectParam)
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackReduceDamage, 0)
		}
	//添加某buff
	case sEffectAddBuff:
		for _, target := range targets {
			action := newFightLogEffectAction(logSkill, target.Pos, false)
			newFightLogEffectResult(action.Effect, feedbackAddBuff, 0, -1)
			addBuffById(fightData, target, action.Effect, int(effectParam), true)
		}
	//消除某类型buff
	case sEffectClearBuffType:
		for _, target := range targets {
			action := newFightLogEffectAction(logSkill, target.Pos, false)
			newFightLogEffectResult(action.Effect, feedbackClearBuff, 0, -1)
			entity.Effect.ClearBuff += clearBuffs(fightData, target, action.Effect, 0, int(effectParam), 0)
		}
	//随机消除目标N个增益buff
	case sEffectClearGainBuff:
		for _, target := range targets {
			action := newFightLogEffectAction(logSkill, target.Pos, false)
			newFightLogEffectResult(action.Effect, feedbackClearBuff, 0, -1)
			entity.Effect.ClearBuff += randClearBuffByGainType(fightData, target, action.Effect, gainBuff, int(effectParam))
		}
	//随机消除目标N个减益buff
	case sEffectClearDebuff:
		for _, target := range targets {
			action := newFightLogEffectAction(logSkill, target.Pos, false)
			newFightLogEffectResult(action.Effect, feedbackClearBuff, 0, -1)
			entity.Effect.ClearBuff += randClearBuffByGainType(fightData, target, action.Effect, deBuff, int(effectParam))
		}
	//免疫某类型buff
	case sEffectImmuneBuff:
		for _, target := range targets {
			target.ImmuneBuff[int(effectParam)] = true
			action := newFightLogEffectAction(logSkill, target.Pos, false)
			newFightLogEffectResult(action.Effect, feedbackImmunebuff, int(effectParam), -1)
			entity.Effect.ClearBuff += clearBuffs(fightData, target, action.Effect, 0, int(effectParam), 0)
		}
	//获得随机增益BUFF
	case sEffectRandomGain:
		for _, target := range targets {
			action := newFightLogEffectAction(logSkill, target.Pos, false)
			newFightLogEffectResult(action.Effect, feedbackAddBuff, 0, -1)
			randGainBuff(fightData, target, action.Effect, gainBuff, int(effectParam))
		}
	//获得随机减益BUFF
	case sEffectRandomDebuff:
		for _, target := range targets {
			action := newFightLogEffectAction(logSkill, target.Pos, false)
			newFightLogEffectResult(action.Effect, feedbackAddBuff, 0, -1)
			randGainBuff(fightData, target, action.Effect, deBuff, int(effectParam))
		}
	//某类型BUFF持续时间增加
	case sEffectAddRound:
		for _, target := range targets {
			action := newFightLogEffectAction(logSkill, target.Pos, false)
			newFightLogEffectResult(action.Effect, feedbackBuffAddRound, 0, -1)
			changeBuffRound(fightData, target, action.Effect, int(effectExParam), int(effectParam), 0)
		}
	//某类型BUFF持续时间减少
	case sEffectReduceRound:
		for _, target := range targets {
			action := newFightLogEffectAction(logSkill, target.Pos, false)
			newFightLogEffectResult(action.Effect, feedbackBuffReduceRound, 0, -1)
			changeBuffRound(fightData, target, action.Effect, -int(effectExParam), int(effectParam), 0)
		}
	//所有增益BUFF持续时间增加
	case sEffectGainAddRound:
		for _, target := range targets {
			action := newFightLogEffectAction(logSkill, target.Pos, false)
			newFightLogEffectResult(action.Effect, feedbackBuffAddRound, 0, -1)
			changeBuffRound(fightData, target, action.Effect, int(effectParam), 0, gainBuff)
		}
	//所有增益BUFF持续时间减少
	case sEffectGainReduceRound:
		for _, target := range targets {
			action := newFightLogEffectAction(logSkill, target.Pos, false)
			newFightLogEffectResult(action.Effect, feedbackBuffReduceRound, 0, -1)
			changeBuffRound(fightData, target, action.Effect, -int(effectParam), 0, gainBuff)
		}
	//每拥有一个敌人，攻击上升百分比
	//每拥有一个友军，攻击上升百分比
	case sEffectEnemyAttackImp:
		fallthrough
	case sEffectPartnerAttackImp:
		for _, target := range targets {
			target.Effect.Effect[effect] += int(effectParam)
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackAttackRise, 0)
		}
	//攻击百分比加成
	case sEffectAttackPct:
		for _, target := range targets {
			target.Effect.Effect[effect] += int(effectParam)
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackAttack, 0)
		}
	//防御百分比加成
	case sEffectDefensePct:
		for _, target := range targets {
			target.Effect.Effect[effect] += int(effectParam)
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackDefense, 0)
		}
	//暴击百分比加成
	case sEffectCritPct:
		for _, target := range targets {
			target.Effect.Effect[effect] += int(effectParam)
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackCritPct, 0)
		}
	//无视防御(全程有效)
	case sEffectWholeIgnoreDefense:
		for _, target := range targets {
			target.WholeEffect[effect] += int(effectParam)
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackIgnoreDefense, 0)
		}
	//无视防御(一次有效)
	case sEffectIgnoreDefense:
		for _, target := range targets {
			target.Effect.Effect[effect] += int(effectParam)
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackIgnoreDefense, 0)
		}
	//重复一次行动流程
	case sEffectReAction:
		for _, target := range targets {
			target.ReAction = true
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackReAction, 0)
		}
	//特殊攻击(全程有效)
	case sEffectSpecAttack:
		for _, target := range targets {
			target.WholeEffect[effect] = 1
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackSpecAttack, 0)
		}
	//修改某属性(固定值)
	case sEffectChangeAttr:
		fallthrough
	//修改某属性(百分比)
	case sEffectChangeAttrPct:
		for _, target := range targets {
			attrs := target.Attrs
			if effect == sEffectChangeAttr {
				attrs[int(effectParam)] = math.Max(0, attrs[int(effectParam)]+effectExParam)
			} else {
				attrs[int(effectParam)] = float64(int(math.Max(0, attrs[int(effectParam)]*(1+effectExParam/10000))))
			}
			switch int(effectParam) {
			case c.AttrAttackBase:
				fallthrough
			case c.AttrExAttackPct:
				fallthrough
			case c.AttrExAttack:
				attr.CalcMinAttack(attrs)
				attr.CalcMaxAttack(attrs)
			case c.AttrMinAttackBase:
				attr.CalcMinAttack(attrs)
			case c.AttrMaxAttackBase:
				attr.CalcMaxAttack(attrs)
			case c.AttrDefenseBase:
				fallthrough
			case c.AttrExDefensePct:
				fallthrough
			case c.AttrExDefense:
				attr.CalcDefense(attrs)
			case c.AttrSpeedBase:
				fallthrough
			case c.AttrExSpeed:
				attr.CalcSpeed(attrs)
			}
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackChangeAttr, 0)
		}
	//复活目标
	case sEffectReliveTarget:
		for _, target := range targets {
			if !target.IsDead {
				continue
			}
			fightData.Entities[target.Pos] = target
			attrs := target.Attrs
			for t, v := range target.RawAttrs {
				attrs[t] = v
			}
			attrs[c.AttrHp] = float64(int(attrs[c.AttrHp] * effectParam / 10000))
			target.IsAction = false
			for t := range target.Buffs {
				delete(target.Buffs, t)
			}
			for t := range target.ImmuneBuff {
				delete(target.ImmuneBuff, t)
			}
			target.IsDead = false
			target.ReAction = false
			resetEntityEffect(target)
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackRelive, int(target.Attrs[c.AttrHp]))
		}
	}
}

func execSkillDamageEffect(fightData *t.FightData, entity *t.FightEntity, targets []*t.FightEntity, effectConf *t.SkillEffectConfig) {
	beAttacks := make(map[int]bool)
	var deadCount int
	for _, target := range targets {
		//被攻击时
		entityFighting(triggerBeAttack, fightData, target, entity)

		//队友被攻击时
		for _, partner := range getOriginalTarget(fightData, target, selTargetPartner) {
			if partner.Pos != target.Pos && !beAttacks[partner.Pos] {
				beAttacks[partner.Pos] = true
				entityFighting(triggerPartnerBeAttack, fightData, partner, entity)
			}
		}
		if target.IsDead {
			deadCount++
		}
	}

	if entity.IsDead || deadCount == len(targets) || fightData.FightResult != 0 {
		return
	}

	effectParam := getLordSkillEffectParam(fightData, entity, effectConf)
	effectExParam := getLordSkillEffectExParam(fightData, entity, effectConf)
	logSkill := newFightLogSkillAction(fightData, entity.Pos, effectConf.Skill, effectConf.Index)
	effectParam *= (entity.RaceRatio + 1) / 10000

	for _, target := range targets {
		if target.IsDead {
			continue
		}
		//计算伤害
		calcDamage(fightData, entity, target, effectConf.Effect, effectParam, effectExParam, logSkill)
		if !target.IsDead {
			entityFighting(triggerBeAttackFinish, fightData, target, entity)
		}
		if entity.IsDead || fightData.FightResult != 0 {
			return
		}
	}
}

func calcDamage(fightData *t.FightData, entity, target *t.FightEntity, effect int, effectParam, effectExParam float64, logSkill *t.FightLogSkill) {
	//无敌
	if checkBuff(target.Buffs, buffInvincible) {
		newFightLogEffectActionResult(logSkill, target.Pos, feedbackInvincible, 0)
		return
	}

	var (
		entityAttrs   = entity.Attrs
		targetAttrs   = target.Attrs
		targetEffects = target.Effect.Effect
		damage        float64
		critRatio     float64 = 1 //暴击倍率
		specAttack    bool
	)
	//领主
	if entity.HeroPos == 0 {
		switch effect {
		case sEffectDamage:
			lord := getLordByEntity(fightData, target)
			lordAttrs := lord.Entity.Attrs
			damage = entityAttrs[c.AttrLordDamage]*(1+entityAttrs[c.AttrLordDamagePct]/10000) - lordAttrs[c.AttrLordDamageSub]*(1+lordAttrs[c.AttrLordDamageSubPct]/10000)
			damage = math.Max(damage, entityAttrs[c.AttrLordDamage]*0.1)*effectParam + effectExParam
		case sEffectMaxHpDamage:
			damage = target.RawAttrs[c.AttrHp] * effectParam
		case sEffectHpDamage:
			damage = target.Attrs[c.AttrHp] * effectParam
		case sEffectRealDamage:
			damage = entityAttrs[c.AttrLordDamage] * (1 + entityAttrs[c.AttrLordDamagePct]/10000) * effectParam
		}
	} else {
		//闪避
		if base.Rand(1, 10000) <= int(targetAttrs[c.AttrDodge]-entityAttrs[c.AttrHit]) {
			//自己发生了闪避
			entityFighting(triggerSelfDodge, fightData, target, entity)
			//队友发生了闪避
			for _, partner := range getOriginalTarget(fightData, target, selTargetPartner) {
				if partner.Pos != target.Pos {
					entityFighting(triggerPartnerDodge, fightData, partner, entity)
				}
			}
			newFightLogEffectActionResult(logSkill, target.Pos, feedbackDodge, 0)
			return
		}

		switch effect {
		case sEffectDamage:
			//暴击
			crit := entityAttrs[c.AttrCrit] - targetAttrs[c.AttrCritDefense] + float64(targetEffects[sEffectCritPct])
			if base.Rand(1, 10000) < int(crit) {
				critRatio = entityAttrs[c.AttrCritDamage] / 10000
			}

			attack := calcAttack(fightData, entity)
			defense := calcDefense(fightData, target)

			//无视防御
			var ignoreDefense int
			if v, ok := targetEffects[sEffectIgnoreDefense]; ok {
				ignoreDefense = v
				delete(targetEffects, sEffectIgnoreDefense)
			}
			ignoreDefense += target.WholeEffect[sEffectWholeIgnoreDefense]
			if ignoreDefense > 0 {
				defense *= math.Max(0, 1-float64(ignoreDefense)/10000)
			}
			damage = math.Max(attack-defense, attack*0.1)*effectParam + effectExParam

			//伤害加深百分比
			damageAddPct := int(entityAttrs[c.AttrDamageAddPct])
			if effects, ok := entity.WholeTargetEffect[target.Pos]; ok {
				if v, ok := effects[sEffectDebuffImprove]; ok {
					damageAddPct += v * calcGainTypeCount(target.Buffs, deBuff)
				}
				if v, ok := effects[sEffectGainImprove]; ok {
					damageAddPct += v * calcGainTypeCount(target.Buffs, gainBuff)
				}
			}

			//伤害减免百分比
			damageSubPct := int(targetAttrs[c.AttrDamageSubPct])
			if v, ok := targetEffects[sEffectReduceDamagePct]; ok {
				damageSubPct += v
				delete(targetEffects, sEffectReduceDamagePct)
			}

			damage *= critRatio * (1 + float64(damageAddPct-damageSubPct)/10000)

			//伤害减免固定值
			damageSub := int(targetAttrs[c.AttrDamageSub])
			if v, ok := targetEffects[sEffectReduceDamage]; ok {
				damageSub += v
				delete(targetEffects, sEffectReduceDamage)
			}

			damage += entityAttrs[c.AttrDamageAdd] - float64(damageSub)

			//伤害修正
			entityLord := getLordByEntity(fightData, entity)
			if entityLord.Power > 0 {
				targetLord := getLordByEntity(fightData, target)
				if targetLord.Power > 0 {
					power := float64(entityLord.Power) / float64(targetLord.Power)
					if power < 0.8 {
						power += 0.2
					} else if power > 1.2 {
						power -= 0.2
					} else {
						power = 1
					}
					damage *= power
				}
			}
		case sEffectMaxHpDamage:
			damage = target.RawAttrs[c.AttrHp] * effectParam
		case sEffectHpDamage:
			damage = target.Attrs[c.AttrHp] * effectParam
		case sEffectRealDamage:
			damage = calcAttack(fightData, entity) * effectParam
		}
	}

	damage = float64(int(math.Max(0, damage)))
	action := newFightLogEffectAction(logSkill, target.Pos, false)
	//护盾
	v := calcShield(fightData, target, damage, action.Effect)
	if damage > v {
		newFightLogEffectResult(action.Effect, feedbackShield, int(damage-v), -1)
	}
	damage = v

	//吸血
	suckBlood := entityAttrs[c.AttrSuckBlood] + float64(entity.Effect.Effect[sEffectSuckBlood]) + float64(entity.WholeEffect[sEffectWholeSuckBlood])
	if suckBlood > 0 {
		hp := int(damage * suckBlood / 10000)
		old := entityAttrs[c.AttrHp]
		entityAttrs[c.AttrHp] = math.Min(entity.RawAttrs[c.AttrHp], old+float64(hp))
		entityAction := newFightLogEffectAction(logSkill, entity.Pos, true)
		newFightLogEffectResult(entityAction.Effect, feedbackSuckBloodAction, int(entityAttrs[c.AttrHp]-old), -1)
	}

	_, specAttack = entity.WholeEffect[sEffectSpecAttack]
	if !specAttack {
		//反伤
		thorns := targetAttrs[c.AttrThorns] + float64(target.WholeEffect[sEffectThorns])
		if thorns > 0 {
			entityAction := newFightLogEffectAction(logSkill, entity.Pos, true)
			hp := int(damage * thorns / 10000)
			//如果攻击者有护盾
			hpTmp := int(calcShield(fightData, entity, float64(hp), entityAction.Effect))
			if hp > hpTmp {
				newFightLogEffectResult(entityAction.Effect, feedbackShield, hp-hpTmp, -1)
			}
			hp = hpTmp
			old := entityAttrs[c.AttrHp]
			entityAttrs[c.AttrHp] = math.Max(0, old-float64(hp))
			newFightLogEffectResult(entityAction.Effect, feedbackthornsHpReduce, int(entityAttrs[c.AttrHp]-old), -1)
		}
	}

	old := targetAttrs[c.AttrHp]
	//log.Infof("entity:%d,old:%f,damage:%f,new:%f", target.Pos, old, damage, targetAttrs[c.AttrHp])
	targetAttrs[c.AttrHp] = math.Max(0, old-damage)

	targetAction := newFightLogEffectAction(logSkill, target.Pos, true)
	if int(critRatio) != 1 {
		newFightLogEffectResult(targetAction.Effect, feedbackCrit, -int(damage), -1)
	} else {
		newFightLogEffectResult(targetAction.Effect, feedbackDamage, -int(damage), -1)
	}

	checkAlive(fightData, target)

	if entity.HeroPos == 0 {
		return
	}

	checkAlive(fightData, entity)
	if fightData.FightResult != 0 {
		return
	}

	if int(critRatio) != 1 {
		//自己暴击时
		if !entity.IsDead {
			entityFighting(triggerSelfCrit, fightData, entity, target)
		}
		//队友暴击
		for _, partner := range getOriginalTarget(fightData, entity, selTargetPartner) {
			if partner.Pos != entity.Pos {
				entityFighting(triggerPartnerCrit, fightData, partner, target)
			}
		}
	}

	//反击
	if !specAttack && !target.IsDead && target.WholeEffect[sEffectHitBack] == 1 {
		newFightLogEffectResult(targetAction.Effect, feedbackHitBackAction, 0, -1)
		entityFighting(triggerHitBack, fightData, target, entity)
	}

	//技能造成伤害时
	if !target.IsDead {
		entityFighting(triggerSkillDamage, fightData, entity, target)
	} else {
		//目标死亡时
		entityFighting(triggerTargetDead, fightData, entity, target)
	}
}

func triggerBuff(fightData *t.FightData, entity *t.FightEntity, buff *t.FightBuff, logEffect *t.FightLogEffect, isTrigger bool) {
	value := float64(buff.Value)
	//减益
	if buff.GainType == deBuff {
		value = -value
	}
	if isTrigger {
		fightData.BuffIndex++
		buff.Index = fightData.BuffIndex
	} else {
		value = -value
		delete(entity.Buffs, buff.Guid)
	}
	switch buff.Type {
	//攻击
	case buffAttack:
		switch buff.ValueType {
		case fixedAddtion:
			entity.Attrs[c.AttrMinAttack] = math.Max(0, entity.Attrs[c.AttrMinAttack]+value)
			entity.Attrs[c.AttrMaxAttack] = math.Max(0, entity.Attrs[c.AttrMaxAttack]+value)
		case percentAddtion:
			value /= 10000
			entity.Attrs[c.AttrMinAttack] = math.Max(0, entity.Attrs[c.AttrMinAttack]+float64(int(entity.RawAttrs[c.AttrMinAttack]*value)))
			entity.Attrs[c.AttrMaxAttack] = math.Max(0, entity.Attrs[c.AttrMaxAttack]+float64(int(entity.RawAttrs[c.AttrMaxAttack]*value)))
		}
	//防御
	case buffDefense:
		switch buff.ValueType {
		case fixedAddtion:
			entity.Attrs[c.AttrDefense] = math.Max(0, entity.Attrs[c.AttrDefense]+value)
		case percentAddtion:
			value /= 10000
			entity.Attrs[c.AttrDefense] = math.Max(0, entity.Attrs[c.AttrDefense]+float64(int(entity.RawAttrs[c.AttrDefense]*value)))
		}
	//速度
	case buffSpeed:
		switch buff.ValueType {
		case fixedAddtion:
			entity.Attrs[c.AttrSpeed] = math.Max(0, entity.Attrs[c.AttrSpeed]+value)
		case percentAddtion:
			value /= 10000
			entity.Attrs[c.AttrSpeed] = math.Max(0, entity.Attrs[c.AttrSpeed]+float64(int(entity.RawAttrs[c.AttrSpeed]*value)))
		}
	//暴击
	case buffCrit:
		switch buff.ValueType {
		case percentAddtion:
			entity.Attrs[c.AttrCrit] = math.Max(0, entity.Attrs[c.AttrCrit]+value)
		}
	//命中
	case buffHit:
		switch buff.ValueType {
		case percentAddtion:
			entity.Attrs[c.AttrHit] = math.Max(0, entity.Attrs[c.AttrHit]+value)
		}
	//护盾
	case buffShield:
		if isTrigger && buff.ValueType == percentAddtion {
			buff.ValueType = fixedAddtion
			buff.Value = int(entity.RawAttrs[c.AttrHp] * value / 10000)
		}
	//免疫
	case buffImmune:
		if isTrigger {
			clearBuffs(fightData, entity, logEffect, buff.Guid, 0, 0)
		}
	//变羊（眩晕加破甲）
	case buffDizzyAndDefenseSub:
		clearBuffs(fightData, entity, logEffect, buff.Guid, 0, 0)
		if isTrigger {
			for t := range entity.Attrs {
				if t != c.AttrHp {
					entity.Attrs[t] = 0
				}
			}
		} else {
			for t := range entity.Attrs {
				if t != c.AttrHp {
					entity.Attrs[t] = entity.RawAttrs[t]
				}
			}
			delete(entity.BuffEffects, buffDizzyAndDefenseSub)
		}
	//吸血
	case buffSuckBlood:
		switch buff.ValueType {
		case percentAddtion:
			entity.Attrs[c.AttrSuckBlood] = math.Max(0, entity.Attrs[c.AttrSuckBlood]+value)
		}
	//沉默
	//眩晕
	case buffSilence:
		fallthrough
	case buffDizzy:
		if isTrigger {
			entity.BuffEffects[buff.Type] = true
		} else {
			delete(entity.BuffEffects, buff.Type)
		}
	}
}

func buffChangeHp(fightData *t.FightData, entity *t.FightEntity, buff *t.FightBuff, logEffect *t.FightLogEffect) bool {
	if entity.IsDead {
		return false
	}
	if _, ok := buffHpTypes[buff.Type]; !ok {
		return true
	}

	//无法恢复
	if checkBuff(entity.Buffs, buffUnRecover) {
		newFightLogEffectResult(logEffect, feedbackUnRecoverAction, 0, -1)
		return true
	}

	value := float64(buff.Value)
	switch buff.Type {
	//生命恢复(当前生命)
	case buffHpRecover:
		if buff.ValueType == percentAddtion {
			value = entity.Attrs[c.AttrHp] * value / 10000
		}
	//持续掉血(最大生命)
	case buffConDamageMax:
		value = -value
		if buff.ValueType == percentAddtion {
			value = entity.RawAttrs[c.AttrHp] * value / 10000
		}
	//持续掉血(当前生命)
	case buffConDamage:
		value = -value
		if buff.ValueType == percentAddtion {
			value = entity.Attrs[c.AttrHp] * value / 10000
		}
	//生命恢复(生命上限)
	case buffHpRecoverMax:
		if buff.ValueType == percentAddtion {
			value = entity.RawAttrs[c.AttrHp] * value / 10000
		}
	default:
		return true
	}
	oldValue := entity.Attrs[c.AttrHp]
	if value < 0 {
		value = -calcShield(fightData, entity, -value, logEffect)
		entity.Attrs[c.AttrHp] = float64(int(math.Max(0, entity.Attrs[c.AttrHp]+value)))
		newFightLogEffectResult(logEffect, feedbackBuffHpReduce, int(entity.Attrs[c.AttrHp]-oldValue), buff.BuffId)
		return checkAlive(fightData, entity)
	}

	value = math.Min(entity.RawAttrs[c.AttrHp]-entity.Attrs[c.AttrHp], value)
	entity.Attrs[c.AttrHp] += float64(int(value))
	newFightLogEffectResult(logEffect, feedbackBuffHpAdd, int(entity.Attrs[c.AttrHp]-oldValue), buff.BuffId)

	return true
}

//计算护盾抵消伤害值
func calcShield(fightData *t.FightData, entity *t.FightEntity, damage float64, logEffect *t.FightLogEffect) float64 {
	count := int(damage)
	for _, buff := range getBuffs(entity.Buffs, buffShield, true) {
		if count >= buff.Value {
			count -= buff.Value
			triggerBuff(fightData, entity, buff, logEffect, false)
			newFightLogBuffRound(logEffect, buff.Guid, buff.BuffId, -1)
		} else {
			buff.Value -= count
			count = 0
		}
		if count == 0 {
			break
		}
	}

	return float64(count)
}

func entityDead(fightData *t.FightData, entity *t.FightEntity) {
	entity.IsDead = true
	delete(fightData.Entities, entity.Pos)
	newFightLogDead(fightData, entity.Pos)

	if checkOver(fightData) {
		return
	}

	//自己死亡时
	entityFighting(triggerSelfDead, fightData, entity, nil)
	//队友死亡时
	for _, partner := range getOriginalTarget(fightData, entity, selTargetPartner) {
		if fightData.FightResult != 0 {
			return
		}
		entityFighting(triggerPartnerDead, fightData, partner, entity)
	}
}
