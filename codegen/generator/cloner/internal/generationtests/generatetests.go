package generationtests

//go:generate go run -race ./_generate_test -e _generate_test,implementor
//go:generate go run -race github.com/ngicks/go-codegen/codegen cloner -v --chan-disallow --ignore-generated --dir ../testtargets --pkg ./...
