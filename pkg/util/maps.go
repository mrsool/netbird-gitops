package util

// SliceToMap returns map[keyFn(v)] = v for each v in arr
func SliceToMap[K ~[]S, S interface{}, V comparable](arr K, keyFn func(S) V) map[V]S {
	ret := make(map[V]S)
	for _, v := range arr {
		ret[keyFn(v)] = v
	}
	return ret
}
