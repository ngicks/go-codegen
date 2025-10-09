package edges

import "github.com/ngicks/go-codegen/codegen/pkg/typegraph/testdata/faketarget"

type (
	MereArray  [5]faketarget.FakeTarget2[string, *MereChan]
	MereSlice  []faketarget.FakeTarget
	MereMap    map[string]faketarget.FakeTarget2[int, MereChan]
	MereChan   chan faketarget.FakeTarget
	MereStruct struct {
		A *faketarget.FakeTarget
		B faketarget.FakeTarget2[int, int]
	}
)

type Complex struct {
	A *map[string][]*[3]map[int]MereArray
}
