package typedefine

import (
	"time"
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
	Rank       *SystemStaticRankData
	Guild      *SystemStaticGuildData
	Data       *SystemStaticCommonData
	OpenServer time.Time

	DynamicData *SystemDynamicData
}

type SystemStaticActorData struct {
}

type SystemStaticRankData struct {
}

type SystemStaticGuildData struct {
}

type SystemStaticCommonData struct {
	NewDay int64
}

type SystemDynamicData struct {
}
