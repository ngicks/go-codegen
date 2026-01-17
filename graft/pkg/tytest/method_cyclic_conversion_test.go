package tytest

import (
	"go/types"
	"testing"

	"github.com/ngicks/go-codegen/graft/internal/loader"
	"gotest.tools/v3/assert"
)

func TestCyclicConversionMethods(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/cyclic_conversion")
	pkg := pkgs[0].Types

	cmset := CyclicConversionMethods{
		Reverse: "UndRaw",
		Convert: "UndPlain",
	}
	a := pkg.Scope().Lookup("A")
	assert.Assert(t, cmset.IsImplementor(a.Type().(*types.Named)))
	b := pkg.Scope().Lookup("B")
	cmsetRev := cmset
	cmsetRev.From = true
	assert.Assert(t, cmsetRev.IsImplementor(b.Type().(*types.Named)))
	ap := pkg.Scope().Lookup("AP")
	assert.Assert(t, cmset.IsImplementor(ap.Type().(*types.Named)))
	ai := pkg.Scope().Lookup("AI")
	assert.Assert(t, cmset.IsImplementor(ai.Type().(*types.Named)))
	notImplementor := pkg.Scope().Lookup("NotImplementor")
	assert.Assert(t, !cmset.IsImplementor(notImplementor.Type().(*types.Named)))
}
