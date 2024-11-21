package codegen

import (
	"go/parser"
	"go/token"
	"testing"

	"gotest.tools/v3/assert"
)

var (
	testComments = `package main

import "fmt"

//codegen:ignore

//codegen:generated

// codegen:ignore

// codegen:generated

// aaaa
//codegen:ignore
//codegen:generated
// bbbb

/*
codegen:ignore
*/

/*
codegen:generated
*/

// not found

// codegen:generateddawa
`
	testCommentsParseResult = []undCommentParseResult{
		{UndDirection: Direction{ignore: true}},
		{UndDirection: Direction{generated: true}},
		{UndDirection: Direction{ignore: true}},
		{UndDirection: Direction{generated: true}},
		{UndDirection: Direction{ignore: true}},
		{UndDirection: Direction{ignore: true}},
		{UndDirection: Direction{generated: true}},
		{NotFound: true},
		{Err: true},
	}
)

type undCommentParseResult struct {
	Err          bool
	NotFound     bool
	UndDirection Direction
}

func TestDirective_Parse(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "hello.go", testComments, parser.ParseComments|parser.AllErrors)
	if err != nil {
		panic(err)
	}

	for i, cg := range f.Comments {
		d, found, err := ParseComment(cg)

		expected := testCommentsParseResult[i]
		if expected.Err {
			assert.Assert(t, err != nil)
		} else {
			assert.NilError(t, err)
		}
		assert.Equal(t, !expected.NotFound, found)
		assert.Equal(t, testCommentsParseResult[i].UndDirection, d)
	}
}
