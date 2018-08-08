package typedefine

type ActorExData struct {
	Pf     string //平台
	NewDay int64  //上次newday时间

	ChangeName int                 //是否改过名
	Exp        int                 //经验
	MainFB     int                 //主线副本
	ClientData map[int]string      //前端数据记录
	Buildings  map[int]int         //建筑
	WorldBoss  *ActorWorldBossData //世界boss
}

type ActorWorldBossData struct {
	Count   int           //挑战次数
	Recover int64         //上次恢复时间
	CD      map[int]int64 //挑战cd id:time
}

func (actor *Actor) GetMainFuben() int {
	exData := actor.GetExData()
	return exData.MainFB
}

func (actor *Actor) GetBuildings() map[int]int {
	exData := actor.GetExData()
	if exData.Buildings == nil {
		exData.Buildings = make(map[int]int)
	}
	return exData.Buildings
}

func (actor *Actor) GetBuildingLevel(id int) int {
	buildings := actor.GetBuildings()
	return buildings[id]
}
