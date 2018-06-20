package bag

import (
	. "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/gconfig"
	. "github.com/sencydai/gameworld/typedefine"
)

func init() {

}

func GetBag(actor *Actor) *ActorBaseBagData {
	baseData := actor.GetBaseData()
	if baseData.Bag == nil {
		baseData.Bag = &ActorBaseBagData{}
	}
	return baseData.Bag
}

func SliceAwards(awards map[int]Award) []Award {
	items := make([]Award, len(awards))
	for i := 1; i <= len(awards); i++ {
		items[i] = awards[i]
	}
	return items
}

func MapAwards(awards []Award) map[int]Award {
	items := make(map[int]Award)
	for i, award := range awards {
		items[i+1] = award
	}
	return items
}

func FlushAwards(actor *Actor, awards map[int]Award) map[int]Award {
	items := make(map[int]map[int]int)
	for _, award := range awards {
		types, ok := items[award.Type]
		if !ok {
			types = make(map[int]int)
			items[award.Type] = types
		}
		if award.Type == ATItem && award.Id <= 1000 {
			id, count := parseVirtualCurrency(actor, award.Id, award.Count)
			types[id] += count
		} else {
			types[award.Id] += types[award.Count]
		}
	}

	awards = make(map[int]Award)
	for t, award := range items {
		for id, count := range award {
			awards[id] = Award{Type: t, Id: id, Count: count}
		}
	}
	return awards
}

func parseItemGroup(actor *Actor, id, count int) {
	groups := gconfig.GItemGroupConfig[id]
	//rType := groups[1].RandomType
	for i := 0; i < count; i++ {
		items := make(map[int]AwardRandom)
		for j, value := range groups {
			award := value.Award
			items[j] = AwardRandom{Type: award.Type, Id: award.Id, Count: award.Id, Rand: value.Ratio}
		}
	}
}
