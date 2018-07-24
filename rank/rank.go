package rank

import (
	"fmt"
)

type RankData struct {
	MaxCount int
	RankList []*RankItem
	idPoints map[int64]int64
}

type RankItem struct {
	Id    int64
	Point int64
}

func NewRank(maxCount int) *RankData {
	rankData := &RankData{
		MaxCount: maxCount,
		RankList: make([]*RankItem, 0),
		idPoints: make(map[int64]int64),
	}

	return rankData
}

func (rankData *RankData) OnDataLoaded() {
	rankData.idPoints = make(map[int64]int64)
	for _, item := range rankData.RankList {
		rankData.idPoints[item.Id] = item.Point
	}
}

func (rankData *RankData) SetMaxRankCount(maxCount int) {
	rankData.MaxCount = maxCount
	if maxCount > 0 && len(rankData.RankList) > maxCount {
		for i := maxCount; i < len(rankData.RankList); i++ {
			data := rankData.RankList[i]
			delete(rankData.idPoints, data.Id)
		}
		rankData.RankList = rankData.RankList[0:maxCount]
	}
}

//GetIDRank 排名
func (rankData *RankData) GetIDRank(id int64) int {
	point, ok := rankData.idPoints[id]
	if !ok {
		return -1
	}

	rankList := rankData.RankList
	low, high := 0, len(rankList)-1
	for low <= high {
		mid := (low + high) / 2
		data := rankList[mid]
		if data.Id == id {
			return mid
		} else if data.Point == point {
			for i := mid + 1; i <= high; i++ {
				data := rankList[i]
				if data.Id == id {
					return i
				}
			}
			for i := mid - 1; i >= low; i-- {
				data := rankList[i]
				if data.Id == id {
					return i
				}
			}
		} else if data.Point < point {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	panic(fmt.Sprintf("rank index error: %d %d", id, point))
}

func (rankData *RankData) GetIdPoint(id int64) (int64, bool) {
	point, ok := rankData.idPoints[id]
	return point, ok
}

func (rankData *RankData) Insert(id, point int64) {
	value, ok := rankData.idPoints[id]
	if ok {
		if value == point {
			return
		}
		index := rankData.GetIDRank(id)
		rankData.RankList = append(rankData.RankList[0:index], rankData.RankList[index+1:]...)
		delete(rankData.idPoints, id)

		if point < value {
			rankData.insert(id, point, index, len(rankData.RankList)-1)
		} else {
			rankData.insert(id, point, 0, index-1)
		}
	} else {
		rankData.insert(id, point, 0, len(rankData.RankList)-1)
	}
}

func (rankData *RankData) insert(id, point int64, low, high int) {
	rankList := rankData.RankList
	for low <= high {
		mid := (low + high) / 2
		data := rankList[mid]
		if data.Point == point {
			for low = mid + 1; low <= high; low++ {
				data := rankList[low]
				if data.Point != point {
					break
				}
			}
		} else if data.Point < point {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	count := len(rankData.RankList)
	//队尾
	if low == count {
		if rankData.MaxCount > 0 && count >= rankData.MaxCount {
			return
		}
		rankData.RankList = append(rankData.RankList, &RankItem{Id: id, Point: point})
		rankData.idPoints[id] = point
		return
	}

	//队首
	if low == 0 {
		rankData.RankList = append([]*RankItem{&RankItem{Id: id, Point: point}}, rankData.RankList...)
	} else {
		tmp := append([]*RankItem{&RankItem{Id: id, Point: point}}, rankData.RankList[low:]...)
		rankData.RankList = append(rankData.RankList[0:low], tmp...)
	}
	rankData.idPoints[id] = point
	if rankData.MaxCount <= 0 || len(rankData.RankList) <= rankData.MaxCount {
		return
	}

	count = len(rankData.RankList) - 1
	item := rankData.RankList[count]
	delete(rankData.idPoints, item.Id)
	rankData.RankList = rankData.RankList[0:count]
}
