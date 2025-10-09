package directive

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strconv"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"gotest.tools/v3/assert"
)

// Test utility function for parsing directive comments with Direction struct
func parseDirective(comments *ast.CommentGroup) (Direction, bool, error) {
	parsed, found := ParseAst(comments, DirectivePrefix)
	if !found {
		return Direction{}, false, nil
	}

	var dir Direction
	if err := parsed.Unmarshal(&dir); err != nil {
		return Direction{}, false, err
	}

	// Check for unknown directives
	for key := range parsed {
		if key != DirectiveCommentIgnore && key != DirectiveCommentGenerated {
			return dir, true, fmt.Errorf("unknown: %v", key)
		}
	}

	return dir, true, nil
}

// Test utility function for parsing DST directive comments with Direction struct
func parseDirectiveDst(comments dst.NodeDecs) (Direction, bool, error) {
	parsed, found := Parse(slices.Values(afterLastEmptyLine(comments.Start)), DirectivePrefix)
	if !found {
		return Direction{}, false, nil
	}

	var dir Direction
	if err := parsed.Unmarshal(&dir); err != nil {
		return Direction{}, false, err
	}

	// Check for unknown directives
	for key := range parsed {
		if key != DirectiveCommentIgnore && key != DirectiveCommentGenerated {
			return dir, true, fmt.Errorf("unknown: %v", key)
		}
	}

	return dir, true, nil
}

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
		{Direction: Direction{Ignore: true}},
		{Direction: Direction{Generated: true}},
		{Direction: Direction{Ignore: true}},
		{Direction: Direction{Generated: true}},
		{Direction: Direction{Ignore: true}},
		{Direction: Direction{Ignore: true}},
		{Direction: Direction{Generated: true}},
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
		d, found, err := parseDirective(cg)

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
type J struct{}`
	testDstCommentsParseResult = []directiveCommentParseResult{
		{Direction: Direction{Ignore: true}},    // 0
		{Direction: Direction{Generated: true}}, // 1
		{Direction: Direction{Ignore: true}},    // 2
		{Direction: Direction{Generated: true}}, // 3
		{Direction: Direction{Ignore: true}},    // 4
		{Direction: Direction{Ignore: true}},    // 5
		{Direction: Direction{Generated: true}}, // 6
		{NotFound: true},                        // 7
		{Err: true},                             // 8
		{Direction: Direction{Generated: true}}, // 9
		{Direction: Direction{Ignore: true}},    // 10
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
			d, found, err := parseDirectiveDst(*decl.Decorations())

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

var (
	testFieldDescDstSrc = `package target
type A struct {
	_ int
	// 1

	// 2

	// 3
	/* 4 */ A /* 5 */ string /* 6 */ ` + "`json:\"a\"`" + ` // 7
	/* 8 */
	// 9
	B                 string /* 10 */
	// 11
	C                 int
	// 12
	// 13
	_                 int
}
`
)

func Test_comment_parse_dst(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "hello.go", testFieldDescDstSrc, parser.ParseComments|parser.AllErrors)
	if err != nil {
		panic(err)
	}
	dec := decorator.NewDecorator(fset)
	df, err := dec.DecorateFile(f)
	if err != nil {
		panic(err)
	}

	var dts *dst.TypeSpec

L:
	for _, decl := range df.Decls {
		g, ok := decl.(*dst.GenDecl)
		if !ok {
			continue
		}
		if g.Tok != token.TYPE {
			continue
		}
		for _, s := range g.Specs {
			ts := s.(*dst.TypeSpec)
			if ts.Name.Name == "A" {
				dts = ts
				break L
			}
		}
	}
	if dts == nil {
		panic("not found")
	}

	st := dts.Type.(*dst.StructType)
	a := st.Fields.List[1]
	b := st.Fields.List[2]
	c := st.Fields.List[3]

	assert.DeepEqual(t, []string{"// 3", "/* 4 */"}, afterLastEmptyLine(a.Decs.Start))
	assert.DeepEqual(t, []string{"// 7"}, clip1(a.Decs.End))

	assert.DeepEqual(t, []string{"/* 8 */", "\n", "// 9"}, afterLastEmptyLine(b.Decs.Start))
	assert.DeepEqual(t, []string{"/* 10 */"}, clip1(b.Decs.End))

	assert.DeepEqual(t, []string{"// 11"}, afterLastEmptyLine(c.Decs.Start))
	assert.DeepEqual(t, []string(nil), clip1(c.Decs.End))
}