package edges

type Target struct {
	Value string
}

type (
	SliceType   []Target
	ArrayType   [5]Target
	MapType     map[string]Target
	ChanType    chan Target
	PointerType *Target
)

type StructType struct {
	A Target
	B *Target
	C []Target
}

type NestedType struct {
	A *map[string][]Target
}

type GenericType[T, U any] struct {
	T T
	U U
}

type InstantiatedType struct {
	G GenericType[Target, map[[2]Target][]Target]
}
