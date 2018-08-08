package gconfig

import (
	"fmt"
	"time"

	"github.com/sencydai/gameworld/base"
	c "github.com/sencydai/gameworld/constdefine"
	t "github.com/sencydai/gameworld/typedefine"
)

func ParseTime(timeType int, timeSubType int, time interface{}) interface{} {
	switch timeType {
	case c.TimeTypeUnlimit:
		return ""
	case c.TimeTypeWeek:
		if data, err := json.Marshal(time); err != nil {
			panic(err)
		} else {
			timeWeek := t.TimeWeek{}
			if err = json.Unmarshal(data, &timeWeek); err != nil {
				panic(err)
			}

			if data, err = json.Marshal(timeWeek.Week); err != nil {
				panic(err)
			}

			switch timeSubType {
			case c.TimeSubDay:
				w := make(map[int]int)
				if err = json.Unmarshal(data, &w); err != nil {
					panic(err)
				}
				timeWeek.Week = base.ReverseKeyValue(w)
			case c.TimeSubLast:
				w := make(map[int]map[int]int)
				if err = json.Unmarshal(data, &w); err != nil {
					panic(err)
				}
				timeWeek.Week = w
			default:
				panic(fmt.Errorf("error time sub type %d", timeSubType))
			}
			return timeWeek
		}
	case c.TimeTypeOpenServer:
		if data, err := json.Marshal(time); err != nil {
			panic(err)
		} else {
			openserver := t.TimeOpenServer{}
			if err = json.Unmarshal(data, &openserver); err != nil {
				panic(err)
			}

			return openserver
		}
	case c.TimeTypeFixed:
		if data, err := json.Marshal(time); err != nil {
			panic(err)
		} else {
			timeFixed := t.TimeFixed{}
			if err = json.Unmarshal(data, &timeFixed); err != nil {
				panic(err)
			}

			return timeFixed
		}
	default:
		return ""
	}
}

func CheckTimeStatus(now time.Time, timeType, timeSubtype int, timeData interface{}) (int, int, int) {
	switch timeType {
	case c.TimeTypeUnlimit:
		return c.TimeStatusUnlimit, 0, 0
	case c.TimeTypeWeek:
		timeWeek := timeData.(t.TimeWeek)
		weekday := base.WeekDay(now)
		switch timeSubtype {
		case c.TimeSubDay:
			y, m, d := now.Year(), now.Month(), now.Day()
			start := base.Date(y, m, d, timeWeek.SHour, timeWeek.SMin, timeWeek.SSec)
			//now < start
			if now.Before(start) {
				return c.TimeStatusOutside, int(start.Sub(now) / time.Second), 0
			}
			end := base.Date(y, m, d, timeWeek.EHour, timeWeek.EMin, timeWeek.ESec)
			//now >= end
			if !now.Before(end) {
				return c.TimeStatusOutside, int((base.Day - now.Sub(start)) / time.Second), 0
			}

			weeks := timeWeek.Week.(map[int]int)
			if _, ok := weeks[weekday]; !ok {
				return c.TimeStatusOutside, int((base.Day - now.Sub(start)) / time.Second), 0
			}
			return c.TimeStatusInRange, int(start.Unix()), int(end.Sub(now) / time.Second)
		case c.TimeSubLast:
			var startWeek, endWeek int
			for _, weeks := range timeWeek.Week.(map[int]map[int]int) {
				if weekday >= weeks[1] && weekday <= weeks[2] {
					startWeek, endWeek = weeks[1], weeks[2]
					break
				}
			}
			if startWeek == 0 && endWeek == 0 {
				y, m, d := now.Year(), now.Month(), now.Day()
				start := base.Date(y, m, d, timeWeek.SHour, timeWeek.SMin, timeWeek.SSec)
				//now < start
				if now.Before(start) {
					return c.TimeStatusOutside, int((base.Day + start.Sub(now)) / time.Second), 0
				}
				return c.TimeStatusOutside, int((base.Day - now.Sub(start)) / time.Second), 0
			}
			start := now.AddDate(0, 0, startWeek-weekday)
			start = base.Date(start.Year(), start.Month(), start.Day(), timeWeek.SHour, timeWeek.SMin, timeWeek.SSec)
			//now < start
			if now.Before(start) {
				return c.TimeStatusOutside, int(start.Sub(now) / time.Second), 0
			}
			end := now.AddDate(0, 0, endWeek-weekday)
			end = base.Date(end.Year(), end.Month(), end.Day(), timeWeek.EHour, timeWeek.EMin, timeWeek.ESec)
			//now >= end
			if !now.Before(end) {
				y, m, d := now.Year(), now.Month(), now.Day()
				start := base.Date(y, m, d, timeWeek.SHour, timeWeek.SMin, timeWeek.SSec)
				//now < start
				if now.Before(start) {
					return c.TimeStatusOutside, int((base.Day + start.Sub(now)) / time.Second), 0
				}
				return c.TimeStatusOutside, int((base.Day - now.Sub(start)) / time.Second), 0
			}

			return c.TimeStatusInRange, int(start.Unix()), int(end.Sub(now) / time.Second)
		}
	case c.TimeTypeOpenServer:
		timeOpen := timeData.(t.TimeOpenServer)
		openServer := t.GetSysOpenServerTime()
		y, m, d := openServer.Year(), openServer.Month(), openServer.Day()
		end := base.Date(y, m, d, timeOpen.EHour, timeOpen.EMin, timeOpen.ESec)
		end = end.AddDate(0, 0, timeOpen.EDay-1)
		//now >= end
		if !now.Before(end) {
			return c.TimeStatusExpire, 0, 0
		}
		start := base.Date(y, m, d, timeOpen.SHour, timeOpen.SMin, timeOpen.SSec)
		start = start.AddDate(0, 0, timeOpen.SDay-1)
		//now < start
		if now.Before(start) {
			return c.TimeStatusOutside, int(start.Sub(now) / time.Second), 0
		}

		switch timeSubtype {
		case c.TimeSubDay:
			y, m, d := now.Year(), now.Month(), now.Day()
			start := base.Date(y, m, d, timeOpen.SHour, timeOpen.SMin, timeOpen.SSec)
			//now < start
			if now.Before(start) {
				return c.TimeStatusOutside, int(start.Sub(now) / time.Second), 0
			}
			end := base.Date(y, m, d, timeOpen.EHour, timeOpen.EMin, timeOpen.ESec)
			//now >= end
			if !now.Before(end) {
				return c.TimeStatusOutside, int((base.Day - now.Sub(start)) / time.Second), 0
			}
			return c.TimeStatusInRange, int(start.Unix()), int(end.Sub(now) / time.Second)
		case c.TimeSubLast:
			return c.TimeStatusInRange, int(start.Unix()), int(end.Sub(now) / time.Second)
		}
	case c.TimeTypeFixed:
		timeFixed := timeData.(t.TimeFixed)
		end := base.Date(timeFixed.EYear, time.Month(timeFixed.EMonth), timeFixed.EDay, timeFixed.EHour, timeFixed.EMin, timeFixed.ESec)
		//now >= end
		if !now.Before(end) {
			return c.TimeStatusExpire, 0, 0
		}
		start := base.Date(timeFixed.SYear, time.Month(timeFixed.SMonth), timeFixed.SDay, timeFixed.SHour, timeFixed.SMin, timeFixed.SSec)
		//now < start
		if now.Before(start) {
			return c.TimeStatusOutside, int(start.Sub(now) / time.Second), 0
		}
		switch timeSubtype {
		case c.TimeSubDay:
			y, m, d := now.Year(), now.Month(), now.Day()
			start := base.Date(y, m, d, timeFixed.SHour, timeFixed.SMin, timeFixed.SSec)
			//now < start
			if now.Before(start) {
				return c.TimeStatusOutside, int(start.Sub(now) / time.Second), 0
			}
			end := base.Date(y, m, d, timeFixed.EHour, timeFixed.EMin, timeFixed.ESec)
			//now >= end
			if !now.Before(end) {
				return c.TimeStatusOutside, int((base.Day - now.Sub(start)) / time.Second), 0
			}
			return c.TimeStatusInRange, int(start.Unix()), int(end.Sub(now) / time.Second)
		case c.TimeSubLast:
			return c.TimeStatusInRange, int(start.Unix()), int(end.Sub(now) / time.Second)
		}
	}
	return -1, 0, 0
}
