package typedefine

type ActorDynamicData struct {
	Attr      *ActorDynamicAttrData
	LordModel int //领主模型
	FightData *FightData
}

//ActorDynamicAttrData 玩家属性
type ActorDynamicAttrData struct {
	Lord  *ActorDynamicLordAttrData
	Heros map[int]*ActorDynamicHeroAttrData
}

//ActorDynamicLordAttrData 领主属性
type ActorDynamicLordAttrData struct {
	Total map[int]float64

	SkillPower int

	Level  map[int]float64
	Equip  map[int]map[int]float64
	Talent map[int]map[int]float64
}

//ActorDynamicHeroAttrData 英雄属性
type ActorDynamicHeroAttrData struct {
	Total map[int]float64

	Lord  map[int]float64
	Base  map[int]float64
	Level map[int]float64
	Stage map[int]float64
}

func (actor *Actor) GetAttr() *ActorDynamicAttrData {
	data := actor.GetDynamicData()
	if data.Attr == nil {
		data.Attr = &ActorDynamicAttrData{
			Lord: &ActorDynamicLordAttrData{
				make(map[int]float64),
				0,
				make(map[int]float64),
				make(map[int]map[int]float64),
				make(map[int]map[int]float64),
			},
			Heros: make(map[int]*ActorDynamicHeroAttrData),
		}
	}
	return data.Attr
}

func (actor *Actor) GetLordAttr() *ActorDynamicLordAttrData {
	data := actor.GetAttr()
	return data.Lord
}

func (actor *Actor) GetHerosAttr() map[int]*ActorDynamicHeroAttrData {
	data := actor.GetAttr()
	return data.Heros
}

func (actor *Actor) GetHeroAttr(guid int) *ActorDynamicHeroAttrData {
	heros := actor.GetHerosAttr()
	hero, ok := heros[guid]
	if !ok {
		hero = &ActorDynamicHeroAttrData{
			make(map[int]float64),
			make(map[int]float64),
			make(map[int]float64),
			make(map[int]float64),
			make(map[int]float64),
		}
		heros[guid] = hero
	}
	return hero
}

func (actor *Actor) GetLordModel() int {
	data := actor.GetDynamicData()
	return data.LordModel
}

func (actor *Actor) GetFightData() *FightData {
	data := actor.GetDynamicData()
	return data.FightData
}
