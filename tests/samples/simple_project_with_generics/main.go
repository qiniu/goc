//go:build go1.18

package main

import (
	"fmt"
	"strconv"
)

func main() {
	x := []int{1, 3, 5, 7, 9}
	y := mapSlice(func(x int) string {
		return strconv.Itoa(x * x)
	}, x)

	for i, elem := range x {
		fmt.Printf("%d -> '%s'\n", elem, y[i])
	}
}

func mapSlice[T any, U any](fn func(T) U, src []T) []U {
	dst := make([]U, len(src))
	for i, x := range src {
		dst[i] = fn(x)
	}
	return dst
}
