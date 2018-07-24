package fight

import (
	c "github.com/sencydai/gameworld/constdefine"
)

const (
	posRate        = 100 //站位
	maxRound       = 30  //回合数
	addAttackRatio = 0.1 //达到回合数每回合增加攻击万分比
)

//战斗胜负结果
const (
	Win  = 1
	Lose = 2
)

//战斗类型
const (
	MainFuben        = c.SystemMainFuben        //主线副本
	HookEvent        = c.SystemHookEvent        //野外挂机杀怪
	DailyFuben       = c.SystemDailyFuben       //每日副本
	EliteFuben       = c.SystemEliteFuben       //精英副本
	Mine             = c.SystemMine             //挖矿
	EndlessTower     = c.SystemEndlessTower     //无尽塔
	Legend           = c.SystemLegendBattle     //传奇战役
	DanChallenge     = c.SystemDanChallenge     //竞技段位晋级赛
	LadderChallenge  = c.SystemLadderChallenge  //天梯
	WorldBoss        = c.SystemWorldBoss        //世界boss
	AreaEvent        = c.SystemAreaEvent        //地区事件
	HonorRoad        = c.SystemHonorRoad        //荣耀之路
	CourageChallenge = c.SystemCourageChallenge //勇气试炼
	RobberFuben      = c.SystemRobberFuben      //强盗来袭
)

const (
	fixedAddtion   = 1 //固定值加成
	percentAddtion = 2 //百分比加成
)

//选择目标类型
const (
	selTargetPartner      = 1
	selTargetEnemy        = 2
	selTargetAll          = 3
	selTargetDeathPartner = 4
)

var (
	//行动流程
	actionFlow = []int{
		triggerAction,
		triggerAttack,
		triggerAttackFinish,
		triggerActionFinish,
	}

	buffHpTypes = map[int]bool{
		buffHpRecover:    true,
		buffConDamageMax: true,
		buffConDamage:    true,
		buffHpRecoverMax: true,
	}
)

//技能类型
const (
	activeSkill  = 1 //主动
	passiveSkill = 2 //被动
)

//技能触发点
const (
	triggerBegin           = 1  //战斗开始时
	triggerAction          = 2  //行动时
	triggerAttack          = 3  //攻击时
	triggerAttackFinish    = 4  //攻击完成
	triggerActionFinish    = 5  //行动结束
	triggerBeAttack        = 6  //被攻击时
	triggerBeAttackFinish  = 7  //被攻击后
	triggerPartnerBeAttack = 8  //队友被攻击时
	triggerSelfDodge       = 9  //自己闪避时
	triggerPartnerDodge    = 10 //队友闪避时
	triggerPartnerCrit     = 11 //队友暴击时
	triggerSelfCrit        = 12 //自己暴击时
	triggerSelfDamage      = 13 //自己受到伤害时
	triggerHitBack         = 14 //反击
	triggerSelfDead        = 15 //自己死亡时
	triggerPartnerDead     = 16 //队友死亡时
	triggerSkillDamage     = 17 //技能造成伤害时
	triggerClearBuff       = 18 //消除buff
	triggerGainBuff        = 19 //获得buff
	triggerTargetDead      = 20 //目标死亡时
)

//技能目标
const (
	targetREnemyN           = 1 //随机敌方X个目标
	targetRAttackOneX       = 2 //随机选择一个攻击目标和与其距离X的所有敌人
	targetRPartnerN         = 3 //随机我方X个目标
	targetSelf              = 4 //自己
	targetDefTarget         = 5 //默认目标
	targetRandomN           = 6 //随机x个目标(包括敌我)
	targetDeathPartnerN     = 7 //随机已方已经死亡的N个目标
	targetDefTargetPartnerN = 8 //选择与默认目标距离为X的所有队友(不包括默认目标)
)

//技能目标特定参数
const (
	targetSpecMaxHp      = 1 //当前生命最高
	targetSpecMinHp      = 2 //当前生命最低
	targetSpecMaxSpeed   = 3 //速度最高
	targetSpecMaxAttack  = 4 //攻击最高
	targetSpecMaxDefense = 5 //防御最高
	targetSpecMinDefense = 6 //防御最低
)

//技能前置条件
const (
	preCondFeature       = 1  //目标拥有XX标签时
	preCondBuff          = 2  //目标拥有XXBUFF时
	preCondHpGe          = 3  //目标生命>=
	preCondHpLt          = 4  //目标生命<
	preCondPartnerLe     = 5  //拥有队友数量<=X个
	preCondPartnerGt     = 6  //拥有队友数量>X个
	preCondClearEffectLe = 7  //消除效果数量<=X个
	preCondClearEffectGt = 8  //消除效果数量>
	preCondNBuff         = 9  //目标拥有大于等于N个buff时
	preCondNDebuff       = 10 //目标拥有大于等于N个debuff时
)

//buff类型
const (
	buffAttack             = 1  //攻击
	buffDefense            = 2  //防御
	buffSpeed              = 3  //速度
	buffCrit               = 4  //暴击
	buffHit                = 5  //命中
	buffShield             = 6  //护盾
	buffImmune             = 7  //免疫
	buffInvincible         = 8  //无敌
	buffHpRecover          = 9  //生命恢复(当前生命)
	buffConDamageMax       = 10 //持续掉血(最大生命)
	buffConDamage          = 11 //持续掉血(当前生命)
	buffSilence            = 12 //沉默
	buffDizzy              = 13 //眩晕
	buffUnRecover          = 14 //无法恢复
	buffDizzyAndDefenseSub = 15 //变羊（眩晕加破甲）
	buffHpRecoverMax       = 16 //生命恢复(生命上限)
	buffSuckBlood          = 17 //吸血
)

//buff增益类型
const (
	gainBuff = 1
	deBuff   = 2
)

//技能效果
const (
	sEffectDamage             = 1  //技能伤害
	sEffectMultlAttack        = 2  //攻击多次
	sEffectDebuffImprove      = 3  //目标每拥有一个减益BUFF伤害提高
	sEffectGainImprove        = 4  //目标每拥有一个增益BUFF伤害提高
	sEffectThorns             = 5  //反伤(全程有效)
	sEffectSuckBlood          = 6  //吸血(多次)
	sEffectHpRecover          = 7  //生命恢复（值）
	sEffectHpRecoverPct       = 8  //生命恢复（%）
	sEffectHitBack            = 9  //反击(全程有效)
	sEffectReduceDamage       = 10 //减少受到伤害（值）
	sEffectReduceDamagePct    = 11 //减少受到伤害（%）
	sEffectAddBuff            = 12 //添加某buff
	sEffectClearBuffType      = 13 //消除某类型buff
	sEffectImmuneBuff         = 14 //免疫某类型buff
	sEffectRandomGain         = 15 //获得随机增益BUFF
	sEffectRandomDebuff       = 16 //获得随机减益BUFF
	sEffectAddRound           = 17 //某类型BUFF持续时间增加
	sEffectReduceRound        = 18 //某类型BUFF持续时间减少
	sEffectGainAddRound       = 19 //所有增益BUFF持续时间增加
	sEffectGainReduceRound    = 20 //所有增益BUFF持续时间减少
	sEffectEnemyAttackImp     = 21 //每拥有一个敌人，攻击上升百分比
	sEffectEnemyDefenseImp    = 22 //每拥有一个敌人，防御上升百分比
	sEffectPartnerAttackImp   = 23 //每拥有一个友军，攻击上升百分比
	sEffectPartnerDefenseImp  = 24 //每拥有一个友军，防御上升百分比
	sEffectReliveTarget       = 25 //复活目标
	sEffectReAction           = 26 //重复一次行动流程
	sEffectSpecAttack         = 27 //特殊攻击(全程有效)
	sEffectChangeAttr         = 28 //修改某属性(固定值)
	sEffectChangeAttrPct      = 29 //修改某属性(百分比)
	sEffectWholeSuckBlood     = 30 //全程吸血(全程有效)
	sEffectAttackPct          = 31 //攻击百分比加成
	sEffectDefensePct         = 32 //防御百分比加成
	sEffectCritPct            = 33 //暴击加成
	sEffectMaxHpDamage        = 34 //最大生命值的百分比伤害
	sEffectHpDamage           = 35 //当前生命值的百分比伤害
	sEffectWholeIgnoreDefense = 36 //无视防御(全程有效)
	sEffectIgnoreDefense      = 37 //无视防御(一次生效)
	sEffectClearGainBuff      = 38 //随机消除目标N个增益buff
	sEffectClearDebuff        = 39 //随机消除目标N个减益buff
	sEffectRealDamage         = 40 //真实技能伤害
	sEffectGainDefenseImprove = 41 //目标每拥有一个增益BUFF防御提高
)

//战斗日志类型
const (
	logTypeActionFinish = 1 //行动结束
	logTypeBuff         = 2 //buff
	logTypeSkill        = 3 //技能
	logTypeDead         = 4 //死亡
)

//技能行动类型
const (
	actionSelTarget = 1
	actionAction    = 2
)

//战斗结果反馈
const (
	feedbackDamageImprove   = 1  //伤害提高
	feedbackThorns          = 2  //反伤
	feedbackSuckBlood       = 3  //吸血
	feedbackHpRecover       = 4  //生命恢复
	feedbackHitBack         = 5  //反击
	feedbackReduceDamage    = 6  //减少受到伤害
	feedbackAttackRise      = 7  //攻击上升
	feedbackDefenseRise     = 8  //防御上升
	feedbackInvincible      = 9  //无敌
	feedbackDodge           = 10 //闪避
	feedbackBuffHpAdd       = 11 //buff生命值增加
	feedbackBuffHpReduce    = 12 //buff生命值减少
	feedbackDamage          = 13 //伤害，生命值减少
	feedbackCrit            = 14 //暴击，生命值减少
	feedbackAddBuff         = 15 //添加buff
	feedbackClearBuff       = 16 //消除buff
	feedbackImmunebuff      = 17 //免疫buff
	feedbackBuffAddRound    = 18 //buff时间增加
	feedbackBuffReduceRound = 19 //buff时间减少
	feedbackthornsHpReduce  = 22 //反伤生命值减少
	//feedbackEntityHpSync = 23		  //实体血量同步
	feedbackShield          = 24 //护盾减伤
	feedbackHitBackAction   = 25 //反击起作用时
	feedbackUnRecoverAction = 26 //无法恢复起作用时
	feedbackImmuneAction    = 27 //免疫起作用时
	feedbackSuckBloodAction = 28 //吸血起作用时
	feedbackRelive          = 29 //复活
	feedbackReAction        = 30 //重复一次行动流程
	feedbackSpecAttack      = 31 //特殊攻击
	feedbackChangeAttr      = 32 //改变属性
	feedbackAttack          = 31 //攻击加成
	feedbackDefense         = 32 //防御加成
	feedbackCritPct         = 33 //暴击加成
	feedbackIgnoreDefense   = 34 //无视防御
)
