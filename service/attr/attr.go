package attr

import (
	"fmt"
	"math"
	"strconv"

	c "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/data"
	"github.com/sencydai/gameworld/service"
	t "github.com/sencydai/gameworld/typedefine"
)

func init() {
	service.RegGm("attr", onGmAttr)
}

func onGmAttr(values map[string]string) (int, string) {
	aid, _ := strconv.ParseFloat(values["actor"], 64)
	actor := data.GetActor(int64(aid))
	if actor == nil {
		return -1, fmt.Sprintf("not find actor %d", int64(aid))
	}

	pos, _ := strconv.Atoi(values["pos"])
	t, _ := strconv.Atoi(values["type"])
	value, _ := strconv.Atoi(values["value"])
	v := int(math.Max(0, float64(value)))

	if pos == 0 {
		lordAttr := actor.GetLordAttr()
		attrs := lordAttr.GM
		attrs[t] = float64(v)

		refreshLordTotalAttr(actor)
	} else {
		heros := actor.GetFightHeros()
		guid, ok := heros[pos]
		if !ok {
			return -1, "no hero in pos"
		}

		heroAttr := actor.GetHeroAttr(guid)
		heroAttr.GM[t] = float64(v)
		refreshHeroTotalAttr(actor, actor.GetHeroStaticData(guid))
	}

	return 0, "success"
}

//RefreshAttr 更新玩家属性
func RefreshAttr(actor *t.Actor) {
	//领主属性
	refreshLordAttr(actor)

	//英雄属性
	for _, guid := range actor.GetFightHeros() {
		RefreshHeroAttr(actor, actor.GetHeroStaticData(guid))
	}
}

//CalcEntityFightAttr 计算实体战斗属性
func CalcEntityFightAttr(totalAttr map[int]float64) {
	CalcMinAttack(totalAttr)
	CalcMaxAttack(totalAttr)
	CalcDefense(totalAttr)
	CalcHp(totalAttr)
	CalcSpeed(totalAttr)
}

//CalcMinAttack 战斗最小攻击
func CalcMinAttack(attr map[int]float64) {
	//最小攻击力 =（基础攻击 之和 + 基础最小 之和）*（100% + 额外百分比攻击 之和）+ 额外固定值攻击 之和
	value := (attr[c.AttrAttackBase]+attr[c.AttrMinAttackBase])*(1+attr[c.AttrExAttackPct]/10000) + attr[c.AttrExAttack]
	attr[c.AttrMinAttack] = float64(int(value))
}

//CalcMaxAttack 战斗最大攻击
func CalcMaxAttack(attr map[int]float64) {
	//最大攻击力 = （基础攻击 之和 + 基础最大 之和）*（100% + 额外百分比攻击 之和）+ 额外固定值攻击 之和
	value := (attr[c.AttrAttackBase]+attr[c.AttrMaxAttackBase])*(1+attr[c.AttrExAttackPct]/10000) + attr[c.AttrExAttack]
	attr[c.AttrMaxAttack] = float64(int(value))
}

//CalcDefense 战斗护甲
func CalcDefense(attr map[int]float64) {
	//护甲 = 基础护甲 之和 *（100% + 额外百分比护甲 之和）+ 额外固定值护甲 之和
	value := attr[c.AttrDefenseBase]*(1+attr[c.AttrExDefensePct]/10000) + attr[c.AttrExDefense]
	attr[c.AttrDefense] = float64(int(value))
}

//CalcHp 战斗生命
func CalcHp(attr map[int]float64) {
	//生命 = 基础生命 之和 *（100% + 额外百分比生命 之和）+ 额外固定值生命 之和
	value := attr[c.AttrHpBase]*(1+attr[c.AttrExHpPct]/10000) + attr[c.AttrExHp]
	attr[c.AttrHp] = float64(int(value))
}

//CalcSpeed 战斗速度
func CalcSpeed(attr map[int]float64) {
	//速度 = 基础速度 之和 + 额外速度 之和
	attr[c.AttrSpeed] = attr[c.AttrSpeedBase] + attr[c.AttrExSpeed]
}
