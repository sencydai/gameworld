package lordequip

import (
	"bytes"
	"fmt"
	"time"

	"github.com/sencydai/gamecommon/pack"
	proto "github.com/sencydai/gamecommon/protocol"
	"github.com/sencydai/gameworld/base"
	c "github.com/sencydai/gameworld/constdefine"
	"github.com/sencydai/gameworld/dispatch"
	g "github.com/sencydai/gameworld/gconfig"
	"github.com/sencydai/gameworld/service"
	"github.com/sencydai/gameworld/service/bag"
	"github.com/sencydai/gameworld/timer"
	t "github.com/sencydai/gameworld/typedefine"

	"github.com/sencydai/gameworld/log"
)

func init() {
	service.RegActorLogin(onActorLogin)
	service.RegActorBeforeLogin(onActorBeforeLogin)
	bag.RegObtainLordEquip(onObtainEquip)

	dispatch.RegActorMsgHandle(proto.Lord, proto.LordCEquipChange, onChangeEquip)
	dispatch.RegActorMsgHandle(proto.Lord, proto.LordCEquipStreng, onStrengEquip)
}

func onObtainEquip(actor *t.Actor, id, count int) {
	conf, ok := g.GLordEquipConfig[id]
	if !ok {
		log.Errorf("not find lord equip %d", id)
		return
	}

	equipData := actor.GetLordEquipData()
	posData := equipData.Equips[conf.Pos]

	//限时
	if conf.LimitTime != 0 {
		name := fmt.Sprintf("lordEquip_%d", id)
		if _, ok := posData.Unlock[id]; !ok {
			posData.Unlock[id] = int(time.Now().Unix())
		} else {
			timer.StopTimer(actor, name)
		}
		posData.Unlock[id] += conf.LimitTime * count * 3600 * 24
		timer.After(actor, name, base.IntSecond(posData.Unlock[id]), equipTimeout, conf.Pos, id)
	} else {
		posData.Unlock[id] = -1
	}

	//装备解锁
	writer := pack.AllocPack(proto.Lord, proto.LordSEquipUnlock, conf.Pos, id, posData.Unlock[id])
	actor.ReplyWriter(writer)

	//如果没有身着装备，默认穿上
	if posData.Id == -1 {
		posData.Id = id
		onSendChangeEquip(actor, conf.Pos, id)
	}
}

//装备过期
func equipTimeout(actor *t.Actor, pos, id int) {
	equipData := actor.GetLordEquipData()
	posData := equipData.Equips[pos]
	delete(posData.Unlock, id)

	writer := pack.AllocPack(proto.Lord, proto.LordSEquipTimeout, pos, id)
	actor.ReplyWriter(writer)

	if posData.Id == id {
		//找优先级最高的穿上
		min, selId := -1, -1
		for equipId := range posData.Unlock {
			conf := g.GLordEquipConfig[equipId]
			if conf.Rarity > min {
				min = conf.Rarity
				selId = equipId
			}
		}
		posData.Id = selId

		onSendChangeEquip(actor, pos, selId)
	}
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
	for pos, data := range equipData.Equips {
		pack.Write(writer, pos, data.Stage, data.Level, data.Id, int16(len(data.Unlock)))
		for id, timeout := range data.Unlock {
			pack.Write(writer, id, timeout)
		}
	}
	actor.ReplyWriter(writer)
}

func onActorBeforeLogin(actor *t.Actor, offSec int) {
	equipData := actor.GetLordEquipData()
	now := int(time.Now().Unix())
	for pos, equip := range equipData.Equips {
		for id, timeout := range equip.Unlock {
			if timeout > 0 {
				if timeout <= now {
					equipTimeout(actor, pos, id)
				} else {
					timer.After(actor, fmt.Sprintf("lordEquip_%d", id), base.IntSecond(now-timeout), equipTimeout, pos, id)
				}
			}
		}
	}
}

//更换装备
func onChangeEquip(actor *t.Actor, reader *bytes.Reader) {
	var id int
	pack.Read(reader, &id)
	conf, ok := g.GLordEquipConfig[id]
	if !ok {
		return
	}

	equipData := actor.GetLordEquipData()
	posData := equipData.Equips[conf.Pos]
	if posData.Id == id {
		return
	}

	if _, ok := posData.Unlock[id]; !ok {
		return
	}
	posData.Id = id
	onSendChangeEquip(actor, conf.Pos, id)
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
