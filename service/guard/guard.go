package guard

import (
	g "github.com/sencydai/gameworld/gconfig"

	"github.com/sencydai/gameworld/service"
	t "github.com/sencydai/gameworld/typedefine"
)

func init() {
	service.RegActorLogin(onActorLogin)
	service.RegActorUpgrade(onActorUpgrade)
}

func onActorLogin(actor *t.Actor, offSec int) {

}

//玩家升级
func onActorUpgrade(actor *t.Actor, oldLevel int) {
	guard := actor.GetGuardData()
	if guard.Level != 0 {
		return
	}

	guard.Level = 1
	conf := g.GGuardLevelConfig[1]
	if conf.ModelId != 0 {
		guard.Model = conf.ModelId
	}
}
