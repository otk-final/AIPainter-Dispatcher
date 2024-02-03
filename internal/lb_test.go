package internal

import (
	"cmp"
	"slices"
	"testing"
)

func TestLoad(t *testing.T) {

	arr := []int{1, 23, 3}

	slices.SortFunc(arr, func(a, b int) int {
		return cmp.Compare(a, b)
	})

	t.Log(arr)
}
