package typedefine

type LordBaseConfig struct {
	RenameCost int
}

type LordConfig struct {
	Camp   int
	Sex    int
	Head   int
	Frame  int
	Chat   int
	Job    int
	Heros  map[int]int
	Model  int
	Awards map[int]Award
}

type LordLevelConfig struct {
	Level   int
	NeedExp int
	Attrs   map[int]Attr
}

type LordEquipConfig struct {
	Id        int
	Type      int
	Name      string
	Pos       int
	Rarity    int
	LimitTime int
	Attrs     map[int]Attr
}

type LordEquipStrengConfig struct {
	Stage int
	Level int
	Cost  map[int]AwardItem
	Attrs map[int]Attr
}

type ItemConfig struct {
	Id        int
	Name      string
	Type      int
	SubType   int
	Rarity    int
	Worth     int
	OpenKey   int
	AwardType int
	Award     map[int]AwardRandom
	ComTarget *struct {
		Type int
		Id   int
	}
	ComCount int
}

type ItemGroupConfig struct {
	Id         int
	Index      int
	RandomType int
	Ratio      int
	Award      Award
}

type VirtualCurrencyConfig struct {
	VituralId int
	RealId    int
	AttrType  int
}
