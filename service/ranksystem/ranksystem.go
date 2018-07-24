package ranksystem

import (
	"bytes"

	"github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"
	c "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/data"
	"github.com/sencydai/gameworld/dispatch"
	t "github.com/sencydai/gameworld/typedefine"
)

func init() {
	dispatch.RegActorMsgHandle(proto.Rank, proto.RankCRankData, onGetRankData)
}

func onGetRankData(actor *t.Actor, reader *bytes.Reader) {
	var rankName string
	pack.Read(reader, &rankName)

	rankData := t.GetRank(rankName)
	if rankData == nil {
		return
	}

	writer := pack.AllocPack(proto.Rank, proto.RankSRankData, rankName, int16(len(rankData.RankList)))

	switch rankName {
	case c.RankLevel:
		fallthrough
	case c.RankPower:
		for _, item := range rankData.RankList {
			player := data.GetActor(item.Id)
			pack.Write(writer,
				float64(item.Id),
				player.ActorName,
				float64(item.Point),
				player.GetLordHead(),
				player.GetLordFrame(),
				player.Level,
				0,
				player.Power,
				"",
			)
		}
	}

	actor.ReplyWriter(writer)
}
