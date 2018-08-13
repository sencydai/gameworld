package lordequip

import (
	"bytes"

	c "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/dispatch"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/service/bag"
	t "github.com/sencydai/gameworld/typedefine"
)

func init() {
	service.RegActorLogin(onActorLogin)

	dispatch.RegActorMsgHandle(proto.Lord, proto.LordCEquipStreng, onStrengEquip)
}

func onSendChangeEquip(actor *t.Actor, pos, id int) {
	writer := pack.AllocPack(proto.Lord, proto.LordSChangeEquip, pos, id)
	actor.ReplyWriter(writer)
}

func onActorLogin(actor *t.Actor, offSec int) {
	equipData := actor.GetLordEquipData()
	writer := pack.AllocPack(
		proto.Lord,
		proto.LordSEquipInit,
		equipData.Pos,
		int16(len(equipData.Equips)),
	)
	for i := 1; i <= c.LEPMax; i++ {
		equip := equipData.Equips[i]
		pack.Write(writer, equip.Stage, equip.Level)
	}
	actor.ReplyWriter(writer)
}

//装备强化
func onStrengEquip(actor *t.Actor, reader *bytes.Reader) {
	equipData := actor.GetLordEquipData()
	posData := equipData.Equips[equipData.Pos]
	stageConfs := g.GLordEquipStrengConfig[posData.Stage]
	curConf := stageConfs[posData.Level]
	var ok bool
	levelConf, ok := stageConfs[posData.Level+1]
	//满级
	if !ok {
		//满阶
		if stageConfs, ok = g.GLordEquipStrengConfig[posData.Stage+1]; !ok {
			return
		}
		levelConf = stageConfs[0]
	}

	//物品数量不足
	if !bag.DeductItems(actor, curConf.Cost, true, "lordEquipStreng") {
		return
	}

	posData.Stage, posData.Level = levelConf.Stage, levelConf.Level
	if equipData.Pos == c.LEPMax {
		equipData.Pos = 1
	} else {
		equipData.Pos++
	}

	writer := pack.AllocPack(proto.Lord,
		proto.LordSStrengEquip,
		equipData.Pos,
		posData.Stage,
		posData.Level,
	)
	actor.ReplyWriter(writer)
}
