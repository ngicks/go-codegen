package codegen

import (
	"go/parser"
	"go/token"
	"strconv"
	"testing"

	"github.com/dave/dst/decorator"
	"gotest.tools/v3/assert"
)

type directiveCommentParseResult struct {
	Err       bool
	NotFound  bool
	Direction Direction
}

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
	testCommentsParseResult = []directiveCommentParseResult{
		{Direction: Direction{ignore: true}},
		{Direction: Direction{generated: true}},
		{Direction: Direction{ignore: true}},
		{Direction: Direction{generated: true}},
		{Direction: Direction{ignore: true}},
		{Direction: Direction{ignore: true}},
		{Direction: Direction{generated: true}},
		{NotFound: true},
		{Err: true},
	}
)

func TestDirective_ast(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "hello.go", testComments, parser.ParseComments|parser.AllErrors)
	if err != nil {
		panic(err)
	}

	for i, cg := range f.Comments {
		d, found, err := ParseDirectiveComment(cg)

		expected := testCommentsParseResult[i]
		if expected.Err {
			assert.Assert(t, err != nil)
		} else {
			assert.NilError(t, err)
		}
		assert.Equal(t, !expected.NotFound, found)
		assert.Equal(t, testCommentsParseResult[i].Direction, d)
	}
}

var (
	testDstComments = `package main

import "fmt"

//codegen:ignore
type A struct{}

//codegen:generated
func B() {}

// codegen:ignore
func C() {}

// codegen:generated
func D() {}

// aaaa
//codegen:ignore
//codegen:generated
// bbbb
func (a A) Sample() {}

/*
codegen:ignore
*/
type E struct{}

/*
codegen:generated
*/
type F struct{}

// not found
type G struct{}

// codegen:generateddawa
type H struct{}

//codegen:ignored

// foo
//codegen:generated
//codegen:ignored
// bar
type I struct{}


/*
codegen:generated
*/

/*
codegen:ignore
*/
type J struct{}
`
	testDstCommentsParseResult = []directiveCommentParseResult{
		{Direction: Direction{ignore: true}},    // 0
		{Direction: Direction{generated: true}}, // 1
		{Direction: Direction{ignore: true}},    // 2
		{Direction: Direction{generated: true}}, // 3
		{Direction: Direction{ignore: true}},    // 4
		{Direction: Direction{ignore: true}},    // 5
		{Direction: Direction{generated: true}}, // 6
		{NotFound: true},                        // 7
		{Err: true},                             // 8
		{Direction: Direction{generated: true}}, // 9
		{Direction: Direction{ignore: true}},    // 10
	}
)

func TestDirective_dst(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "hello.go", testDstComments, parser.ParseComments|parser.AllErrors)
	if err != nil {
		panic(err)
	}
	dec := decorator.NewDecorator(fset)
	df, err := dec.DecorateFile(f)
	if err != nil {
		panic(err)
	}

	for i, decl := range df.Decls[1:] {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			d, found, err := ParseDirectiveCommentDst(*decl.Decorations())

			expected := testDstCommentsParseResult[i]
			if expected.Err {
				assert.Assert(t, err != nil)
			} else {
				assert.NilError(t, err)
			}
			assert.Equal(t, !expected.NotFound, found)
			assert.Equal(t, testDstCommentsParseResult[i].Direction, d)
		})
	}
}
