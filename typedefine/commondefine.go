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

type MonsterLord struct {
	Id       int
	Template int
	Level    int
}

type MonsterHero struct {
	Pos      int
	Id       int
	Template int
	Level    int
}

type TimeWeek struct {
	Week  interface{} //map[int]int | map[int]map[int]int
	SHour int
	SMin  int
	SSec  int
	EHour int
	EMin  int
	ESec  int
}

type TimeOpenServer struct {
	SDay  int
	EDay  int
	SHour int
	SMin  int
	SSec  int
	EHour int
	EMin  int
	ESec  int
}

type TimeFixed struct {
	SYear  int
	SMonth int
	SDay   int
	EYear  int
	EMonth int
	EDay   int
	SHour  int
	SMin   int
	SSec   int
	EHour  int
	EMin   int
	ESec   int
}
