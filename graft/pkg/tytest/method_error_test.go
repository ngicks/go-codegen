package tytest

import (
	"go/types"
	"testing"

	"github.com/ngicks/go-codegen/graft/internal/loader"
	"gotest.tools/v3/assert"
)

func TestErrorMethod(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/error")
	pkg := pkgs[0].Types

	a := pkg.Scope().Lookup("A").Type().(*types.Named)

	aMethod := ErrorMethod{Name: "A"}
	bMethod := ErrorMethod{Name: "B"}
	cMethod := ErrorMethod{Name: "C"}
	dMethod := ErrorMethod{Name: "D"}

	assert.Assert(t, aMethod.IsImplementor(a))
	assert.Assert(t, bMethod.IsImplementor(a))
	assert.Assert(t, !cMethod.IsImplementor(a))
	assert.Assert(t, !dMethod.IsImplementor(a))
}
