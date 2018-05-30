package base

import (
	"math/rand"
	"reflect"
)

func Rand(start, end int) int {
	if start >= end {
		return end
	}

	return start + int(rand.Int31n(int32(end-start+1)))
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
