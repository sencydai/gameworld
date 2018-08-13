package bag

import (
	"fmt"
	"sort"
	"strings"

	c "github.com/sencydai/gameworld/constdefine"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
	"github.com/sencydai/gameworld/service"
	t "github.com/sencydai/gameworld/typedefine"

	"github.com/sencydai/gameworld/log"
)

var (
	singleItemIds map[int]int
	multItemIds   map[int][]int
)

func init() {
	service.RegConfigLoadFinish(onConfigLoadFinish)
}

func onConfigLoadFinish(isGameStart bool) {
	singleItemIds = make(map[int]int)
	multItemIds = make(map[int][]int)
}

func GetCurrencyBag(actor *t.Actor) map[int]int {
	bag := actor.GetBagData()
	if bag.Currency == nil {
		bag.Currency = make(map[int]int)
	}
	return bag.Currency
}

func GetAccCurrencyBag(actor *t.Actor) map[int]int {
	bag := actor.GetBagData()
	if bag.AccCurrency == nil {
		bag.AccCurrency = make(map[int]int)
	}
	return bag.AccCurrency
}

func GetItemsBag(actor *t.Actor) map[int]map[int]int {
	bag := actor.GetBagData()
	if bag.Items == nil {
		bag.Items = make(map[int]map[int]int)
	}
	return bag.Items
}

func GetItemTypeBag(actor *t.Actor, it c.ItemType) map[int]int {
	if it < c.ITCurrency || it > c.ITMax {
		return nil
	}
	switch it {
	case c.ITCurrency:
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

func GetItemCount(actor *t.Actor, id int) int {
	items := GetItemTypeBag(actor, g.GItemConfig[id].Type)
	return items[id]
}

func GetAccCurrencyCount(actor *t.Actor, ct c.CurrencyType) int {
	bag := GetAccCurrencyBag(actor)
	return bag[ct]
}

func GetItemsCountBySubType(actor *t.Actor, it c.ItemType, subType int, isSort bool) []t.AwardItem {
	items := GetItemTypeBag(actor, it)
	if len(items) == 0 {
		return nil
	}
	subItems := make([]t.AwardItem, 0)
	for id, count := range items {
		conf := g.GItemConfig[id]
		if conf.SubType == subType {
			subItems = append(subItems, t.AwardItem{Id: id, Count: count})
		}
	}

	if isSort {
		sort.SliceStable(subItems, func(i, j int) bool {
			confI := g.GItemConfig[i]
			confJ := g.GItemConfig[j]
			return confI.Worth > confJ.Worth
		})
	}

	return subItems
}

func GetItemSingleId(it c.ItemType, subType int) int {
	mask := (it << 16) + subType
	if id, ok := singleItemIds[mask]; ok {
		return id
	}

	for id, conf := range g.GItemConfig {
		if conf.Type == it && conf.SubType == subType {
			singleItemIds[mask] = id
			return id
		}
	}

	panic(fmt.Errorf("not find item single id: itemtype(%d),subtype(%d)", it, subType))
}

func GetItemMultIds(it c.ItemType, subType int) []int {
	mask := (it << 16) + subType
	if ids, ok := multItemIds[mask]; ok {
		return ids
	}

	ids := make([]int, 0)
	for id, conf := range g.GItemConfig {
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

func parseVirtualCurrency(actor *t.Actor, id, count int) (int, int) {
	itemConf := g.GItemConfig[id]
	if itemConf.Type != c.ITCurrency || itemConf.SubType <= c.CTMax {
		return id, count
	}
	virtualConf := g.GVirtualCurrencyConfig[id]
	id = virtualConf.RealId
	if actor != nil {
		lordAttr := actor.GetLordAttr()
		if ratio, ok := lordAttr.Total[virtualConf.AttrType]; ok {
			count += int(float64(count) * ratio / 10000)
		}
	}

	return id, count
}

func PutItem2Bag(actor *t.Actor, id, count int, source c.AwardSource, logText string) bool {
	return PutAward2Bag(actor, c.ATItem, id, count, false, source, logText)
}

func PutItems2Bag(actor *t.Actor, items map[int]t.AwardItem, source c.AwardSource, logText string) bool {
	awards := make(map[int]t.Award)
	for index, item := range items {
		awards[index] = t.Award{Type: c.ATItem, Id: item.Id, Count: item.Count}
	}

	return PutAwards2Bag(actor, awards, false, false, source, logText)
}

func PutItems2Bag2(actor *t.Actor, items map[int]int, source c.AwardSource, logText string) bool {
	awards := make(map[int]t.Award)
	for id, count := range items {
		awards[len(awards)+1] = t.Award{Type: c.ATItem, Id: id, Count: count}
	}

	return PutAwards2Bag(actor, awards, false, false, source, logText)
}

func PutItems2BagRatio(actor *t.Actor, items map[int]t.AwardItem, ratio float64, source c.AwardSource, logText string) bool {
	awards := make(map[int]t.Award)
	for index, item := range items {
		awards[index] = t.Award{Type: c.ATItem, Id: item.Id, Count: int(float64(item.Count) * ratio)}
	}

	return PutAwards2Bag(actor, awards, false, false, source, logText)
}

func CheckDeductItem(actor *t.Actor, id, count int) bool {
	total := GetItemCount(actor, id)
	return total >= count
}

func CheckDeductItems(actor *t.Actor, items map[int]t.AwardItem) bool {
	for _, item := range items {
		if !CheckDeductItem(actor, item.Id, item.Count) {
			return false
		}
	}
	return true
}

func CheckDeductItems2(actor *t.Actor, items map[int]int) bool {
	for id, count := range items {
		if !CheckDeductItem(actor, id, count) {
			return false
		}
	}
	return true
}

func DeductItem(actor *t.Actor, id, count int, check bool, logText string) (ok bool, realLeft int) {
	if check && !CheckDeductItem(actor, id, count) {
		return false, 0
	}

	itemConf := g.GItemConfig[id]
	items := GetItemTypeBag(actor, itemConf.Type)
	items[id] -= count
	realLeft = items[id]
	left := realLeft
	if realLeft <= 0 {
		left = 0
		delete(items, id)
	}

	if itemConf.Type == c.ITCurrency {
		writer := pack.AllocPack(
			proto.Bag,
			proto.BagSCurrencyDelete,
			id,
			left,
			-count,
		)
		actor.ReplyWriter(writer)
	} else {
		writer := pack.AllocPack(
			proto.Bag,
			proto.BagSItemDelete,
			itemConf.Type,
			id,
			left,
			-count,
		)
		actor.ReplyWriter(writer)
	}

	if len(logText) > 0 {
		log.Optf("DeductItem %s: actor(%d),id(%d),deduct(%d),left(%d)", logText, actor.ActorId, id, count, realLeft)
	}

	return true, realLeft
}

func DeductItems(actor *t.Actor, items map[int]t.AwardItem, check bool, logText string) bool {
	if check && !CheckDeductItems(actor, items) {
		return false
	}

	text := make([]string, len(items))
	var index int
	for _, item := range items {
		_, left := DeductItem(actor, item.Id, item.Count, false, "")
		text[index] = fmt.Sprintf("id(%d),deduct(%d),left(%d)", item.Id, item.Count, left)
		index++
	}

	log.Optf("DeductItem %s: actor(%d),%s", logText, actor.ActorId, strings.Join(text, ";"))

	return true
}

func DeductItems2(actor *t.Actor, items map[int]int, check bool, logText string) bool {
	if check && !CheckDeductItems2(actor, items) {
		return false
	}

	text := make([]string, len(items))
	var index int
	for id, count := range items {
		_, left := DeductItem(actor, id, count, false, "")
		text[index] = fmt.Sprintf("id(%d),deduct(%d),left(%d)", id, count, left)
		index++
	}

	log.Optf("DeductItem %s: actor(%d),%s", logText, actor.ActorId, strings.Join(text, ";"))

	return true
}

func onSendItemBagInfo(actor *t.Actor) {
	itemBag := GetItemsBag(actor)
	writer := pack.AllocPack(proto.Bag, proto.BagSItemInit, int16(len(itemBag)))
	for t, items := range itemBag {
		pack.Write(writer, int16(t), int16(len(items)))
		for id, count := range items {
			pack.Write(writer, id, count)
		}
	}
	actor.ReplyWriter(writer)

	currencyBag := GetCurrencyBag(actor)
	writer = pack.AllocPack(proto.Bag, proto.BagSCurrencyInit, int16(len(currencyBag)))
	for id, count := range currencyBag {
		pack.Write(writer, id, count)
	}
	actor.ReplyWriter(writer)
}
