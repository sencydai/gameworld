package constdefine

//属性
const (
	AttrAttackBase       = 1  //基础攻击力
	AttrMinAttackBase    = 2  //基础最小攻击
	AttrMaxAttackBase    = 3  //基础最大攻击
	AttrDefenseBase      = 4  //基础护甲
	AttrHpBase           = 5  //基础生命
	AttrSpeedBase        = 6  //基础速度
	AttrExAttack         = 7  //额外固定值攻击
	AttrExDefense        = 8  //额外固定值护甲
	AttrExHp             = 9  //额外固定值生命
	AttrExSpeed          = 10 //额外固定值速度
	AttrExAttackPct      = 11 //额外百分比攻击(百分比)
	AttrExDefensePct     = 12 //额外百分比护甲(百分比)
	AttrExHpPct          = 13 //额外百分比生命(百分比)
	AttrCrit             = 14 //暴击率(百分比)
	AttrCritDefense      = 15 //暴击抵抗(百分比)
	AttrCritDamage       = 16 //暴击伤害(百分比)
	AttrHit              = 17 //命中(百分比)
	AttrDodge            = 18 //闪避(百分比)
	AttrDamageAddPct     = 19 //伤害加深(百分比)
	AttrDamageSubPct     = 20 //伤害减免(百分比)
	AttrSuckBlood        = 21 //吸血(百分比)
	AttrThorns           = 22 //反伤(百分比)
	AttrAttackCom        = 23 //攻击指挥固定值
	AttrDefenseCom       = 24 //防御指挥固定值
	AttrLordDamage       = 25 //领主伤害固定值
	AttrLordDamageSub    = 26 //领主减伤固定值
	AttrAttackComPct     = 27 //攻击指挥百分比(百分比)
	AttrDefenseComPct    = 28 //防御指挥百分比(百分比)
	AttrLordDamagePct    = 29 //领主伤害百分比(百分比)
	AttrLordDamageSubPct = 30 //领主减伤百分比(百分比)
	AttrNearDamageAdd    = 31 //近战英雄伤害加深(百分比)
	AttrLongDamageAdd    = 32 //远程英雄伤害加深(百分比)
	AttrPhysDamageAdd    = 33 //物理英雄伤害加深(百分比)
	AttrMagicDamageAdd   = 34 //魔法英雄伤害加深(百分比)
	AttrNearDamageSub    = 35 //近战英雄伤害减免(百分比)
	AttrLongDamageSub    = 36 //远程英雄伤害减免(百分比)
	AttrPhysDamageSub    = 37 //物理英雄伤害减免(百分比)
	AttrMagicDamageSub   = 38 //魔法英雄伤害减免(百分比)
	AttrDamageAdd        = 39 //伤害加深(固定值)
	AttrDamageSub        = 40 //伤害减免(固定值)

	AttrMax = 40

	AttrMinAttack = 41 //战斗最小攻击
	AttrMaxAttack = 42 //战斗最大攻击
	AttrDefense   = 43 //战斗护甲
	AttrHp        = 44 //战斗生命
	AttrSpeed     = 45 //战斗速度
)

//英雄特性
const (
	FeasureNear  = 1 //近战英雄
	FeasureLong  = 2 //远程英雄
	FeasurePhys  = 3 //物理英雄
	FeasureMagic = 4 //魔法英雄
)
