package input

import "sync"

// SimpleStruct is a basic struct for testing automark/autoimpl workflow
type SimpleStruct struct {
	Name  string
	Value int
}

// StructWithPointer contains pointer fields
type StructWithPointer struct {
	Data  *string
	Count *int
}

// StructWithSlice contains slice fields
type StructWithSlice struct {
	Items   []string
	Numbers []int
}

// StructWithMap contains map fields
type StructWithMap struct {
	Metadata map[string]string
	Counts   map[string]int
}

// StructWithNoCopy contains no-copy types like sync.Mutex
type StructWithNoCopy struct {
	Name  string
	Mutex sync.Mutex
}

// StructWithChannel contains channel fields
type StructWithChannel struct {
	Name   string
	Events chan string
}

// GenericStruct is a generic type for testing
type GenericStruct[T any] struct {
	Value T
	List  []T
}

// NestedStruct contains other structs
type NestedStruct struct {
	Simple  SimpleStruct
	Pointer *SimpleStruct
}
