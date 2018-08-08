package protocol

const (
	Chat          byte = 9   //聊天
	Mail          byte = 41  //邮件
	Lord          byte = 50  //领主
	Hero          byte = 51  //英雄
	Fight         byte = 52  //战斗
	Building      byte = 53  //建筑
	Award         byte = 54  //奖励
	Store         byte = 55  //商店
	RandExplore   byte = 56  //随机探索
	Athletics     byte = 57  //竞技
	Mine          byte = 58  //矿洞
	Rank          byte = 59  //排行榜
	Altar         byte = 60  //祭坛
	Task          byte = 61  //任务
	Base          byte = 63  //基础
	HookContEvent byte = 54  //挂机连续事件
	Bag           byte = 179 //背包
	Equip         byte = 180 //装备
	Fuben         byte = 181 //副本
	Guild         byte = 183 //公会
	Friend        byte = 184 //好友
	Activity      byte = 186 //活动
	System        byte = 255 // 系统
)

//基础
const (
	BaseCUpdateClientData byte = 5 //更新前端数据

	BaseSSyncTime       byte = 1 //同步服务器时间
	BaseSOpenSystemList byte = 2 //系统开放列表
	BaseSOpenSystemSync byte = 3 //系统开放状态同步
	BaseSClientData     byte = 4 //同步前端数据
	BaseSNewDay         byte = 6 //newday
)

//聊天
const (
	ChatCSendChatMsg = 1

	ChatSSendChatMsg = 1
	ChatSTips        = 2 //系统tips
	ChatSSysMsg      = 4 //系统消息
)

// 战斗
const (
	FightCGetAwards byte = 2  // 领取奖励
	FightCGiveup    byte = 8  //放弃战斗
	FightCNextRound byte = 11 //下一回合

	FightSResult           byte = 1  //战斗结果
	FightSGetAwards        byte = 2  // 领取奖励
	FightSInit             byte = 3  //战斗初始化
	FightSLogs             byte = 4  //战斗流程
	FightSUpdateDamageRank byte = 7  //伤害榜更新
	FightSGiveup           byte = 8  //放弃战斗
	FightSHpSync           byte = 10 //血量同步
)

// 副本
const (
	FubenCLoginMainFuben      byte = 2  //登陆主线副本
	FubenCUpdateWorldBossInfo byte = 21 //更新boss信息
	FubenCWorldBossChallenge  byte = 23 //挑战世界boss

	FubenSMainFuben                   byte = 1  //主线副本
	FubenSWorldBossInit               byte = 20 //世界boss初始化
	FubenSUpdateWorldBossInfo         byte = 21 //更新boss信息
	FubenSWorldBossSyncChallengeCount byte = 22 //世界boss挑战次数同步
)

// 系统
const (
	SystemCLogin       byte = 1 // 登陆
	SystemCCreateActor byte = 2 // 创建角色
	SystemCActorList   byte = 4 // 查询角色列表
	SystemCLoginGame   byte = 5 //登陆游戏
	SystemCRandomName  byte = 6 // 随机名称

	SystemSLogin       byte = 1 //登陆
	SystemSCreateActor byte = 2 // 创建角色
	SystemSActorLists  byte = 4 //查询角色列表
	SystemSLoginGame   byte = 5 //登陆游戏
	SystemSRandomName  byte = 6 //随机名称
)

//背包
const (
	BagCOpenBox byte = 12 //开启宝箱
	BagCCompose byte = 15 //合成

	BagSItemInit       byte = 1  //物品初始化
	BagSItemDelete     byte = 2  //物品删除
	BagSHeroInit       byte = 3  //英雄初始化
	BagSHeroUpdate     byte = 4  //英雄修改
	BagSHeroDelete     byte = 5  //英雄删除
	BagSEquipInit      byte = 6  //装备初始化
	BagSEquipUpdate    byte = 7  //装备修改
	BagSEquipDelete    byte = 8  //装备删除
	BagSArtiInit       byte = 9  //神器初始化
	BagSArtiUpdate     byte = 10 //神器修改
	BagSArtiDelete     byte = 11 //神器删除
	BagSCurrencyInit   byte = 13 //货币初始化
	BagSCurrencyDelete byte = 14 //货币删除
	BagSAddAwards      byte = 17 //添加奖励
)

//领主
const (
	LordCDecorChange      byte = 5  //更换装饰
	LordCEquipChange      byte = 9  //更换装备
	LordCEquipStreng      byte = 10 //装备强化
	LordCChangeJob        byte = 12 //领主转职
	LordCTalentLearn      byte = 14 //学习天赋
	LordCTalentUpgrade    byte = 15 //升级天赋
	LordCGetVipAwards     byte = 20 //领取VIP奖励
	LordCLookupLord       byte = 21 //查看领主信息
	LordCLookupHero       byte = 22 //查看英雄
	LordCRandomName       byte = 23 //请求随机名称
	LordCChangeName       byte = 24 //改名
	LordCSkillStage       byte = 30 //技能进阶
	LordCSkillUpgrade     byte = 31 //技能升级
	LordCSkillExchangePos byte = 32 //技能位置

	LordSCreateActor   byte = 1  //
	LordSBaseInfo      byte = 2  //基础数据
	LordSDecorInit     byte = 3  //装饰初始化
	LordSDecorUnlock   byte = 4  //装饰解锁
	LordSDecorChange   byte = 5  //更换装饰
	LordSEquipInit     byte = 6  //装备初始化
	LordSEquipUnlock   byte = 7  //装备解锁
	LordSEquipTimeout  byte = 8  //装备过期
	LordSChangeEquip   byte = 9  //更换装备
	LordSStrengEquip   byte = 10 //装备强化
	LordSUpgrade       byte = 11 //领主升级
	LordSChangeJob     byte = 12 //领主转职
	LordSTalentInit    byte = 13 //天赋初始化
	LordSTalentLearn   byte = 14 //学习天赋
	LordSTalentUpgrade byte = 15 //升级天赋
	LordSRefreshPower  byte = 18 //领主战力更新
	LordSLookupLord    byte = 21 //查看领主信息
	LordSRandomName    byte = 23 //请求随机名称
	LordSSkillInit     byte = 28 //领主技能初始化
	LordSNewSkill      byte = 29 //新技能开放
	LordSSkillStage    byte = 30 //技能进阶
	LordSSkillUpgrade  byte = 31 //技能升级
)

//英雄
const (
	HeroCSetArmyHeroPos byte = 2  //设置部队英雄位置
	HeroCOneKeyUpgrade  byte = 3  //一键升级
	HeroCUpgradeStage   byte = 4  //英雄升阶
	HeroCChangeJob      byte = 5  //英雄转职
	HeroCWearEquip      byte = 7  //穿着装备
	HeroCStrengEquip    byte = 8  //强化装备
	HeroCResolveEquip   byte = 9  //装备分解
	HeroCRecastEquip    byte = 10 //装备重铸
	HeroCWearArti       byte = 12 //穿着神器
	HeroCStrengArti     byte = 13 //神器强化
	HeroCHeroDismiss    byte = 18 //英雄遣散
	HeroCHeroRebuild    byte = 19 //英雄重修
	HeroCResolveArti    byte = 21 //神器分解

	HeroSArmyInit     byte = 1  //部队初始化
	HeroSEquipInit    byte = 6  //装备初始化
	HeroSArtifactInit byte = 11 //神器初始化
	HeroSRefreshPower byte = 20 //英雄战力更新
)

//建筑
const (
	BuildingCUpgrade byte = 3 //升级建筑

	BuildingSInit   byte = 1 //初始化
	BuildingSUpdate byte = 2 //新增或升级
)

//排行榜
const (
	RankCRankData byte = 1 //请求排行榜数据

	RankSRankData byte = 1 //请求排行榜数据
)

//跨服
const (
	CrossLookLordReq = 1
	CrossLookLordRes = 2
)
