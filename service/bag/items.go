package bag

import (
	"fmt"
	"sort"

	. "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service"
	. "github.com/sencydai/gameworld/typedefine"
)

var (
	singleItemIds map[int]int
	multItemIds   map[int][]int
)

func init() {
	service.RegConfigLoadFinish(onConfigLoadFinish)
}

func onConfigLoadFinish() {
	singleItemIds = make(map[int]int)
	multItemIds = make(map[int][]int)
}

func GetCurrencyBag(actor *Actor) map[int]int {
	bag := GetBag(actor)
	if bag.Currency == nil {
		bag.Currency = make(map[int]int)
	}
	return bag.Currency
}

func GetAccCurrencyBag(actor *Actor) map[int]int {
	bag := GetBag(actor)
	if bag.AccCurrency == nil {
		bag.AccCurrency = make(map[int]int)
	}
	return bag.AccCurrency
}

func GetItemsBag(actor *Actor) map[int]map[int]int {
	bag := GetBag(actor)
	if bag.Items == nil {
		bag.Items = make(map[int]map[int]int)
	}
	return bag.Items
}

func GetItemTypeBag(actor *Actor, it ItemType) map[int]int {
	if it < ITCurrency || it > ITMax {
		return nil
	}
	switch it {
	case ITCurrency:
		return GetCurrencyBag(actor)
	default:
		bag := GetItemsBag(actor)
		items, ok := bag[it]
		if !ok {
			items = make(map[int]int)
			bag[it] = items
		}
		return items
	}

}

func GetItemCount(actor *Actor, id int) int {
	items := GetItemTypeBag(actor, gconfig.GItemConfig[id].Type)
	return items[id]
}

func GetAccCurrencyCount(actor *Actor, ct CurrencyType) int {
	bag := GetAccCurrencyBag(actor)
	return bag[ct]
}

func GetItemsCountBySubType(actor *Actor, it ItemType, subType int, isSort bool) []AwardItem {
	items := GetItemTypeBag(actor, it)
	if len(items) == 0 {
		return nil
	}
	subItems := make([]AwardItem, 0)
	for id, count := range items {
		conf := gconfig.GItemConfig[id]
		if conf.SubType == subType {
			subItems = append(subItems, AwardItem{Id: id, Count: count})
		}
	}

	if isSort {
		sort.SliceStable(subItems, func(i, j int) bool {
			confI := gconfig.GItemConfig[i]
			confJ := gconfig.GItemConfig[j]
			return confI.Worth > confJ.Worth
		})
	}

	return subItems
}

func GetItemSingleId(it ItemType, subType int) int {
	mask := (it << 16) + subType
	if id, ok := singleItemIds[mask]; ok {
		return id
	}

	for id, conf := range gconfig.GItemConfig {
		if conf.Type == it && conf.SubType == subType {
			singleItemIds[mask] = id
			return id
		}
	}

	panic(fmt.Errorf("not find item single id: itemtype(%d),subtype(%d)", it, subType))
}

func GetItemMultIds(it ItemType, subType int) []int {
	mask := (it << 16) + subType
	if ids, ok := multItemIds[mask]; ok {
		return ids
	}

	ids := make([]int, 0)
	for id, conf := range gconfig.GItemConfig {
		if conf.Type == it && conf.SubType == subType {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		panic(fmt.Errorf("not find item mult ids: itemtype(%d),subtype(%d)", it, subType))
	}

	multItemIds[mask] = ids
	return ids
}

func parseVirtualCurrency(actor *Actor, id, count int) (int, int) {
	itemConf := gconfig.GItemConfig[id]
	if itemConf.Type != ITCurrency || itemConf.SubType <= CTMax {
		return id, count
	}
	virtualConf := gconfig.GVirtualCurrencyConfig[id]
	id = virtualConf.RealId
	if actor != nil {
		lordAttr := actor.GetLordAttr()
		if ratio, ok := lordAttr.Total[virtualConf.AttrType]; ok {
			count += count * ratio / 10000
		}
	}

	return id, count
}
