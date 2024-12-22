package clonepublicfieldonly

import "archive/tar"

// nolint
type exampleStruct struct {
	f1 tar.Header
	f2 Example2
	f3 ExampleNested
	f4 []ignored
	f5 ExampleCloner
}

// nolint
//
//codegen:ignore
type Example2 struct {
	A map[string]bool
	b map[string]bool
}

//codegen:ignore
type ExampleNested struct {
	A tar.Header
}

// don't make type clone-by-assign
//
// nolint
//
//codegen:ignore
type ignored struct {
	foo []string
}

//codegen:ignore
type ExampleCloner struct {
	A ExampleClonerImpl
}

//codegen:ignore
type ExampleClonerImpl struct{}

func (ExampleClonerImpl) Clone() ExampleClonerImpl {
	return ExampleClonerImpl{}
}

// nolint
type exampleMap map[string]tar.Header
