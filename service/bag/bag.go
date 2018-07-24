package bag

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sencydai/gameworld/data"

	"github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"
	"github.com/sencydai/gameworld/base"
	c "github.com/sencydai/gameworld/constdefine"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/log"
	t "github.com/sencydai/gameworld/typedefine"

	"github.com/sencydai/gameworld/service"
)

var (
	onObtainLordExp      func(actor *t.Actor, count int)
	onObtainAreaPrestige func(actor *t.Actor, count int)
	onObtainGuildContri  func(actor *t.Actor, count int)
	onObtainGuildFund    func(actor *t.Actor, count int)
	onObtainLordEquip    func(actor *t.Actor, id, count int)
	onObtainLordDecor    func(actor *t.Actor, ldt c.LordDecorType, id int)
)

const (
	maxArtifactCount = 500
)

func init() {
	service.RegActorCreate(onActorCreate)
	service.RegActorLogin(onActorLogin)
	service.RegGm("award", onGmAwards)
}

func onGmAwards(values map[string]string) (int, string) {
	aid, _ := strconv.ParseFloat(values["actor"], 64)
	actor := data.GetOnlineActor(int64(aid))
	if actor == nil {
		return -1, fmt.Sprintf("not find online actor: %d", int64(aid))
	}
	t, _ := strconv.Atoi(values["type"])
	id, _ := strconv.Atoi(values["id"])
	count, _ := strconv.Atoi(values["count"])
	if count < 0 {
		count = 0
	}

	if !PutAward2Bag(actor, t, id, count, true, c.ASCommon, "gm") {
		return -2, "put award failed"
	}

	return 0, "success"
}

func onActorLogin(actor *t.Actor, offSec int) {
	onSendItemBagInfo(actor)
	onSendHeroBagInfo(actor)
	onSendEquipBagInfo(actor)
	onSendArtifactBagInfo(actor)
}

func onSendHeroBagInfo(actor *t.Actor) {
	heros := actor.GetBagHeroData()
	writer := pack.AllocPack(proto.Bag, proto.BagSHeroInit, int16(len(heros.Heros)))
	for guid, hero := range heros.Heros {
		pack.Write(writer,
			guid,
			int8(hero.PosType),
			int16(hero.Pos),
			hero.Map,
			hero.Id,
			int16(hero.Level),
			hero.Exp,
			int16(hero.Stage),
		)
	}
	actor.ReplyWriter(writer)
}

func onSendEquipBagInfo(actor *t.Actor) {
	equips := actor.GetBagEquipData()
	writer := pack.AllocPack(proto.Bag, proto.BagSEquipInit, int16(len(equips.Equips)))
	for guid, equip := range equips.Equips {
		pack.Write(writer, guid, equip.Pos, equip.Id, equip.Level)
	}
	actor.ReplyWriter(writer)
}

func onSendArtifactBagInfo(actor *t.Actor) {
	artiData := actor.GetBagArtifactData()
	writer := pack.AllocPack(proto.Bag, proto.BagSArtiInit, int16(len(artiData.Artis)))
	for guid, arti := range artiData.Artis {
		pack.Write(writer, guid, arti.Pos, arti.Id, arti.Stage, arti.Level)
	}
	actor.ReplyWriter(writer)
}

func RegObtainLordExp(handle func(actor *t.Actor, count int)) {
	onObtainLordExp = handle
}

func RegObtainAreaPrestige(handle func(actor *t.Actor, count int)) {
	onObtainAreaPrestige = handle
}

func RegObtainGuildContri(handle func(actor *t.Actor, count int)) {
	onObtainGuildContri = handle
}

func RegObtainGuildFund(handle func(actor *t.Actor, count int)) {
	onObtainGuildFund = handle
}

func RegObtainLordDecor(handle func(actor *t.Actor, ldt c.LordDecorType, id int)) {
	onObtainLordDecor = handle
}

func RegObtainLordEquip(handle func(actor *t.Actor, id, count int)) {
	onObtainLordEquip = handle
}

func onActorCreate(actor *t.Actor) {
	lordConf := g.GLordConfig[actor.Camp][actor.Sex]
	PutAwards2Bag(actor, lordConf.Awards, true, false, c.ASNoAction, "createActor")
}

func AwardsString(awards map[int]t.Award) string {
	text := make([]string, len(awards))
	index := 0
	for _, award := range awards {
		text[index] = fmt.Sprintf("%d,%d,%d", award.Type, award.Id, award.Count)
		index++
	}

	return strings.Join(text, ";")
}

func FlushAwards(actor *t.Actor, awards map[int]t.Award) map[int]t.Award {
	items := make(map[int]map[int]int)
	for _, award := range awards {
		types, ok := items[award.Type]
		if !ok {
			types = make(map[int]int)
			items[award.Type] = types
		}
		if award.Type == c.ATItem && award.Id <= 1000 {
			id, count := parseVirtualCurrency(actor, award.Id, award.Count)
			types[id] += count
		} else {
			types[award.Id] += award.Count
		}
	}

	values := make(map[int]t.Award)
	for ty, award := range items {
		for id, count := range award {
			values[len(values)+1] = t.Award{Type: ty, Id: id, Count: count}
		}
	}
	return values
}

func ParseItemGroup(actor *t.Actor, id, count int) map[int]t.Award {
	awards := make(map[int]t.Award)
	groups := g.GItemGroupConfig[id]
	rType := groups[1].RandomType
	for i := 0; i < count; i++ {
		rAwards := make(map[int]t.AwardRandom)
		for j, value := range groups {
			award := value.Award
			rAwards[j] = t.AwardRandom{Type: award.Type, Id: award.Id, Count: award.Id, Rand: value.Ratio}
		}
		for _, award := range GetRandomAwards(actor, rType, rAwards, false) {
			awards[len(awards)+1] = award
		}
	}

	return awards
}

func GetRandomAwards(actor *t.Actor, rType c.RandomType, rAwards map[int]t.AwardRandom, flush bool) map[int]t.Award {
	awards := make(map[int]t.Award)
	switch rType {
	//概率
	case c.RTProbability:
		for _, rAward := range rAwards {
			if base.Rand(1, 10000) <= rAward.Rand {
				//物品组
				if rAward.Type == c.ATItemGroup {
					for _, award := range ParseItemGroup(actor, rAward.Id, rAward.Count) {
						awards[len(awards)+1] = t.Award{Type: award.Type, Id: award.Id, Count: award.Count}
					}
				} else {
					awards[len(awards)+1] = t.Award{Type: rAward.Type, Id: rAward.Id, Count: rAward.Count}
				}
			}
		}
	//权值
	case c.RTWeight:
		total := 0
		for _, rAward := range rAwards {
			total += rAward.Rand
		}
		rand := base.Rand(1, total)
		for _, rAward := range rAwards {
			if rand <= rAward.Rand {
				//物品组
				if rAward.Type == c.ATItemGroup {
					for _, award := range ParseItemGroup(actor, rAward.Id, rAward.Count) {
						awards[len(awards)+1] = t.Award{Type: award.Type, Id: award.Id, Count: award.Count}
					}
				} else {
					awards[len(awards)+1] = t.Award{Type: rAward.Type, Id: rAward.Id, Count: rAward.Count}
				}
				break
			}
			rand -= rAward.Rand
		}
	}

	if flush {
		awards = FlushAwards(actor, awards)
	}

	return awards
}

func GetRealAwards(actor *t.Actor, awards map[int]t.Award) map[int]t.Award {
	values := make(map[int]t.Award)
	for _, award := range awards {
		if award.Type == c.ATItemGroup {
			for _, item := range ParseItemGroup(actor, award.Id, award.Count) {
				values[len(values)+1] = t.Award{Type: item.Type, Id: item.Id, Count: item.Count}
			}
		} else {
			values[len(values)+1] = award
		}
	}

	return FlushAwards(actor, values)
}

func maxOweHeroCount(actor *t.Actor) int {
	level := actor.GetBuildingLevel(c.BuildingHouse)
	if level == 0 {
		return 0
	}

	conf := g.GBuildingLevelConfig[c.BuildingHouse][level]
	return conf.MaxHeros
}

func CheckPutAwards(actor *t.Actor, awards map[int]t.Award, flush bool) bool {
	if flush {
		awards = GetRealAwards(actor, awards)
	}
	var items map[int]int
	for _, award := range awards {
		switch award.Type {
		case c.ATHero:
			fallthrough
		case c.ATEquip:
			fallthrough
		case c.ATArtifact:
			if items == nil {
				items = make(map[int]int)
			}
			items[award.Type] += award.Count
		}
	}

	for ty, total := range items {
		switch ty {
		case c.ATHero:
			bagHero := actor.GetBagHeroData()
			for _, hero := range bagHero.Heros {
				if hero.PosType == 0 || hero.PosType != c.HPTExpedition {
					total++
				}
			}
			if total > maxOweHeroCount(actor) {
				return false
			}
		case c.ATEquip:
			equipData := actor.GetBagEquipData()
			for _, equip := range equipData.Equips {
				if equip.Pos == 0 {
					total++
				}
			}
			if total > g.GEquipBaseConfig.MaxCount {
				return false
			}
		case c.ATArtifact:
			artiBag := actor.GetBagArtifactData()
			if total+len(artiBag.Artis) > maxArtifactCount {
				return false
			}
		}
	}

	return true
}

func NewHero(actor *t.Actor, id int) *t.HeroStaticData {
	bagHeros := actor.GetBagHeroData()
	bagHeros.MaxId++
	guid := bagHeros.MaxId
	hero := &t.HeroStaticData{
		Guid:  guid,
		Id:    id,
		Level: 1,
		Exp:   0,
		Stage: 0,
		Power: 1,
	}
	bagHeros.Heros[guid] = hero
	return hero
}

func PutAwards2Bag(actor *t.Actor, awards map[int]t.Award, flush bool, checkFull bool, source c.AwardSource, logText string) bool {
	if flush {
		awards = GetRealAwards(actor, awards)
	}
	if checkFull && !CheckPutAwards(actor, awards, false) {
		return false
	}

	writer := pack.AllocPack(proto.Bag, proto.BagSAddAwards, source, int16(len(awards)))
	for _, award := range awards {
		pack.Write(writer, award.Type, award.Id, award.Count)
		switch award.Type {
		//物品
		case c.ATItem:
			itemConf, ok := g.GItemConfig[award.Id]
			if !ok {
				log.Errorf("not find item conf %d", award.Id)
				continue
			}
			add := true
			//货币
			if itemConf.Type == c.ITCurrency {
				switch itemConf.SubType {
				//领主经验
				case c.CTLExp:
					add = false
					if onObtainLordExp != nil {
						onObtainLordExp(actor, award.Count)
					}
				//地区声望
				case c.CTAreaPrestige:
					add = false
					if onObtainAreaPrestige != nil {
						onObtainAreaPrestige(actor, award.Count)
					}
				//公会贡献(给个人)
				case c.CTGuildContri:
					if onObtainGuildContri != nil {
						onObtainGuildContri(actor, award.Count)
					}
				//公会资金(给公会)
				case c.CTGuildFund:
					add = false
					if onObtainGuildFund != nil {
						onObtainGuildFund(actor, award.Count)
					}
				}

				if add {
					accCurrency := GetAccCurrencyBag(actor)
					accCurrency[award.Id] += award.Count
					//todo: 更新任务
				}
			}

			if add {
				items := GetItemTypeBag(actor, itemConf.Type)
				items[award.Id] += award.Count
				pack.Write(writer, items[award.Id])
			} else {
				pack.Write(writer, 0)
			}

		//领主头像
		case c.ATLHead:
			if onObtainLordDecor != nil {
				onObtainLordDecor(actor, c.LDTHead, award.Id)
			}
		//领主头像边框
		case c.ATLFrame:
			if onObtainLordDecor != nil {
				onObtainLordDecor(actor, c.LDTFrame, award.Id)
			}
		//领主聊天框
		case c.ATLChat:
			if onObtainLordDecor != nil {
				onObtainLordDecor(actor, c.LDTChat, award.Id)
			}
		//领主装备
		case c.ATLEquip:
			if onObtainLordEquip != nil {
				onObtainLordEquip(actor, award.Id, award.Count)
			}
		//英雄
		case c.ATHero:
			if _, ok := g.GHeroConfig[award.Id]; !ok {
				log.Errorf("not find hero conf %d", award.Id)
				continue
			}
			for i := 0; i < award.Count; i++ {
				hero := NewHero(actor, award.Id)
				pack.Write(writer, hero.Guid)
			}
		//装备
		case c.ATEquip:
			if _, ok := g.GEquipConfig[award.Id]; !ok {
				log.Errorf("not find equip conf %d", award.Id)
				continue
			}
			equipData := actor.GetBagEquipData()
			for i := 0; i < award.Count; i++ {
				equipData.MaxId++
				equip := &t.EquipStaticData{
					Guid: equipData.MaxId,
					Id:   award.Id,
				}
				equipData.Equips[equip.Guid] = equip
				pack.Write(writer, equip.Guid)
			}
		//神器
		case c.ATArtifact:
			if _, ok := g.GArtifactConfig[award.Id]; !ok {
				log.Errorf("not find artifact conf %d", award.Id)
				continue
			}
			artiData := actor.GetBagArtifactData()
			for i := 0; i < award.Count; i++ {
				artiData.MaxId++
				arti := &t.ArtifactStaticData{
					Guid: artiData.MaxId,
					Id:   award.Id,
				}
				artiData.Artis[arti.Guid] = arti
				pack.Write(writer, arti.Guid)
			}
		}
	}

	actor.ReplyWriter(writer)

	log.Infof("PutAwards2Bag %s: actor(%d),awards(%s)", logText, actor.ActorId, AwardsString(awards))

	return true
}

func PutAwards2BagRatio(actor *t.Actor, awards map[int]t.Award, ratio float64, flush bool, checkFull bool, source c.AwardSource, logText string) bool {
	values := make(map[int]t.Award)
	for index, award := range awards {
		values[index] = t.Award{
			Type:  award.Type,
			Id:    award.Id,
			Count: int(float64(award.Count) * ratio),
		}
	}

	return PutAwards2Bag(actor, values, flush, checkFull, source, logText)
}

func PutAward2Bag(actor *t.Actor, at c.AwardType, id, count int, checkFull bool, source c.AwardSource, logText string) bool {
	awards := make(map[int]t.Award)
	awards[1] = t.Award{Type: at, Id: id, Count: count}
	return PutAwards2Bag(actor, awards, at == c.ATItemGroup, checkFull, source, logText)
}

func onSendHeroStaticData(actor *t.Actor, hero *t.HeroStaticData) {
	actor.Reply(proto.Bag, proto.BagSHeroUpdate,
		hero.Guid,
		hero.PosType,
		hero.Pos,
		hero.Map,
		hero.Id,
		hero.Level,
		hero.Exp,
		hero.Stage,
	)
}
