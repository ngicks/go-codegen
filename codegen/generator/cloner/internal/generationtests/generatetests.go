package generationtests

//go:generate go run ./_generate_test -e _generate_test,implementor
//go:generate go run ../../../../ cloner -v --chan-disallow --ignore-generated --dir ../testtargets --pkg ./...
