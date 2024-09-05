package util

import (
	"cmp"
	"slices"
)

// SortedEqual returns true if a and b are equal if sorted
func SortedEqual[A ~[]S, S cmp.Ordered](a, b A) bool {
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

// Map returns mapping of arr using mapFn
func Map[A ~[]S, S interface{}, V interface{}](arr A, mapFn func(S) V) []V {
	var ret []V
	for _, x := range arr {
		ret = append(ret, mapFn(x))
	}
	return ret
}

// Diff returns elements in a not in b and vice versa
func Diff[A ~[]S, S comparable](a, b A) (A, A) {
	var retA, retB A
	mapA := make(map[S]interface{})
	mapB := make(map[S]interface{})
	for _, k := range a {
		mapA[k] = nil
	}
	for _, k := range b {
		mapB[k] = nil
	}
	for _, k := range a {
		if _, ok := mapB[k]; !ok {
			retA = append(retA, k)
		}
	}
	for _, k := range b {
		if _, ok := mapA[k]; !ok {
			retB = append(retB, k)
		}
	}

	return retA, retB
}

// Select returns elements of arr where cmp(element)==true
func Select[A ~[]S, S interface{}](arr A, cmp func(S) bool) A {
	var ret A
	for _, v := range arr {
		if cmp(v) {
			ret = append(ret, v)
		}
	}
	return ret
}
