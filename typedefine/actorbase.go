package typedefine

import (
	c "github.com/sencydai/gameworld/constdefine"
)

type ActorBaseData struct {
	Bag *ActorBaseBagData //背包

	Job     *ActorBaseJobData               //职业等级
	Decor   map[int]*ActorBaseDecorData     //领主装饰	类型：data
	LEquip  *ActorBaseLordEquipData         //领主装备
	LTalent *ActorBaseLordTalentData        //天赋
	LSkill  map[int]*ActorBaseLordSkillData //领主技能 位置: data
	Guard   *ActorBaseGuardData             //亲卫

	FightHeros  map[int]int         //出战英雄 位置:guid
	AssistHeros map[int]int         //助战英雄 位置:guid
	Equips      map[int]map[int]int //装备 出战位置: 装备部位: guid
	Artifacts   map[int]int         //神器 位置：guid
}

type ActorBaseBagData struct {
	Currency    map[int]int         //货币id:数量
	AccCurrency map[int]int         //历史累积货币 id:数量
	Items       map[int]map[int]int //道具 类型:id:数量

	Hero     *ActorBaseBagHeroData     //英雄
	Equip    *ActorBaseBagEquipData    //装备
	Artifact *ActorBaseBagArtifactData //神器
}

//ActorBaseJobData 转职
type ActorBaseJobData struct {
	Jobs  map[int]byte
	Level int
}

//ActorBaseDecorData 装饰
type ActorBaseDecorData struct {
	Id     int
	Unlock map[int]byte
}

//ActorBaseLordEquipData 领主装备
type ActorBaseLordEquipData struct {
	Pos    int                                //当前强化位置
	Equips map[int]*ActorBaseLordEquipPosData //位置:位置信息
}

type ActorBaseLordEquipPosData struct {
	Stage int //等阶
	Level int //等级
}

//ActorBaseLordTalentData 天赋
type ActorBaseLordTalentData struct {
	Count int         //已消耗总点数
	Learn map[int]int //已学习 id : level
}

//ActorBaseLordSkillData 领主技能
type ActorBaseLordSkillData struct {
	Id    int
	Index int
	Level int
}

//ActorBaseGuardData 亲卫
type ActorBaseGuardData struct {
	Level int
	Model int
}

//ActorBaseVipData vip
type ActorBaseVipData struct {
	Level int
	Exp   int
	Get   map[int]byte
}

type ActorBaseBagHeroData struct {
	MaxId int
	Heros map[int]*HeroStaticData
}

type HeroStaticData struct {
	Guid    int
	PosType c.LordEquipPos
	Pos     int
	Map     int

	Id    int
	Level int
	Exp   int
	Stage int

	Power int
}

type ActorBaseBagEquipData struct {
	MaxId  int
	Equips map[int]*EquipStaticData
}

type EquipStaticData struct {
	Guid  int
	Pos   int
	Id    int
	Level int
}

type ActorBaseBagArtifactData struct {
	MaxId int
	Artis map[int]*ArtifactStaticData
}

type ArtifactStaticData struct {
	Guid  int
	Pos   int
	Id    int
	Stage int
	Level int
}

func (actor *Actor) GetBagData() *ActorBaseBagData {
	baseData := actor.GetBaseData()
	if baseData.Bag == nil {
		baseData.Bag = &ActorBaseBagData{}
	}

	return baseData.Bag
}

func (actor *Actor) GetJobData() *ActorBaseJobData {
	baseData := actor.GetBaseData()
	if baseData.Job == nil {
		baseData.Job = &ActorBaseJobData{Jobs: make(map[int]byte)}
	}

	return baseData.Job
}

func (actor *Actor) GetDecorData() map[int]*ActorBaseDecorData {
	baseData := actor.GetBaseData()
	if baseData.Decor == nil {
		baseData.Decor = make(map[int]*ActorBaseDecorData)
	}

	return baseData.Decor
}

func (actor *Actor) GetLordHead() int {
	decor := actor.GetDecorData()
	head, ok := decor[c.LDTHead]
	if !ok {
		return -1
	}

	return head.Id
}

func (actor *Actor) GetLordFrame() int {
	decor := actor.GetDecorData()
	frame, ok := decor[c.LDTFrame]
	if !ok {
		return -1
	}

	return frame.Id
}

func (actor *Actor) GetLordChat() int {
	decor := actor.GetDecorData()
	chat, ok := decor[c.LDTChat]
	if !ok {
		return -1
	}

	return chat.Id
}

func (actor *Actor) GetLordEquipData() *ActorBaseLordEquipData {
	baseData := actor.GetBaseData()
	if baseData.LEquip == nil {
		baseData.LEquip = &ActorBaseLordEquipData{
			Pos:    1,
			Equips: make(map[int]*ActorBaseLordEquipPosData),
		}
		for i := 1; i <= c.LEPMax; i++ {
			baseData.LEquip.Equips[i] = &ActorBaseLordEquipPosData{
				Stage: 0,
				Level: 0,
			}
		}
	}

	return baseData.LEquip
}

func (actor *Actor) GetLordTalentData() *ActorBaseLordTalentData {
	baseData := actor.GetBaseData()
	if baseData.LTalent == nil {
		baseData.LTalent = &ActorBaseLordTalentData{
			Learn: make(map[int]int),
		}
	}

	return baseData.LTalent
}

func (actor *Actor) GetLordSkillData() map[int]*ActorBaseLordSkillData {
	baseData := actor.GetBaseData()
	if baseData.LSkill == nil {
		baseData.LSkill = make(map[int]*ActorBaseLordSkillData)
	}

	return baseData.LSkill
}

func (actor *Actor) GetGuardData() *ActorBaseGuardData {
	baseData := actor.GetBaseData()
	if baseData.Guard == nil {
		baseData.Guard = &ActorBaseGuardData{}
	}

	return baseData.Guard
}

func (actor *Actor) GetBagHeroData() *ActorBaseBagHeroData {
	bagData := actor.GetBagData()
	if bagData.Hero == nil {
		bagData.Hero = &ActorBaseBagHeroData{Heros: make(map[int]*HeroStaticData)}
	}

	return bagData.Hero
}

func (actor *Actor) GetHeroStaticData(guid int) *HeroStaticData {
	heroData := actor.GetBagHeroData()
	return heroData.Heros[guid]
}

func (actor *Actor) GetBagEquipData() *ActorBaseBagEquipData {
	bagData := actor.GetBagData()
	if bagData.Equip == nil {
		bagData.Equip = &ActorBaseBagEquipData{Equips: make(map[int]*EquipStaticData)}
	}
	return bagData.Equip
}

func (actor *Actor) GetEquipStaticData(guid int) *EquipStaticData {
	equipData := actor.GetBagEquipData()
	return equipData.Equips[guid]
}

func (actor *Actor) GetBagArtifactData() *ActorBaseBagArtifactData {
	bagData := actor.GetBagData()
	if bagData.Artifact == nil {
		bagData.Artifact = &ActorBaseBagArtifactData{
			Artis: make(map[int]*ArtifactStaticData),
		}
	}
	return bagData.Artifact
}

func (actor *Actor) GetArtifactStaticData(guid int) *ArtifactStaticData {
	artiData := actor.GetBagArtifactData()
	return artiData.Artis[guid]
}

func (actor *Actor) GetFightHeros() map[int]int {
	baseData := actor.GetBaseData()
	if baseData.FightHeros == nil {
		baseData.FightHeros = make(map[int]int)
	}

	return baseData.FightHeros
}

func (actor *Actor) GetAssistHeros() map[int]int {
	baseData := actor.GetBaseData()
	if baseData.AssistHeros == nil {
		baseData.AssistHeros = make(map[int]int)
	}

	return baseData.AssistHeros
}

func (actor *Actor) GetEquipDatas() map[int]map[int]int {
	baseData := actor.GetBaseData()
	if baseData.Equips == nil {
		baseData.Equips = make(map[int]map[int]int)
	}
	return baseData.Equips
}

func (actor *Actor) GetHeroEquips(heroPos int) map[int]int {
	equipData := actor.GetEquipDatas()
	equips, ok := equipData[heroPos]
	if !ok {
		equips = make(map[int]int)
		equipData[heroPos] = equips
	}

	return equips
}

func (actor *Actor) GetArtifactDatas() map[int]int {
	baseData := actor.GetBaseData()
	if baseData.Artifacts == nil {
		baseData.Artifacts = make(map[int]int)
	}

	return baseData.Artifacts
}
