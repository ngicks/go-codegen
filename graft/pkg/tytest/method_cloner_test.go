package tytest

import (
	"go/types"
	"testing"

	"github.com/ngicks/go-codegen/graft/internal/loader"
	"github.com/ngicks/go-iterator-helper/hiter"
	"gotest.tools/v3/assert"
)

func TestClonerMethod(t *testing.T) {
	pkgs := loader.LoadPackagesTest(t, "./testdata/cloner")
	pkg := pkgs[0].Types

	method := ClonerMethod{Name: "Clone"}

	c := pkg.Scope().Lookup("C")
	assert.Assert(t, method.IsImplementor(c.Type()))
	cp := pkg.Scope().Lookup("CP")
	assert.Assert(t, method.IsImplementor(cp.Type()))
	param := pkg.Scope().Lookup("Param")
	assert.Assert(t, !method.IsImplementor(param.Type()))
	assert.Assert(t, method.IsFuncImplementor(param.Type()))

	total := pkg.Scope().Lookup("Total").Type().(*types.Named).Underlying().(*types.Struct)
	for i := range hiter.Range(0, 3) {
		assert.Assert(t, method.IsImplementor(total.Field(i).Type()))
	}
	for i := range hiter.Range(3, 4) {
		assert.Assert(t, !method.IsImplementor(total.Field(i).Type()))
		assert.Assert(t, method.IsFuncImplementor(total.Field(i).Type()))
	}
}
