package generationtests

//go:generate go run -race ./_generate_test -e _generate_test,implementor
//go:generate go run -race ../../../../ cloner -v --chan-disallow --ignore-generated --dir ../testtargets --pkg ./...
