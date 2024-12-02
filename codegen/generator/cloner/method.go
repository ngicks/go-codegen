package cloner

import (
	"io"

	"github.com/ngicks/go-codegen/codegen/codegen"
	"github.com/ngicks/go-codegen/codegen/typegraph"
)

func generateMethod(w io.Writer, node *typegraph.Node) (err error) {
	printf, flush := codegen.BufPrintf(w)
	defer func() {
		fErr := flush()
		if err == nil {
			err = fErr
		}
	}()

	if node.Type.TypeParams().Len() == 0 {
		err = generateCloner(printf, node)
	} else {
		err = generateClonerFunc(printf, node)
	}
	return
}

func generateCloner(printf func(format string, args ...any), node *typegraph.Node) error {
	return nil
}

func generateClonerFunc(printf func(format string, args ...any), node *typegraph.Node) error {
	return nil
}
