package constdefine

const (
	MaxActorNameLen = 12 //角色名称最大长度
)

//奖励类型
type AwardType = int

const (
	ATItem      AwardType = 1 //物品
	ATItemGroup AwardType = 2 //物品组
	ATLEquip    AwardType = 3 //领主装备
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

	CTMax CurrencyType = 14
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
