package typedefine

import (
	"time"

	"github.com/sencydai/gameworld/rank"
)

const (
	SYSTEM_ACTOR_INDEX      = 1
	SYSTEM_RANK_INDEX       = 2
	SYSTEM_GUILD_INDEX      = 3
	SYSTEM_COMMON_INDEX     = 4
	SYSTEM_OPENSERVER_INDEX = 5
)

type SystemData struct {
	Actors     map[int64]*SystemStaticActorData
	Rank       map[string]*rank.RankData
	Guild      *SystemStaticGuildData
	Data       *SystemStaticCommonData
	OpenServer time.Time

	DynamicData *SystemDynamicData
}

var (
	sysData = &SystemData{
		Actors:      make(map[int64]*SystemStaticActorData),
		Rank:        make(map[string]*rank.RankData),
		Guild:       &SystemStaticGuildData{},
		Data:        &SystemStaticCommonData{},
		OpenServer:  time.Now(),
		DynamicData: &SystemDynamicData{},
	}
)

type SystemStaticActorData struct {
}

type SystemStaticGuildData struct {
}

type SystemStaticCommonData struct {
	NewDay   int64
	Worlboss map[int]*SystemWorldBossStaticData //世界boss id
}

type SystemWorldBossStaticData struct {
	Id int //boss id

	Refresh int64 //下次刷新时间

	Killer   int64           //杀死的玩家
	Lucky    int64           //幸运玩家
	Fighting map[int64]int64 //正在挑战的玩家 id : fight guid

	RawData *FightMonsterRawData
}

type SystemDynamicData struct {
}

func GetSysData() *SystemData {
	return sysData
}

func GetSysActorData(actorId int64) *SystemStaticActorData {
	data, ok := sysData.Actors[actorId]
	if !ok {
		data = &SystemStaticActorData{}
		sysData.Actors[actorId] = data
	}
	return data
}

func GetRank(name string) *rank.RankData {
	return sysData.Rank[name]
}

func DeleteRank(name string) {
	delete(sysData.Rank, name)
}

func NewRank(name string, maxRankCount int) *rank.RankData {
	rankData, ok := sysData.Rank[name]
	if ok {
		if rankData.MaxCount != maxRankCount {
			rankData.SetMaxRankCount(maxRankCount)
		}
		return rankData
	}

	sysData.Rank[name] = rank.NewRank(maxRankCount)
	return sysData.Rank[name]
}

func GetSysGuildData() *SystemStaticGuildData {
	return sysData.Guild
}

func GetSysCommonData() *SystemStaticCommonData {
	return sysData.Data
}

func GetSysOpenServerTime() time.Time {
	return sysData.OpenServer
}

func GetSysDynamicData() *SystemDynamicData {
	return sysData.DynamicData
}
