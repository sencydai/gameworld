package constdefine

const (
	MaxActorNameLen = 12 //角色名称最大长度
)

//奖励类型
type AwardType = int

const (
	ATItem      AwardType = 1 //物品
	ATItemGroup AwardType = 2 //物品组
	//ATLEquip    AwardType = 3 //领主装备
	ATLHead     AwardType = 4 //领主头像
	ATLFrame    AwardType = 5 //领主头像边框
	ATLChat     AwardType = 6 //领主聊天框
	ATHero      AwardType = 7 //英雄
	ATEquip     AwardType = 8 //装备
	ATArtifact  AwardType = 9 //神器
)

//奖励来源
type AwardSource = int

const (
	ASCommon     AwardSource = 0 //通用弹框
	ASFlyFont    AwardSource = 1 //飘字
	ASNoAction   AwardSource = 2 //不处理
	ASHookEvent  AwardSource = 3 //挂机事件(副本挂机)
	ASHookOutput AwardSource = 4 //挂机产出(系统定时产出)
)

//英雄删除来源
type HeroDeleteSource = int

const (
	HDSCommon   HeroDeleteSource = 0
	HDSDissmiss HeroDeleteSource = 1
	HDSRebuild  HeroDeleteSource = 3
)

//装备删除来源
type EquipDeleteSource = int

const (
	EDSResolve EquipDeleteSource = 1
	EDSRecast  EquipDeleteSource = 2
)

//随机类型
type RandomType = int

const (
	RTProbability RandomType = 1 //概率
	RTWeight      RandomType = 2 //权值
)

//物品类型
type ItemType = int

const (
	ITCurrency ItemType = 1 //货币
	ITMaterial ItemType = 2 //材料
	ITBox      ItemType = 3 //宝箱
	ITPiece    ItemType = 4 //碎片

	ITMax ItemType = 4
)

//货币类型
type CurrencyType = int

const (
	CTGold         CurrencyType = 1  //金币
	CTWood         CurrencyType = 2  //木材
	CTDiamond      CurrencyType = 3  //钻石
	CTLExp         CurrencyType = 4  //领主经验
	CTScienceP     CurrencyType = 5  //科技点
	CTHonor        CurrencyType = 6  //荣誉
	CTGuildContri  CurrencyType = 7  //公会贡献(给个人)
	CTGuildFund    CurrencyType = 8  //公会资金(给公会)
	CTDanHonor     CurrencyType = 9  //段位荣誉
	CTGuardExp     CurrencyType = 10 //亲卫经验
	CTAreaPrestige CurrencyType = 11 //地区声望
	CTAreaMaterial CurrencyType = 12 //地区物资
	CTOrigiStone   CurrencyType = 13 //原石

	CTMax CurrencyType = 13
)

//材料类型
type MaterialType = int

const (
	MTExpBook         MaterialType = 1  //经验之书
	MTKnowledgeBook   MaterialType = 2  //知识之书
	MTTransferBook    MaterialType = 3  //转职之书
	MTImprovedStone   MaterialType = 4  //强化石
	MTPalaeoidPiece   MaterialType = 5  //上古碎片
	MTKey             MaterialType = 6  //宝箱钥匙
	MTSoulEssence     MaterialType = 7  //灵魂精华
	MTEquipEssence    MaterialType = 8  //装备精华
	MTArtifactEssence MaterialType = 9  //神器精华
	MTSkillPoint      MaterialType = 10 //领主技能点
)

//领主装饰类型
type LordDecorType = int

const (
	LDTHead  LordDecorType = 1 //头像
	LDTFrame LordDecorType = 2 //边框
	LDTChat  LordDecorType = 3 //聊天框

	LDTMax LordDecorType = 3
)

//领主装备部位
type LordEquipPos = int

const (
	LEPMax LordEquipPos = 8
)

//英雄位置类型
type HeroPosType = int

const (
	HPTFight      HeroPosType = 1 //出战
	HPTAssist     HeroPosType = 2 //助战
	HPTExpedition HeroPosType = 3 //远征
)

//系统类型
const (
	SystemChat             = 1   //聊天
	SystemAltar            = 2   //祭坛
	SystemGuard            = 3   //亲卫
	SystemSoul             = 4   //英灵
	SystemLordEquip        = 5   //领主装备
	SystemLordSkill        = 6   //领主技能
	SystemHeroEquip        = 7   //英雄装备
	SystemHeroArti         = 8   //英雄神器
	SystemStore            = 9   //商店
	SystemQuickFight       = 10  //快速战斗
	SystemLadderChallenge  = 101 //天梯挑战赛
	SystemDanChallenge     = 102 //段位晋级赛
	SystemEliteFuben       = 103 //精英副本
	SystemMine             = 104 //矿产争夺战
	SystemWorldBoss        = 105 //世界boss
	SystemLegendBattle     = 109 //传奇战役
	SystemMainFuben        = 110 //主线副本
	SystemHookEvent        = 111 //野外挂机杀怪
	SystemDailyFuben       = 112 //每日副本（总的）
	SystemEndlessTower     = 113 //无尽塔
	SystemAreaEvent        = 115 //地区事件
	SystemHonorRoad        = 116 //荣耀之路
	SystemCourageChallenge = 117 //勇气试炼
	SystemRobberFuben      = 118 //强盗来袭
)

//建筑
const (
	BuildingMainCity    = 1 //主城
	BuildingMine        = 2 //矿井
	BuildingShop        = 3 //商店
	BuildingHouse       = 4 //房子
	BuildingSmithy      = 5 //铁匠铺
	BuildingHeroAltar   = 6 //英雄祭坛
	BuildingTagBuilding = 7 //标识性建筑
)

//排行榜
const (
	RankLevel      = "rankdb_level"
	RankLevelCount = 100
	RankPower      = "rankdb_fight"
	RankPowerCount = 100
)

//时间类型
const (
	TimeTypeUnlimit    = 0
	TimeTypeWeek       = 1
	TimeTypeOpenServer = 2
	TimeTypeFixed      = 3
)

//时间子类型
const (
	TimeSubDay  = 1
	TimeSubLast = 2
)

//时间状态
const (
	TimeStatusUnlimit = 0
	TimeStatusExpire  = 1
	TimeStatusOutside = 2
	TimeStatusInRange = 3
)
