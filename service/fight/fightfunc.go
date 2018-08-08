package fight

import (
	"bytes"
	"fmt"
	"math"
	"sort"

	"github.com/sencydai/gameworld/base"
	c "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/dispatch"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/service/bag"
	"github.com/sencydai/gameworld/timer"
	t "github.com/sencydai/gameworld/typedefine"
)

func init() {
	dispatch.RegActorMsgHandle(proto.Fight, proto.FightCGetAwards, onGetFightAwards)
	dispatch.RegActorMsgHandle(proto.Fight, proto.FightCGiveup, onGiveupFight)
	dispatch.RegActorMsgHandle(proto.Fight, proto.FightCNextRound, onNextRound)
	service.RegActorLogout(onActorLogout)
}

func checkOver(fightData *t.FightData) bool {
	if fightData.FightResult != 0 {
		return true
	}

	alives := make(map[int]int)
	for _, entity := range fightData.Entities {
		alives[entity.LordIndex]++
	}

	if alives[0] == 0 {
		fightData.FightResult = Lose
	} else if alives[1] == 0 {
		fightData.FightResult = Win
	}

	return fightData.FightResult != 0
}

func checkAlive(fightData *t.FightData, entity *t.FightEntity) bool {
	if entity.HeroPos == 0 {
		return true
	}

	if entity.IsDead {
		return false
	}
	if int(entity.Attrs[c.AttrHp]) > 0 {
		return true
	}

	entityDead(fightData, entity)
	return false
}

func getLordByEntity(fightData *t.FightData, entity *t.FightEntity) *t.FightLord {
	return fightData.Data[entity.LordIndex]
}

func getLordSkillEffectParam(fightData *t.FightData, entity *t.FightEntity, effectConf *t.SkillEffectConfig) (effectParam float64) {
	effectParam = effectConf.EffectParam
	if entity.HeroPos == 0 {
		return
	}
	lord := getLordByEntity(fightData, entity)
	if params, ok := lord.SkillEffectParams[effectConf.Skill]; !ok {
		return
	} else if value, ok := params[effectConf.Index]; ok {
		effectParam = float64(value)
	}

	return
}

func getLordSkillEffectExParam(fightData *t.FightData, entity *t.FightEntity, effectConf *t.SkillEffectConfig) (effectExParam float64) {
	effectExParam = effectConf.EffectExParam
	if entity.HeroPos == 0 {
		return
	}
	lord := getLordByEntity(fightData, entity)
	if params, ok := lord.SkillEffectExParams[effectConf.Skill]; !ok {
		return
	} else if value, ok := params[effectConf.Index]; ok {
		effectExParam = float64(value)
	}

	return
}

func resetEntityEffect(entity *t.FightEntity) {
	effect := entity.Effect
	for k := range effect.Effect {
		delete(effect.Effect, k)
	}
	effect.ClearBuff = 0
	for k := range effect.TargetEffect {
		delete(effect.TargetEffect, k)
	}
}

func addBuffById(fightData *t.FightData, entity *t.FightEntity, logEffect *t.FightLogEffect, buffId int, check bool) *t.FightBuff {
	buffConf := g.GBuffConfig[buffId]
	if check && !checkAddBuff(entity, buffConf.Type) {
		newFightLogEffectResult(logEffect, feedbackImmuneAction, 0, -1)
		return nil
	}

	baseConf := g.GBuffBaseConfig[buffConf.Type]
	effect := buffConf.Effect
	//是否可叠加
	if baseConf.Superposed != 1 {
		for _, buff := range getBuffs(entity.Buffs, buffConf.Type, false) {
			if buff.Round >= effect.Round {
				return nil
			}
			newFightLogBuffRound(logEffect, buff.Guid, buff.BuffId, -1)
			//如果已触发，清除buff效果
			if buff.IsTrigger {
				triggerBuff(fightData, entity, buff, logEffect, false)
			} else {
				delete(entity.Buffs, buff.Guid)
			}
			break
		}
	}

	fightData.BuffIndex++
	buff := &t.FightBuff{
		Guid:       fightData.BuffIndex,
		BuffId:     buffId,
		Type:       buffConf.Type,
		GainType:   buffConf.GainType,
		Point:      effect.Trigger,
		TotalRound: effect.Round,
		Round:      effect.Round,
		ValueType:  effect.Type,
		Value:      effect.Value,
		Index:      fightData.BuffIndex,
	}
	entity.Buffs[buff.Guid] = buff
	if buffConf.Immediate == 1 {
		buff.IsTrigger = true
		triggerBuff(fightData, entity, buff, logEffect, true)
		buff.Round--
		buffChangeHp(fightData, entity, buff, logEffect)
	}

	newFightLogBuffRound(logEffect, buff.Guid, buff.BuffId, buff.Round)

	return buff
}

func checkBuff(buffs map[int]*t.FightBuff, buffType int) bool {
	for _, buff := range buffs {
		if buff.IsTrigger && buff.Type == buffType {
			return true
		}
	}

	return false
}

func checkAddBuff(entity *t.FightEntity, buffType int) bool {
	//有免疫此类型的效果 或 免疫buff
	if _, ok := entity.ImmuneBuff[buffType]; ok {
		return false
	}

	return !checkBuff(entity.Buffs, buffImmune)
}

func calcGainTypeCount(buffs map[int]*t.FightBuff, gainType int) int {
	var count int
	for _, buff := range buffs {
		if buff.GainType == gainType {
			count++
		}
	}

	return count
}

func clearBuffs(fightData *t.FightData, entity *t.FightEntity, logEffect *t.FightLogEffect, buffGuid, buffType, gainType int) int {
	count := len(entity.Buffs)
	for guid, buff := range entity.Buffs {
		if buff.Guid == buffGuid {
			continue
		}
		if buffType > 0 && buff.Type != buffType {
			continue
		}
		if gainType > 0 && buff.GainType != gainType {
			continue
		}
		newFightLogBuffRound(logEffect, guid, buff.BuffId, -1)
		if buff.IsTrigger {
			//如果已触发，清除buff效果
			triggerBuff(fightData, entity, buff, logEffect, false)
		}
		delete(entity.Buffs, guid)
	}

	return len(entity.Buffs) - count
}

func randClearBuffByGainType(fightData *t.FightData, entity *t.FightEntity, logEffect *t.FightLogEffect, gainType, count int) int {
	buffs := make([]*t.FightBuff, 0)
	old := len(entity.Buffs)
	for _, buff := range entity.Buffs {
		if buff.GainType == gainType {
			buffs = append(buffs, buff)
		}
	}
	if len(buffs) > count {
		for i, buff := range base.RandSliceN(count, buffs) {
			buffs[i] = buff.(*t.FightBuff)
		}
		buffs = buffs[0:count]
	}

	for _, buff := range buffs {
		if _, ok := entity.Buffs[buff.Guid]; !ok {
			continue
		}
		newFightLogBuffRound(logEffect, buff.Guid, buff.BuffId, -1)
		//如果已触发，清除buff效果
		if buff.IsTrigger {
			triggerBuff(fightData, entity, buff, logEffect, false)
		} else {
			delete(entity.Buffs, buff.Guid)
		}
	}

	return len(entity.Buffs) - old
}

func randGainBuff(fightData *t.FightData, entity *t.FightEntity, logEffect *t.FightLogEffect, gainType, count int) {
	//是否有免疫buff
	if checkBuff(entity.Buffs, buffImmune) {
		newFightLogEffectResult(logEffect, feedbackImmuneAction, 0, -1)
		return
	}
	buffIds := make([]int, 0)
	for _, conf := range buffGainTypes[gainType] {
		if _, ok := entity.ImmuneBuff[conf.Type]; !ok {
			buffIds = append(buffIds, conf.Id)
		}
	}

	delIds := make(map[int]bool)
	var success int
	for len(buffIds) > 0 {
		for _, v := range base.RandSliceN(count-success, buffIds) {
			buff := addBuffById(fightData, entity, logEffect, v.(int), false)
			if buff != nil {
				success++
				if success >= count || (buff.Type == buffImmune && buff.IsTrigger) {
					return
				}
			}
			delIds[buff.BuffId] = true
		}
		for i := len(buffIds) - 1; i >= 0; i-- {
			if _, ok := delIds[buffIds[i]]; ok {
				buffIds = append(buffIds[0:i], buffIds[i+1:]...)
			}
		}
	}
}

func changeBuffRound(fightData *t.FightData, entity *t.FightEntity, logEffect *t.FightLogEffect, round, buffType, gainType int) {
	for _, buff := range entity.Buffs {
		if buffType > 0 && buff.Type != buffType {
			continue
		}
		if gainType > 0 && buff.GainType != gainType {
			continue
		}

		buff.Round = int(math.Max(-1, float64(buff.Round+round)))
		newFightLogBuffRound(logEffect, buff.Guid, buff.BuffId, buff.Round)
		if buff.Round < 0 {
			if buff.IsTrigger {
				triggerBuff(fightData, entity, buff, logEffect, false)
			} else {
				delete(entity.Buffs, buff.Guid)
			}
		}
	}
}

func getBuffs(buffs map[int]*t.FightBuff, buffType int, isTrigger bool) []*t.FightBuff {
	var items []*t.FightBuff
	for _, buff := range buffs {
		if buffType > 0 && buff.Type != buffType {
			continue
		}

		if isTrigger && !buff.IsTrigger {
			continue
		}

		if items == nil {
			items = make([]*t.FightBuff, 0)
		}

		items = append(items, buff)
	}

	if len(items) == 0 {
		return items
	}

	//护盾
	if buffType == buffShield {
		sort.SliceStable(items, func(i, j int) bool {
			buffi, buffj := items[i], items[j]
			if buffi.Round < buffj.Round {
				return true
			} else if buffi.Round > buffj.Round {
				return false
			}

			return buffi.Index < buffj.Index
		})
	} else {
		sort.SliceStable(items, func(i, j int) bool {
			buffi, buffj := items[i], items[j]
			return buffi.Index < buffj.Index
		})
	}

	return items
}

func getOriginalTarget(fightData *t.FightData, entity *t.FightEntity, targetType int) []*t.FightEntity {
	index := entity.LordIndex
	targets := make([]*t.FightEntity, 0)
	switch targetType {
	case selTargetPartner:
		for _, target := range fightData.Entities {
			if target.LordIndex == index {
				targets = append(targets, target)
			}
		}
	case selTargetEnemy:
		for _, target := range fightData.Entities {
			if target.LordIndex != index {
				targets = append(targets, target)
			}
		}
	case selTargetAll:
		for _, target := range fightData.Entities {
			targets = append(targets, target)
		}
	case selTargetDeathPartner:
		for _, target := range fightData.RawEntities {
			if target.IsDead && target.LordIndex == index {
				targets = append(targets, target)
			}
		}
	}

	return targets
}

func sortTargetSpec(targets []*t.FightEntity, targetSpec int) []*t.FightEntity {
	switch targetSpec {
	//当前生命最高
	case targetSpecMaxHp:
		sort.SliceStable(targets, func(i, j int) bool {
			return targets[i].Attrs[c.AttrHp] > targets[j].Attrs[c.AttrHp]
		})
	//当前生命最低
	case targetSpecMinHp:
		sort.SliceStable(targets, func(i, j int) bool {
			return targets[i].Attrs[c.AttrHp] < targets[j].Attrs[c.AttrHp]
		})
	//速度最高
	case targetSpecMaxSpeed:
		sort.SliceStable(targets, func(i, j int) bool {
			return targets[i].Attrs[c.AttrSpeed] > targets[j].Attrs[c.AttrSpeed]
		})
	//攻击最高
	case targetSpecMaxAttack:
		sort.SliceStable(targets, func(i, j int) bool {
			targeti, targetj := targets[i], targets[j]
			return (targeti.Attrs[c.AttrMinAttack] + targeti.Attrs[c.AttrMaxAttack]) > (targetj.Attrs[c.AttrMinAttack] + targetj.Attrs[c.AttrMaxAttack])
		})
	//防御最高
	case targetSpecMaxDefense:
		sort.SliceStable(targets, func(i, j int) bool {
			return targets[i].Attrs[c.AttrDefense] > targets[j].Attrs[c.AttrDefense]
		})
	//防御最低
	case targetSpecMinDefense:
		sort.SliceStable(targets, func(i, j int) bool {
			return targets[i].Attrs[c.AttrDefense] < targets[j].Attrs[c.AttrDefense]
		})
	}

	return targets
}

func filterTargetByPreCondition(targets []*t.FightEntity, fightData *t.FightData, perCond, perCondParam int) []*t.FightEntity {
	switch perCond {
	//目标拥有XX标签时
	case preCondFeature:
		for i := len(targets) - 1; i >= 0; i-- {
			target := targets[i]
			if _, ok := target.Feature[perCondParam]; !ok {
				targets = append(targets[0:i], targets[i+1:]...)
			}
		}
	//目标拥有XXBUFF时
	case preCondBuff:
		for i := len(targets) - 1; i >= 0; i-- {
			target := targets[i]
			if !checkBuff(target.Buffs, perCondParam) {
				targets = append(targets[0:i], targets[i+1:]...)
			}
		}
	//目标生命>=
	case preCondHpGe:
		ratio := float64(perCondParam) / 10000
		for i := len(targets) - 1; i >= 0; i-- {
			target := targets[i]
			if target.Attrs[c.AttrHp] < target.RawAttrs[c.AttrHp]*ratio {
				targets = append(targets[0:i], targets[i+1:]...)
			}
		}
	//目标生命<
	case preCondHpLt:
		ratio := float64(perCondParam) / 10000
		for i := len(targets) - 1; i >= 0; i-- {
			target := targets[i]
			if target.Attrs[c.AttrHp] >= target.RawAttrs[c.AttrHp]*ratio {
				targets = append(targets[0:i], targets[i+1:]...)
			}
		}
	//拥有队友数量<=X个
	case preCondPartnerLe:
		for i := len(targets) - 1; i >= 0; i-- {
			target := targets[i]
			if len(getOriginalTarget(fightData, target, selTargetPartner))-1 > perCondParam {
				targets = append(targets[0:i], targets[i+1:]...)
			}
		}
	//拥有队友数量>X个
	case preCondPartnerGt:
		for i := len(targets) - 1; i >= 0; i-- {
			target := targets[i]
			if len(getOriginalTarget(fightData, target, selTargetPartner))-1 <= perCondParam {
				targets = append(targets[0:i], targets[i+1:]...)
			}
		}
	//消除效果数量<=X个
	case preCondClearEffectLe:
		for i := len(targets) - 1; i >= 0; i-- {
			target := targets[i]
			if target.Effect.ClearBuff > perCondParam {
				targets = append(targets[0:i], targets[i+1:]...)
			}
		}
	//消除效果数量>
	case preCondClearEffectGt:
		for i := len(targets) - 1; i >= 0; i-- {
			target := targets[i]
			if target.Effect.ClearBuff <= perCondParam {
				targets = append(targets[0:i], targets[i+1:]...)
			}
		}
	//目标拥有大于等于N个buff时
	case preCondNBuff:
		for i := len(targets) - 1; i >= 0; i-- {
			target := targets[i]
			if len(target.Buffs) < perCondParam {
				targets = append(targets[0:i], targets[i+1:]...)
			}
		}
	//目标拥有大于等于N个debuff时
	case preCondNDebuff:
		for i := len(targets) - 1; i >= 0; i-- {
			target := targets[i]
			if calcGainTypeCount(target.Buffs, deBuff) < perCondParam {
				targets = append(targets[0:i], targets[i+1:]...)
			}
		}
	}
	return targets
}

func selectTarget(fightData *t.FightData, entity, targetEntity *t.FightEntity,
	target, effect, targetParam, targetSpec int) []*t.FightEntity {
	var targets []*t.FightEntity
	switch target {
	//随机敌方X个目标
	case targetREnemyN:
		targets = getOriginalTarget(fightData, entity, selTargetEnemy)
		if len(targets) > targetParam {
			if targetSpec > 0 {
				targets = sortTargetSpec(targets, targetSpec)[0:targetParam]
			} else {
				for i, v := range base.RandSliceN(targetParam, targets) {
					targets[i] = v.(*t.FightEntity)
				}
				targets = targets[0:targetParam]
			}
		}
	//随机选择一个攻击目标和与其距离X的所有敌人
	case targetRAttackOneX:
		targets = getOriginalTarget(fightData, entity, selTargetEnemy)
		if len(targets) > 1 {
			var selTarget *t.FightEntity
			if targetSpec > 0 {
				targets = sortTargetSpec(targets, targetSpec)
				selTarget = targets[0]
			} else {
				selTarget = base.RandSliceN(1, targets)[0].(*t.FightEntity)
			}
			for i := len(targets) - 1; i >= 0; i-- {
				v := targets[i]
				if int(math.Abs(float64(selTarget.Pos-v.Pos))) > targetParam {
					targets = append(targets[0:i], targets[i+1:]...)
				}
			}
		}
	//随机我方X个目标
	case targetRPartnerN:
		targets = getOriginalTarget(fightData, entity, selTargetPartner)
		if len(targets) > targetParam {
			if targetSpec > 0 {
				targets = sortTargetSpec(targets, targetSpec)[0:targetParam]
			} else {
				for i, v := range base.RandSliceN(targetParam, targets) {
					targets[i] = v.(*t.FightEntity)
				}
				targets = targets[0:targetParam]
			}
		}
	//自己
	case targetSelf:
		if entity.HeroPos > 0 && (!entity.IsDead || effect == sEffectReliveTarget) {
			targets = []*t.FightEntity{entity}
		}
	//默认目标
	case targetDefTarget:
		if targetEntity != nil && targetEntity.HeroPos > 0 && (!targetEntity.IsDead || effect == sEffectReliveTarget) {
			targets = []*t.FightEntity{targetEntity}
		}
	//随机x个目标(包括敌我)
	case targetRandomN:
		targets = getOriginalTarget(fightData, entity, selTargetAll)
		if len(targets) > targetParam {
			if targetSpec > 0 {
				targets = sortTargetSpec(targets, targetSpec)[0:targetParam]
			} else {
				for i, v := range base.RandSliceN(targetParam, targets) {
					targets[i] = v.(*t.FightEntity)
				}
				targets = targets[0:targetParam]
			}
		}
	//随机已方已经死亡的N个目标
	case targetDeathPartnerN:
		targets = getOriginalTarget(fightData, entity, selTargetDeathPartner)
		if len(targets) > targetParam {
			if targetSpec > 0 {
				targets = sortTargetSpec(targets, targetSpec)[0:targetParam]
			} else {
				for i, v := range base.RandSliceN(targetParam, targets) {
					targets[i] = v.(*t.FightEntity)
				}
				targets = targets[0:targetParam]
			}
		}
	//选择与默认目标距离为X的所有队友(不包括默认目标)
	case targetDefTargetPartnerN:
		if targetEntity != nil {
			targets = getOriginalTarget(fightData, entity, selTargetPartner)
			for i := len(targets) - 1; i >= 0; i-- {
				v := targets[i]
				if v.Pos == targetEntity.Pos || int(math.Abs(float64(targetEntity.Pos-v.Pos))) > targetParam {
					targets = append(targets[0:i], targets[i+1:]...)
				}
			}
		}
	}

	return targets
}

func calcAttack(fightData *t.FightData, entity *t.FightEntity) float64 {
	var (
		effects   = entity.Effect.Effect
		attackImp int
	)
	if v, ok := effects[sEffectEnemyAttackImp]; ok {
		attackImp = len(getOriginalTarget(fightData, entity, selTargetEnemy)) * v
	}

	if v, ok := effects[sEffectPartnerAttackImp]; ok {
		attackImp += (len(getOriginalTarget(fightData, entity, selTargetPartner)) - 1) * v
	}

	attackImp += effects[sEffectAttackPct]

	attrs := entity.Attrs
	if attackImp == 0 {
		return float64(base.Rand(int(attrs[c.AttrMinAttack]), int(attrs[c.AttrMaxAttack])))
	}

	ratio := 1 + float64(attackImp)/10000
	return float64(base.Rand(int(attrs[c.AttrMinAttack]*ratio), int(attrs[c.AttrMaxAttack]*ratio)))
}

func calcDefense(fightData *t.FightData, entity *t.FightEntity) float64 {
	var (
		effects    = entity.Effect.Effect
		defenseImp int
	)
	if v, ok := effects[sEffectEnemyDefenseImp]; ok {
		defenseImp = len(getOriginalTarget(fightData, entity, selTargetEnemy)) * v
		delete(effects, sEffectEnemyDefenseImp)
	}

	if v, ok := effects[sEffectPartnerDefenseImp]; ok {
		defenseImp += (len(getOriginalTarget(fightData, entity, selTargetPartner)) - 1) * v
		delete(effects, sEffectPartnerDefenseImp)
	}

	if v, ok := effects[sEffectGainDefenseImprove]; ok {
		defenseImp += calcGainTypeCount(entity.Buffs, gainBuff) * v
		delete(effects, sEffectGainDefenseImprove)
	}

	defenseImp += effects[sEffectDefensePct]
	attrs := entity.Attrs
	if defenseImp == 0 {
		return attrs[c.AttrDefense]
	}

	return float64(int(attrs[c.AttrDefense] * (1 + float64(defenseImp)/10000)))
}

func newFightLogDead(fightData *t.FightData, entity int) {
	fightData.Logs = append(fightData.Logs, &t.FightLog{
		Entity: entity,
		Type:   logTypeDead,
	})
}

func newFightLogBuff(fightData *t.FightData, entity int) *t.FightLogEffect {
	fLog := &t.FightLog{
		Entity: entity,
		Type:   logTypeBuff,
		Logs: &t.FightLogEffect{
			Rounds:  make([]*t.FightBuffRound, 0),
			Results: make([]*t.FightLogEfffectResult, 0),
		},
	}

	fightData.Logs = append(fightData.Logs, fLog)

	return fLog.Logs.(*t.FightLogEffect)
}

func newFightLogActionFinish(fightData *t.FightData, entity int) {
	fightData.Logs = append(fightData.Logs, &t.FightLog{
		Entity: entity,
		Type:   logTypeActionFinish,
	})
}

func newFightLogSkillSelTarget(fightData *t.FightData, entity, skillId, index int) *t.FightLogSkill {
	logSkill := &t.FightLogSkill{
		Skill:   skillId,
		Index:   index,
		Type:    actionSelTarget,
		Effects: make([]int, 0),
	}
	fightData.Logs = append(fightData.Logs, &t.FightLog{
		Entity: entity,
		Type:   logTypeSkill,
		Logs:   logSkill,
	})

	return logSkill
}

func newFightLogSkillAction(fightData *t.FightData, entity, skillId, index int) *t.FightLogSkill {
	logSkill := &t.FightLogSkill{
		Skill:   skillId,
		Index:   index,
		Type:    actionAction,
		Effects: make([]*t.FightLogSkillEffectAction, 0),
	}
	fightData.Logs = append(fightData.Logs, &t.FightLog{
		Entity: entity,
		Type:   logTypeSkill,
		Logs:   logSkill,
	})

	return logSkill
}

func newFightLogEffectAction(logSkill *t.FightLogSkill, entity int, find bool) *t.FightLogSkillEffectAction {
	effects := logSkill.Effects.([]*t.FightLogSkillEffectAction)
	if find {
		for _, action := range effects {
			if action.Entity == entity {
				return action
			}
		}
	}

	action := &t.FightLogSkillEffectAction{
		Entity: entity,
		Effect: &t.FightLogEffect{
			Rounds:  make([]*t.FightBuffRound, 0),
			Results: make([]*t.FightLogEfffectResult, 0),
		},
	}
	logSkill.Effects = append(effects, action)
	return action
}

func newFightLogEffectActionResult(logSkill *t.FightLogSkill, entity, feedback, value int) {
	action := newFightLogEffectAction(logSkill, entity, false)
	newFightLogEffectResult(action.Effect, feedback, value, -1)
}

func newFightLogBuffRound(logEffect *t.FightLogEffect, guid, id, round int) {
	logEffect.Rounds = append(logEffect.Rounds, &t.FightBuffRound{Guid: guid, Id: id, Round: round})
}

func newFightLogEffectResult(logEffect *t.FightLogEffect, feedback, value, buffId int) {
	if feedback == feedbackSuckBloodAction {
		for _, result := range logEffect.Results {
			if result.Type == feedbackSuckBloodAction {
				result.Value += value
				return
			}
		}
	}

	logEffect.Results = append(logEffect.Results, &t.FightLogEfffectResult{
		Type:   feedback,
		Value:  value,
		BuffId: buffId,
	})
}

func onSendFightLogs(actor *t.Actor, fightData *t.FightData) {
	writer := pack.AllocPack(proto.Fight, proto.FightSLogs, float64(fightData.Guid), fightData.Round, int16(len(fightData.Logs)))
	for _, fLog := range fightData.Logs {
		pack.Write(writer, int16(fLog.Entity), fLog.Type)
		switch fLog.Type {
		case logTypeBuff:
			logBuff := fLog.Logs.(*t.FightLogEffect)
			pack.Write(writer, int16(len(logBuff.Rounds)))
			for _, round := range logBuff.Rounds {
				pack.Write(writer, round.Guid, round.Id, int16(round.Round))
			}
			pack.Write(writer, int16(len(logBuff.Results)))
			for _, result := range logBuff.Results {
				pack.Write(writer, int16(result.Type), result.Value, result.BuffId)
			}
		case logTypeSkill:
			logSkill := fLog.Logs.(*t.FightLogSkill)
			pack.Write(writer, logSkill.Skill, byte(logSkill.Index), byte(logSkill.Type))
			switch logSkill.Type {
			case actionSelTarget:
				effects := logSkill.Effects.([]int)
				pack.Write(writer, int16(len(effects)))
				for _, v := range effects {
					pack.Write(writer, int16(v))
				}
			case actionAction:
				effects := logSkill.Effects.([]*t.FightLogSkillEffectAction)
				pack.Write(writer, int16(len(effects)))
				for _, effect := range effects {
					pack.Write(writer, int16(effect.Entity))
					pack.Write(writer, int16(len(effect.Effect.Rounds)))
					for _, round := range effect.Effect.Rounds {
						pack.Write(writer, round.Guid, round.Id, int16(round.Round))
					}
					pack.Write(writer, int16(len(effect.Effect.Results)))
					for _, result := range effect.Effect.Results {
						pack.Write(writer, int16(result.Type), result.Value, result.BuffId)
					}
				}
			}
		}
	}
	actor.ReplyWriter(writer)

	fightData.Logs = make([]*t.FightLog, 0)
}

func OnFightClear(actor *t.Actor, fightData *t.FightData) {
	data := actor.GetFightData()
	if data == nil || data != fightData || data.RealResult != 0 {
		return
	}
	//log.Infof("OnFightClear: actor:%d,type:%d", actor.ActorId, fightData.Type)

	fightData.RealResult = fightData.FightResult

	if handle, ok := actorFightClearHandles[fightData.Type]; ok {
		handle(actor, fightData)
	}
	writer := pack.AllocPack(proto.Fight, proto.FightSResult,
		float64(fightData.Guid), fightData.Type, fightData.RealResult, int16(len(fightData.Awards)))
	for _, award := range fightData.Awards {
		pack.Write(writer, award.Type, award.Id, award.Count)
	}
	pack.Write(writer, fightData.ClearArgs...)
	actor.ReplyWriter(writer)

	//actor.SendTips(fmt.Sprintf("战斗结束，%d，回合:%d，耗时：%v", fightData.Guid, fightData.Round, time.Since(fightData.StartTime)))
}

func getFightAwards(actor *t.Actor, skip int) {
	fightData := actor.GetFightData()
	if fightData == nil || fightData.RealResult == 0 {
		return
	}
	//log.Infof("getFightAwards: actor:%d,type:%d", actor.ActorId, fightData.Type)
	if handle, ok := actorFightAwardHandles[fightData.Type]; ok {
		handle(actor, fightData)
	}
	if len(fightData.Awards) > 0 {
		bag.PutAwards2Bag(actor, fightData.Awards, false, true, c.ASNoAction, fmt.Sprintf("fight_%d", fightData.Type))
	}

	writer := pack.AllocPack(proto.Fight, proto.FightSGetAwards, fightData.Type, skip, 0)
	if skip == 0 {
		pack.Write(writer, int16(len(fightData.Entities)))
		for _, entity := range fightData.Entities {
			pack.Write(writer, entity.Pos, entity.Attrs[c.AttrHp])
		}
	}
	actor.ReplyWriter(writer)

	dynamicData := actor.GetDynamicData()
	dynamicData.FightData = nil
}

func onGetFightAwards(actor *t.Actor, reader *bytes.Reader) {
	var skip int
	pack.Read(reader, &skip)
	getFightAwards(actor, skip)
}

func onGiveupFight(actor *t.Actor, reader *bytes.Reader) {
	fightData := actor.GetFightData()
	if fightData == nil {
		return
	}

	if _, ok := g.GFightSkipConfig[fightData.Type]; ok {
		return
	}

	dynamicData := actor.GetDynamicData()
	dynamicData.FightData = nil

	timer.StopTimer(actor, fmt.Sprintf("triggerFighting_%d", fightData.Type))
	timer.StopTimer(actor, fmt.Sprintf("OnFightClear_%d", fightData.Type))

	if handle, ok := actorFightGiveupHandles[fightData.Type]; ok {
		handle(actor, fightData)
	}

	writer := pack.AllocPack(proto.Fight, proto.FightSGiveup, float64(fightData.Guid), int16(len(fightData.Entities)))
	for _, entity := range fightData.Entities {
		pack.Write(writer, entity.Pos, entity.Attrs[c.AttrHp])
	}
	actor.ReplyWriter(writer)
}

func onNextRound(actor *t.Actor, reader *bytes.Reader) {
	fightData := actor.GetFightData()
	if fightData == nil || fightData.FightResult != 0 {
		return
	}
	_, ok := actorFightHpSyncHandles[fightData.Type]
	if ok {
		roundFighting(actor, fightData)
	}
}

func onActorLogout(actor *t.Actor) {
	fightData := actor.GetFightData()
	if fightData == nil {
		return
	}
	if fightData.RealResult != 0 {
		getFightAwards(actor, 0)
	} else if handle, ok := actorFightGiveupHandles[fightData.Type]; ok {
		handle(actor, fightData)
	}
}
