package typedefine

//ActorDynamicAttrData 玩家属性
type ActorDynamicAttrData struct {
	Lord  *ActorDynamicLordAttrData
	Heros map[int]*ActorDynamicHeroAttrData
}

//ActorDynamicLordAttrData 领主属性
type ActorDynamicLordAttrData struct {
	Total map[int]int
}

func (actor *Actor) GetAttr() *ActorDynamicAttrData {
	data := actor.GetDynamicData()
	if data.Attr == nil {
		data.Attr = &ActorDynamicAttrData{Lord: &ActorDynamicLordAttrData{}, Heros: make(map[int]*ActorDynamicHeroAttrData)}
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
	return heros[guid]
}

//ActorDynamicHeroAttrData 英雄属性
type ActorDynamicHeroAttrData struct {
	Total map[int]int
}
