package tests

import (
	"cmp"
	"slices"
	"testing"

	"github.com/ngicks/go-codegen/codegen/generator/cloner/internal/testtargets/tree"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"gotest.tools/v3/assert"
)

func TestTree(t *testing.T) {
	org := tree.New[int](cmp.Compare)

	dataSet := []int{5, 7, 7, 2, 34, 4, 6, 7, 8, 89, 12, 3, 68, 3, 12, 2}

	org.Push(dataSet...)

	sorted := slices.Sorted(slices.Values(dataSet))
	assert.DeepEqual(t, sorted, slices.Collect(org.All()))

	cloned := org.CloneFunc(func(i int) int { return i })

	assert.DeepEqual(t, sorted, slices.Collect(org.All()))

	assert.DeepEqual(t, sorted, slices.Collect(cloned.All()))
	additional := []int{5, 7, 9, 3}
	cloned.Push(additional...)
	assert.DeepEqual(t, sorted, slices.Collect(org.All()))
	assert.DeepEqual(t, slices.Sorted(xiter.Concat(slices.Values(dataSet), slices.Values(additional))), slices.Collect(cloned.All()))
}
