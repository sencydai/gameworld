package typedefine

import (
	"time"
)

type FightData struct {
	Guid        int64
	Type        int
	FightResult int
	RealResult  int

	CbArgs    []interface{}
	Awards    map[int]Award
	ClearArgs []interface{}

	Data        []*FightLord
	Order       []int
	Entities    map[int]*FightEntity
	RawEntities map[int]*FightEntity

	Round     int16
	BuffIndex int
	Logs      []*FightLog

	StartTime time.Time
}

type FightLord struct {
	Pos    int
	Model  int
	Name   string
	Gmodel []int
	Power  int
	Equips map[int]int

	PassSkills          []int       //初动技能id
	ActiveSkills        map[int]int //主动技能 pos:skillId
	SkillEffectParams   map[int]map[int]int
	SkillEffectExParams map[int]map[int]int

	Entity  *FightEntity //领主属性
	AttrSum int

	Heros map[int]*FightHeroTemplate //pos : FightHeroTemplate
}

type FightHeroTemplate struct {
	Model     int
	CommSkill int
	Skills    map[int][]int
	RaceRatio float64
	Feature   map[int]bool
}

//实体
type FightEntity struct {
	Pos       int
	LordIndex int
	HeroPos   int

	RaceRatio float64

	RawAttrs map[int]float64
	Attrs    map[int]float64
	Feature  map[int]bool

	SkillCount  map[int]int
	IsAction    bool               //是否正在行动
	Buffs       map[int]*FightBuff //guid : FightBuff
	BuffEffects map[int]bool
	ImmuneBuff  map[int]bool //免疫buff类型
	IsDead      bool
	ReAction    bool //再进行一轮行动

	Effect            *FightSkillEffect   //流程生效(含一次生效)
	WholeEffect       map[int]int         //全程生效的技能效果
	WholeTargetEffect map[int]map[int]int //全程生效的对目标的技能效果
}

//技能效果
type FightSkillEffect struct {
	Effect       map[int]int         //类型 : 值
	ClearBuff    int                 //消除buff数量
	TargetEffect map[int]map[int]int //目标: 类型 : 值
}

type FightBuff struct {
	Guid       int
	BuffId     int
	IsTrigger  bool //是否已触发
	Type       int  //buff类型
	GainType   int  //增益类型
	Point      int  //触发点
	TotalRound int  //总回合
	Round      int  //剩余回合
	ValueType  int
	Value      int //值
	Index      int
}

type FightLogEfffectResult struct {
	Type   int
	Value  int
	BuffId int
}

type FightBuffRound struct {
	Guid  int
	Id    int
	Round int
}

type FightLogSkillEffectAction struct {
	Entity int
	Effect *FightLogEffect
}

//技能日志
type FightLogSkill struct {
	Skill   int
	Index   int
	Type    int
	Effects interface{} // []int || []*FightLogSkillEffectAction
}

//效果日志
type FightLogEffect struct {
	Rounds  []*FightBuffRound
	Results []*FightLogEfffectResult
}

type FightLog struct {
	Entity int
	Type   byte
	Logs   interface{} // *FightLogEffect || *FightLogSkill
}
