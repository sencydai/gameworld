package typedefine

type LordBaseConfig struct {
	RenameCost int
}

type LordConfig struct {
	Camp   int
	Sex    int
	Head   int
	Frame  int
	Chat   int
	Job    int
	Heros  map[int]int
	Model  int
	Awards map[int]Award
}

type LordLevelConfig struct {
	Level   int
	NeedExp int
	Attrs   map[int]Attr
}

type ChangeJobConfig struct {
	Job       int
	NeedLevel int
	Level     int
}

type LordEquipConfig struct {
	Id        int
	Type      int
	Name      string
	Pos       int
	Rarity    int
	LimitTime int
	Attrs     map[int]Attr
}

type LordEquipStrengConfig struct {
	Stage int
	Level int
	Cost  map[int]AwardItem
	Attrs map[int]map[int]Attr
}

type LordHeadConfig struct {
	Id int
}

type LordFrameConfig struct {
	Id int
}

type LordChatConfig struct {
	Id int
}

type ItemConfig struct {
	Id        int
	Name      string
	Type      int
	SubType   int
	Rarity    int
	Worth     int
	OpenKey   int
	AwardType int
	Award     map[int]AwardRandom
	ComTarget *struct {
		Type int
		Id   int
	}
	ComCount int
}

type ItemGroupConfig struct {
	Id         int
	Index      int
	RandomType int
	Ratio      int
	Award      Award
}

type VirtualCurrencyConfig struct {
	VituralId int
	RealId    int
	AttrType  int
}

type TalentConfig struct {
	Talent    int
	Job       int
	PerTalent int
	Cost      int
	TotalCost int
}

type TalentLevelConfig struct {
	Talent  int
	Level   int
	Cost    int
	SkillId int
}

type LordSkillConfig struct {
	Id        int
	Index     int
	SkillId   int
	LordLevel int
	Level     int
}

type LordSkillLevelConfig struct {
	Id            int
	Index         int
	Level         int
	Fight         int
	EffectParam   map[int]int
	EffectExParam map[int]int
}

type LordSkillCostConfig struct {
	Level   int
	Consume map[int]AwardItem
}

type GuardLevelConfig struct {
	Level   int
	NeedExp int
	ModelId int
	Attrs   map[int]Attr
}

type GuardModelConfig struct {
	Id    int
	Level int
	Model int
}

type HeroBaseConfig struct {
	BaseScaleRatio map[int]Attr
	LevelExp       int
	LevelMult      int
	AttrWorthRatio map[int]Attr
	MaxFightHero   int
	MaxAssistHero  int
	StageBase      int
	Rebuild        int
}

type HeroConfig struct {
	Id         int
	Icon       string
	Model      int
	RawHero    int
	Feature    map[int]int
	CommSkill  int
	Skills     map[int]int
	JobLevel   int
	PowerRatio int
}

type HeroRawConfig struct {
	RawHero   int
	Race      int
	Rarity    int
	Attrs     map[int]Attr
	AttrRatio map[int]Attr
	Piece     int
}

type HeroRarityConfig struct {
	Rarity     int
	LevelRatio int
	StageRatio int
	ExpIndex   int
	Worth      int
}

type HeroLevelConfig struct {
	ExpIndex int
	Level    int
	Exp      int
}

type HeroStageConfig struct {
	Stage     int
	NeedLevel int
	NeedGold  int
	NeedBook  int
	NeedHero  int
	AttrRatio int
	JobLevel  int
}

type HeroStageTalentConfig struct {
	RawHero int
	Stage   int
	Attrs   map[int]Attr
}

type HeroChangeJobConfig struct {
	RawHero  int
	JobLevel int
	Heros    map[int]int
}

type HeroChangeJobCostConfig struct {
	Rarity      int
	Level       int
	UpgradeCost map[int]AwardItem
	ChangeCost  map[int]AwardItem
}

type HeroFetterConfig struct {
	Id        int
	Index     int
	Type      int
	Condition map[int]int
	Attr      map[int]Attr
}

type HeroOpenConfig struct {
	Id       int
	Pos      int
	Building int
	VipLevel int
}

type MonsterConfig struct {
	Id        int
	Hero      int
	Icon      string
	Model     int
	Race      int
	Feature   map[int]int
	CommSkill int
	Skills    map[int]int
	AttrCom   map[int]Attr
	AttrRate  map[int]Attr
}

type MonsterTemplateConfig struct {
	Id    int
	Level int
	Attr  map[int]Attr
}

type MonsterLordConfig struct {
	Id          int
	Name        string
	Model       int
	LordActive  map[int]int
	LordPassive map[int]int
	Equips      map[int]int
	GuardModel  map[int]int
}

type MonsterLordTemplateConfig struct {
	Id    int
	Level int
	Attr  map[int]Attr
}

type MainFubenBaseConfig struct {
	MaxOfflineTime int
	MaxEventTime   int
	OfflineRound   struct {
		Min int
		Max int
	}
	QfTime    int
	QfAdd     int
	QfConsume map[int]struct {
		Times int
		Count int
	}
	QtAttr     map[int]Attr
	QtAttrTime int
}

type MainFubenConfig struct {
	Id            int
	Lord          MonsterLord
	Monster       map[int]MonsterHero
	Awards        map[int]AwardRandom
	FixAwards     map[int]Award
	MonsterLevel  int
	OfflineAwards map[int]Award
	OfflineCycle  int
	OfflineRandom map[int]AwardRandom
	HookTime      int
	HookFightTime int
}

type BuffConfig struct {
	Id        int
	Type      int
	GainType  int
	Immediate int
	Effect    struct {
		Trigger int
		Round   int
		Type    int
		Value   int
	}
}

type BuffBaseConfig struct {
	Type       int
	Superposed int
}

type SkillConfig struct {
	Skill       int
	RepeatCount int
	Attr        map[int]Attr
	SkillRandom int
	SkillType   int
	Trigger     int
}

type SkillEffectConfig struct {
	Skill         int
	Index         int
	Random        int
	ActionCD      int
	Target        int
	TargetParam   int
	TargetSpec    int
	PerCond       int
	PerCondParam  int
	ReservePre    int
	Effect        int
	EffectParam   float64
	EffectExParam float64
}

type RaceConfig struct {
	Race  int
	Ratio int
}

type EquipBaseConfig struct {
	MaxCount       int
	StrengExp      int
	StrengMult     int
	StrengConst    int
	InitLevel      int
	AttrWorthRatio map[int]Attr
	AttrScale      map[int]Attr
	ExStrengLimit  int
	Recast         int
}

type EquipConfig struct {
	Id        int
	Name      string
	Pos       int
	Suit      int
	Rarity    int
	Streng    int
	Piece     int
	AttrRatio map[int]Attr
}

type EquipPosConfig struct {
	Pos       int
	AttrRatio map[int]Attr
}

type EquipRarityConfig struct {
	Rarity      int
	StrengRatio map[int]Attr
	Resolve     AwardItem
}

type EquipStrengConfig struct {
	Id        int
	Level     int
	Consume   map[int]AwardItem
	RetAwards map[int]AwardItem
}

type EquipSuitConfig struct {
	Id    int
	Attr2 map[int]Attr
	Attr3 map[int]Attr
	Attr4 map[int]Attr
}

type EquipStrengMasterConfig struct {
	Id    int
	Level int
	Attr  map[int]Attr
}

type ArtifactBaseConfig struct {
	ResolveWorth map[int]int
}

type ArtifactConfig struct {
	Id        int
	Rarity    int
	AttrTypes map[int]int
}

type ArtifactRarityConfig struct {
	Rarity    int
	Stage     int
	BaseAttrs map[int]Attr
	RariAttrs map[int]Attr
}

type ArtifactStrengConfig struct {
	Stage       int
	Level       int
	Consume     int
	StrengAttrs map[int]Attr
}

type BuildingConfig struct {
	Id         int
	Level      int
	OutputType int
}

type BuildingLevelConfig struct {
	Id        int
	Level     int
	LevelType int
	NeedLevel int
	Consume   map[int]AwardItem
	Output    map[int]AwardItem
	MaxHeros  int
}

type ChatBaseConfig struct {
	ChannelCD    map[int]int
	MaxChar      int
	MaxFriend    int
	MaxBlacklist int
}

type SystemNoticeConfig struct {
	Id         int
	SubId      int
	Index      int
	Type       int
	SpecId     map[int]int
	Rarity     int
	Channel    string
	Roll       int
	NoticeType int
	Content    string
}

type SystemTimingNoticeConfig struct {
	Id          int
	TimeType    int
	TimeSubType int
	Time        interface{} //TimeWeek | TimeOpenServer | TimeFixed
	Interval    int
	Channel     string
	Roll        int
	NoticeType  int
	Content     string
}

type SystemConfig struct {
	Type int
}

type SystemOpenConfig struct {
	Type         int
	Level        int
	Fight        int
	TaskId       int
	TimeType     int
	TimeSubType  int
	Time         interface{} //TimeWeek | TimeOpenServer | TimeFixed
	HTimeType    int
	HTimeSubType int
	HTime        interface{} //TimeWeek | TimeOpenServer | TimeFixed
}

type FightSkipConfig struct {
	Type     int
	Level    int
	VipLevel int
}

type WorldBossBaseConfig struct {
	MaxCount int
	Recover  int
	Cd       int
}

type WorldBossConfig struct {
	Id      int
	Lord    MonsterLord
	Monster map[int]MonsterHero
	Cd      int
	Level   int
	Drops   map[int]Award
	Kill    map[int]Award
}

type WorldBossRankConfig struct {
	Id     int
	Index  int
	Upper  int
	Awards map[int]Award
}
