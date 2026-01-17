package tytest

import (
	"go/types"
	"testing"

	"github.com/ngicks/go-codegen/graft/internal/loader"
	"gotest.tools/v3/assert"
)

// run go vet ./pkg/tytest/testdata/nocopy/
//
// root@4628d28d91d5:/git/github.com/ngicks/go-codegen/codegen# go vet ./matcher/testdata/nocopy/
// # github.com/ngicks/go-codegen/codegen/matcher/testdata/nocopy
// matcher/testdata/nocopy/nocopy.go:19:8: assignment copies lock value to d2: github.com/ngicks/go-codegen/codegen/matcher/testdata/nocopy.D contains sync/atomic.Int64 contains sync/atomic.noCopy
// matcher/testdata/nocopy/nocopy.go:24:8: assignment copies lock value to e2: sync/atomic.Int64 contains sync/atomic.noCopy
// matcher/testdata/nocopy/nocopy.go:27:8: assignment copies lock value to g2: github.com/ngicks/go-codegen/codegen/matcher/testdata/nocopy.G contains sync/atomic.Int64 contains sync/atomic.noCopy
// matcher/testdata/nocopy/nocopy.go:32:22: call of discard copies lock value: github.com/ngicks/go-codegen/codegen/matcher/testdata/nocopy.D contains sync/atomic.Int64 contains sync/atomic.noCopy
// matcher/testdata/nocopy/nocopy.go:32:26: call of discard copies lock value: sync/atomic.Int64 contains sync/atomic.noCopy
// matcher/testdata/nocopy/nocopy.go:32:30: call of discard copies lock value: github.com/ngicks/go-codegen/codegen/matcher/testdata/nocopy.G contains sync/atomic.Int64 contains sync/atomic.noCopy
func TestIsNoCopy(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/nocopy")
	pkg := pkgs[0].Types

	a := pkg.Scope().Lookup("A")
	assert.Assert(t, !IsNoCopy(a.Type()))
	assert.Assert(t, !IsNoCopy(a.Type().Underlying()))
	assert.Assert(t, IsNoCopy(a.Type().Underlying().(*types.Struct).Field(0).Type()))

	b := pkg.Scope().Lookup("B")
	assert.Assert(t, !IsNoCopy(b.Type()))
	assert.Assert(t, IsNoCopy(b.Type().Underlying().(*types.Struct).Field(0).Type()))

	c := pkg.Scope().Lookup("C")
	assert.Assert(t, IsNoCopy(c.Type()))

	d := pkg.Scope().Lookup("D")
	assert.Assert(t, IsNoCopy(d.Type()))

	e := pkg.Scope().Lookup("E")
	assert.Assert(t, IsNoCopy(e.Type()))

	f := pkg.Scope().Lookup("F")
	assert.Assert(t, IsNoCopy(f.Type()))

	g := pkg.Scope().Lookup("G")
	assert.Assert(t, IsNoCopy(g.Type()))

	h := pkg.Scope().Lookup("H")
	assert.Assert(t, !IsNoCopy(h.Type()))

	tree := pkg.Scope().Lookup("Tree")
	assert.Assert(t, !IsNoCopy(tree.Type()))
}
