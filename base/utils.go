package base

import (
	"math/rand"
	"reflect"
)

func Rand(start, end int) int {
	if start >= end {
		return end
	}

	return start + rand.Intn(end-start+1)
}

func Max(data ...float64) float64 {
	max := data[0]
	for _, v := range data {
		if max < v {
			max = v
		}
	}

	return max
}

func Min(data ...float64) float64 {
	min := data[0]
	for _, v := range data {
		if min > v {
			min = v
		}
	}

	return min
}

func RandN(n int, datas ...interface{}) []interface{} {
	Len := len(datas)
	if n >= Len {
		results := make([]interface{}, Len)
		for i, v := range rand.Perm(Len) {
			results[i] = datas[v]
		}
		return results
	}

	results := make([]interface{}, n)
	for i, v := range rand.Perm(Len) {
		results[i] = datas[v]
		if i+1 == n {
			break
		}
	}
	return results
}

func RandSliceN(n int, data interface{}) []interface{} {
	values := reflect.ValueOf(data)

	if values.Kind() != reflect.Slice {
		panic("data must a slice")
	}

	Len := values.Len()
	if n >= Len {
		results := make([]interface{}, Len)
		for index, i := range rand.Perm(Len) {
			results[index] = values.Index(i).Interface()
		}
		return results
	}

	results := make([]interface{}, n)
	for index, i := range rand.Perm(Len) {
		results[index] = values.Index(i).Interface()
		if index+1 == n {
			break
		}
	}
	return results
}

func ReflectFunc(cbFunc interface{}, args []interface{}) (reflect.Value, []reflect.Value) {
	cb := reflect.ValueOf(cbFunc)
	count := len(args)

	var values []reflect.Value
	if count > 0 {
		values = make([]reflect.Value, count)
		for i, v := range args {
			values[i] = reflect.ValueOf(v)
		}
	}

	return cb, values
}

func ReverseKeyValue(a map[int]int) map[int]int {
	b := make(map[int]int)
	for k, v := range a {
		b[v] = k
	}

	return b
}
