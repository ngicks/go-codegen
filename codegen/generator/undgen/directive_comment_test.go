package undgen

import (
	"go/parser"
	"go/token"
	"testing"

	"gotest.tools/v3/assert"
)

var (
	testComments = `package main

import "fmt"

//undgen:ignore

//undgen:generated

// undgen:ignore

// undgen:generated

// aaaa
//undgen:ignore
//undgen:generated
// bbbb

/*
undgen:ignore
*/

/*
undgen:generated
*/

// not found

// undgen:generateddawa
`
	testCommentsParseResult = []undCommentParseResult{
		{UndDirection: UndDirection{ignore: true}},
		{UndDirection: UndDirection{generated: true}},
		{UndDirection: UndDirection{ignore: true}},
		{UndDirection: UndDirection{generated: true}},
		{UndDirection: UndDirection{ignore: true}},
		{UndDirection: UndDirection{ignore: true}},
		{UndDirection: UndDirection{generated: true}},
		{NotFound: true},
		{Err: true},
	}
)

type undCommentParseResult struct {
	Err          bool
	NotFound     bool
	UndDirection UndDirection
}

func TestDirective_Parse(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "hello.go", testComments, parser.ParseComments|parser.AllErrors)
	if err != nil {
		panic(err)
	}

	for i, cg := range f.Comments {
		d, found, err := ParseUndComment(cg)

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
