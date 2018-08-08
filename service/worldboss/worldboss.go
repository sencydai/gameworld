package worldboss

import (
	"bytes"
	"fmt"
	"math"
	"time"

	"github.com/sencydai/gameworld/base"
	c "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/data"
	"github.com/sencydai/gameworld/dispatch"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/log"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
	"github.com/sencydai/gameworld/rank"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/service/bag"
	"github.com/sencydai/gameworld/service/fight"
	"github.com/sencydai/gameworld/timer"
	t "github.com/sencydai/gameworld/typedefine"
)

const (
	worldbossRankName = "rank_worldboss_%d"
	timerReset        = "resetworldboss_%d"
	maxRankCount      = 3
)

var (
	recoverSec int
)

func init() {
	service.RegConfigLoadFinish(onConfigLoadFinish)
	service.RegGameStart(onGameStart)
	service.RegSystemTimeChange(onSystemTimeChange)
	service.RegActorLogin(onActorLogin)
	dispatch.RegActorMsgHandle(proto.Fuben, proto.FubenCUpdateWorldBossInfo, onUpdateBossInfo)
	dispatch.RegActorMsgHandle(proto.Fuben, proto.FubenCWorldBossChallenge, onChallenge)
	fight.RegFightHpSync(fight.WorldBoss, onBossHpSync)
	fight.RegFightAward(fight.WorldBoss, onFightAwards)
	fight.RegFightGiveup(fight.WorldBoss, onGiveupFight)
}

func getActorBossData(actor *t.Actor) *t.ActorWorldBossData {
	exData := actor.GetExData()
	if exData.WorldBoss == nil {
		exData.WorldBoss = &t.ActorWorldBossData{
			Recover: time.Now().Unix(),
			CD:      make(map[int]int64),
		}
	}

	return exData.WorldBoss
}

func getBosses() map[int]*t.SystemWorldBossStaticData {
	commonData := t.GetSysCommonData()
	if commonData.Worlboss == nil {
		commonData.Worlboss = make(map[int]*t.SystemWorldBossStaticData)
	}

	return commonData.Worlboss
}

func getBoss(id int) *t.SystemWorldBossStaticData {
	bosses := getBosses()
	return bosses[id]
}

func getBossHp(boss *t.SystemWorldBossStaticData) (total, cur int) {
	for _, hp := range boss.RawData.Heros {
		cur += hp
	}

	for _, hp := range boss.RawData.RawHeros {
		total += hp
	}

	return
}

func onConfigLoadFinish(isGameStart bool) {
	recoverSec = g.GWorldBossBaseConfig.Recover * 3600

	if isGameStart {
		return
	}
	bosses := getBosses()
	//新boss
	for _, conf := range g.GWorldBossConfig {
		if _, ok := bosses[conf.Id]; ok {
			continue
		}

		boss := &t.SystemWorldBossStaticData{Id: conf.Id}
		bosses[conf.Id] = boss
		resetBoss(boss, true)
	}
}

func resetBoss(boss *t.SystemWorldBossStaticData, broadcast bool) {
	boss.Refresh = 0
	boss.Killer = 0
	boss.Lucky = 0
	boss.Fighting = make(map[int64]int64)

	conf := g.GWorldBossConfig[boss.Id]
	boss.RawData = fight.CreatFightMonsterRawData(conf.Monster)

	rank := t.NewRank(fmt.Sprintf(worldbossRankName, conf.Id), -1)
	rank.Reset()

	if broadcast {
		broadcastBossInfo(boss)
	}
}

func broadcastBossInfo(boss *t.SystemWorldBossStaticData) {
	data.LoopActors(func(actor *t.Actor) bool {
		actorData := getActorBossData(actor)
		writer := pack.AllocPack(proto.Fuben, proto.FubenSUpdateWorldBossInfo)
		packBossData(writer, actor, boss, actorData)
		actor.ReplyWriter(writer)
		return true
	})
}

func onGameStart() {
	bosses := getBosses()

	//删除不存在的boss
	for id, boss := range bosses {
		if _, ok := g.GWorldBossConfig[id]; !ok {
			delete(bosses, id)
			t.DeleteRank(fmt.Sprintf(worldbossRankName, id))
		} else {
			for aid := range boss.Fighting {
				delete(boss.Fighting, aid)
			}
		}
	}

	//新boss
	for _, conf := range g.GWorldBossConfig {
		if _, ok := bosses[conf.Id]; ok {
			continue
		}

		boss := &t.SystemWorldBossStaticData{Id: conf.Id}
		bosses[conf.Id] = boss
		resetBoss(boss, false)
	}

	for _, boss := range bosses {
		if boss.Refresh == 0 {
			continue
		}
		now := time.Now().Unix()
		if boss.Refresh <= now {
			resetBoss(boss, false)
		} else {
			timer.After(nil, fmt.Sprintf(timerReset, boss.Id), time.Second*time.Duration(boss.Refresh-now), resetBoss, boss, true)
		}
	}
}

func onSystemTimeChange() {
	bosses := getBosses()
	for _, boss := range bosses {
		if boss.Refresh == 0 {
			continue
		}
		name := fmt.Sprintf(timerReset, boss.Id)
		timer.StopTimer(nil, name)

		now := time.Now().Unix()
		if boss.Refresh <= now {
			resetBoss(boss, true)
		} else {
			timer.After(nil, fmt.Sprintf(timerReset, boss.Id), time.Second*time.Duration(boss.Refresh-now), resetBoss, boss, true)
		}
	}
}

func recoverChallengeCount(actor *t.Actor) {
	data := getActorBossData(actor)
	data.Recover = time.Now().Unix()
	if data.Count > 0 {
		data.Count--
	}

	onSyncChallengeCount(actor)
}

func onSyncChallengeCount(actor *t.Actor) {
	data := getActorBossData(actor)
	actor.Reply(proto.Fuben, proto.FubenSWorldBossSyncChallengeCount,
		g.GWorldBossBaseConfig.MaxCount-data.Count,
		recoverSec-int(time.Now().Unix()-data.Recover),
	)
}

func onActorLogin(actor *t.Actor, offSec int) {
	data := getActorBossData(actor)

	sec := int(math.Max(0, float64(time.Now().Unix()-data.Recover)))
	count := sec / recoverSec
	data.Count = int(math.Max(0, float64(data.Count-count)))
	data.Recover += int64(count * recoverSec)

	timer.Loop(actor, "recoverWorldBoss", time.Second*time.Duration(sec%recoverSec), time.Second*time.Duration(recoverSec), -1, recoverChallengeCount)

	bosses := getBosses()
	writer := pack.AllocPack(proto.Fuben, proto.FubenSWorldBossInit,
		g.GWorldBossBaseConfig.MaxCount-data.Count,
		recoverSec-int(time.Now().Unix()-data.Recover),
		int16(len(bosses)),
	)

	for _, boss := range bosses {
		packBossData(writer, actor, boss, data)
	}

	actor.ReplyWriter(writer)
}

func packBossData(writer *bytes.Buffer, actor *t.Actor, boss *t.SystemWorldBossStaticData, actorData *t.ActorWorldBossData) {
	pack.Write(writer, boss.Id)

	rankData := t.GetRank(fmt.Sprintf(worldbossRankName, boss.Id))

	//等待重生中
	if boss.Refresh > 0 {
		delay := int(math.Max(1, float64(boss.Refresh-time.Now().Unix())))
		item := rankData.RankList[0]
		player := data.GetActor(item.Id)
		pack.Write(writer, delay, player.ActorName)
		return
	}

	//已重生
	total, cur := getBossHp(boss)
	var (
		challenge int
		cd        int
	)
	if _, ok := rankData.GetIdPoint(actor.ActorId); ok {
		challenge = 1
	}
	if join, ok := actorData.CD[boss.Id]; ok {
		cd = int(join - time.Now().Unix())
		if cd <= 0 {
			cd = 0
			delete(actorData.CD, boss.Id)
		}
	}

	pack.Write(writer, 0, total, cur, challenge, cd)
}

func onUpdateBossInfo(actor *t.Actor, reader *bytes.Reader) {
	var id int
	pack.Read(reader, &id)
	boss := getBoss(id)
	if boss == nil || boss.Refresh > 0 {
		return
	}
	actorData := getActorBossData(actor)
	writer := pack.AllocPack(proto.Fuben, proto.FubenSUpdateWorldBossInfo)
	packBossData(writer, actor, boss, actorData)
	actor.ReplyWriter(writer)
}

func onChallenge(actor *t.Actor, reader *bytes.Reader) {
	var id int
	pack.Read(reader, &id)
	boss := getBoss(id)
	if boss == nil || boss.Refresh > 0 {
		return
	}
	actorData := getActorBossData(actor)

	bossConf := g.GWorldBossConfig[id]
	if bossConf.Level > actor.Level {
		return
	}

	rankData := t.GetRank(fmt.Sprintf(worldbossRankName, id))
	damage, damageOK := rankData.GetIdPoint(actor.ActorId)

	//挑战次数不足
	if actorData.Count > g.GWorldBossBaseConfig.MaxCount {
		return
	} else if actorData.Count == g.GWorldBossBaseConfig.MaxCount && !damageOK {
		return
	}
	join, ok := actorData.CD[id]
	if ok {
		if join > time.Now().Unix() {
			return
		}
		delete(actorData.CD, id)
	}

	total, _ := getBossHp(boss)
	fightData := fight.NewPvE(actor, fight.WorldBoss, bossConf.Lord, bossConf.Monster, "", 0, boss.RawData, []interface{}{boss.Id, int(damage), float64(total)}, boss)
	if fightData == nil {
		return
	}
	boss.Fighting[actor.ActorId] = fightData.Guid
	actorData.CD[id] = time.Now().Unix() + int64(g.GWorldBossBaseConfig.Cd)
	if !damageOK {
		actorData.Count++
	}
	onSyncChallengeCount(actor)
}

func onBossHpSync(actor *t.Actor, fightData *t.FightData) {
	boss := fightData.CbArgs[0].(*t.SystemWorldBossStaticData)
	if boss.Killer != 0 {
		return
	}

	rankData := t.GetRank(fmt.Sprintf(worldbossRankName, boss.Id))

	//更新伤害榜
	count := int(math.Min(float64(maxRankCount), float64(len(rankData.RankList))))
	total, cur := getBossHp(boss)
	writer := pack.AllocPack(proto.Fight, proto.FightSUpdateDamageRank, float64(fightData.Guid), total-cur, int16(count))

	for i := 0; i < count; i++ {
		rankItem := rankData.RankList[i]
		player := data.GetActor(rankItem.Id)
		pack.Write(writer, player.ActorName, int(rankItem.Point))
	}

	actor.ReplyWriter(writer)

	var damage int
	var totalHp int
	rawHps := boss.RawData.Heros
	for _, entity := range fightData.RawEntities {
		if entity.LordIndex != 1 {
			continue
		}
		hp := int(entity.Attrs[c.AttrHp])
		damage += rawHps[entity.HeroPos] - hp
		rawHps[entity.HeroPos] = hp
		totalHp += hp
	}

	isDeath := totalHp == 0

	if damage > 0 {
		if old, ok := rankData.GetIdPoint(actor.ActorId); ok {
			damage += int(old)
		}
		rankData.Insert(actor.ActorId, int64(damage))

		//其它玩家boss血量同步
		for aid, guid := range boss.Fighting {
			if aid == actor.ActorId {
				continue
			}
			player := data.GetOnlineActor(aid)
			if player == nil {
				delete(boss.Fighting, aid)
				continue
			}
			playerFight := player.GetFightData()
			if playerFight == nil || playerFight.Guid != guid {
				delete(boss.Fighting, aid)
				continue
			}

			hps := make(map[int]int)
			for _, entity := range playerFight.Entities {
				if entity.LordIndex != 1 {
					continue
				}
				entity.Attrs[c.AttrHp] = float64(rawHps[entity.HeroPos])
				if rawHps[entity.HeroPos] == 0 {
					entity.IsDead = true
					delete(playerFight.Entities, entity.Pos)
				}
				hps[entity.Pos] = rawHps[entity.HeroPos]
			}
			if isDeath && playerFight.FightResult == 0 {
				playerFight.FightResult = fight.Win
			}

			writer := pack.AllocPack(proto.Fight, proto.FightSHpSync, float64(guid), actor.ActorName, int16(len(hps)))
			for pos, hp := range hps {
				pack.Write(writer, pos, hp)
			}
			player.ReplyWriter(writer)
		}
	}

	if isDeath {
		onBossDeath(actor, boss)
	}
}

func onBossDeath(actor *t.Actor, boss *t.SystemWorldBossStaticData) {
	log.Infof("onBossDeath: %d killer %d %s", boss.Id, actor.ActorId, actor.ActorName)
	boss.Killer = actor.ActorId

	//更新伤害榜
	rankData := t.GetRank(fmt.Sprintf(worldbossRankName, boss.Id))
	total, cur := getBossHp(boss)
	writer := pack.NewWriter(total-cur, int16(len(rankData.RankList)))
	for _, rankItem := range rankData.RankList {
		player := data.GetActor(rankItem.Id)
		pack.Write(writer, player.ActorName, int(rankItem.Point))
	}
	packData := writer.Bytes()
	for aid, guid := range boss.Fighting {
		player := data.GetOnlineActor(aid)
		if player == nil {
			continue
		}
		writer := pack.AllocPack(proto.Fight, proto.FightSUpdateDamageRank, float64(guid))
		pack.Write(writer, packData)
		player.ReplyWriter(writer)
	}

	bossConf := g.GWorldBossConfig[boss.Id]
	boss.Refresh = time.Now().Unix() + int64(bossConf.Cd*60)
	timer.After(nil, fmt.Sprintf(timerReset, boss.Id), time.Second*time.Duration(bossConf.Cd*60), resetBoss, boss, true)

	calcLuckActor(boss, rankData, total)
	actorsRankAwards := calcRankAwards(boss, rankData)

	clearWriter := pack.NewWriter()
	//前三名
	count := int(math.Min(float64(maxRankCount), float64(len(rankData.RankList))))
	pack.Write(clearWriter, int16(count))
	for i := 0; i < count; i++ {
		item := rankData.RankList[i]
		player := data.GetActor(item.Id)
		pack.Write(clearWriter, player.ActorName, int(item.Point))
	}

	//幸运玩家
	lucky := data.GetActor(boss.Lucky)
	pack.Write(clearWriter,
		lucky.ActorName,
		lucky.GetLordHead(),
		lucky.GetLordFrame(),
		lucky.Level,
		0,
	)
	pack.Write(clearWriter, float64(boss.Killer))
	clearData := clearWriter.Bytes()

	for index, rankItem := range rankData.RankList {
		player := data.GetActor(rankItem.Id)

		//击杀奖励
		var killAwards map[int]t.Award
		if player.ActorId == boss.Killer {
			killAwards = bag.GetRealAwards(player, bossConf.Kill)
		}
		//幸运奖励
		var luckAwards map[int]t.Award
		if player.ActorId == boss.Lucky {
			luckAwards = bossConf.Drops
		}
		//排名奖励
		rankAwards := actorsRankAwards[player.ActorId]
		playerFight := player.GetFightData()
		if playerFight != nil && playerFight.Guid == boss.Fighting[player.ActorId] {
			awards := make(map[int]t.Award)
			for _, award := range killAwards {
				awards[len(awards)] = award
			}
			for _, award := range luckAwards {
				awards[len(awards)] = award
			}
			for _, award := range rankAwards {
				awards[len(awards)] = award
			}
			playerFight.Awards = bag.FlushAwards(player, awards)
			writer := pack.NewWriter(index+1, int(rankItem.Point), clearData)
			playerFight.ClearArgs = []interface{}{writer.Bytes()}
			if player.ActorId != boss.Killer {
				timer.StopTimer(player, fmt.Sprintf("triggerFighting_%d", playerFight.Type))
				timer.StopTimer(player, fmt.Sprintf("OnFightClear_%d", playerFight.Type))
				fight.OnFightClear(player, playerFight)
			}
		} else {
			//发送奖励邮件
		}
	}

	broadcastBossInfo(boss)
}

func calcLuckActor(boss *t.SystemWorldBossStaticData, rankData *rank.RankData, totalDamage int) {
	rand := int64(base.Rand(0, totalDamage))
	for _, item := range rankData.RankList {
		if rand <= item.Point {
			boss.Lucky = item.Id
			return
		}
		rand -= item.Point
	}
}

func calcRankAwards(boss *t.SystemWorldBossStaticData, rankData *rank.RankData) map[int64]map[int]t.Award {
	maxRank := len(rankData.RankList) - 1
	rankConfs := g.GWorldBossRankConfig[boss.Id]
	var index int
	rankAwards := make(map[int64]map[int]t.Award)
	for i := 1; i <= len(rankConfs); i++ {
		conf := rankConfs[i]
		awards := bag.GetRealAwards(nil, conf.Awards)
		for j := index; i < conf.Upper-1; j++ {
			if j > maxRank {
				return rankAwards
			}
			item := rankData.RankList[j]
			rankAwards[item.Id] = awards
		}
		index = conf.Upper
	}

	return rankAwards
}

func onFightAwards(actor *t.Actor, fightData *t.FightData) {
	boss := fightData.CbArgs[0].(*t.SystemWorldBossStaticData)
	delete(boss.Fighting, actor.ActorId)
}

func onGiveupFight(actor *t.Actor, fightData *t.FightData) {
	boss := fightData.CbArgs[0].(*t.SystemWorldBossStaticData)
	delete(boss.Fighting, actor.ActorId)
}
