package typedefine

type Award struct {
	Type  int
	Id    int
	Count int
}

type AwardItem struct {
	Id    int
	Count int
}

type AwardRandom struct {
	Type  int
	Id    int
	Count int
	Rand  int
}

type Attr struct {
	Type  int
	Value int
}
