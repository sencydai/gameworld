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
