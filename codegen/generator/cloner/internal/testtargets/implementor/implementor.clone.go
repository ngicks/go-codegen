// Code generated by github.com/ngicks/go-codegen/codegen DO NOT EDIT.
// to regenerate the code, refer to help by invoking
// go run github.com/ngicks/go-codegen/codegen cloner --help

package implementor

//codegen:generated
func (v ContainsImplementor[T]) CloneFunc(cloneT func(T) T) ContainsImplementor[T] {
	return ContainsImplementor[T]{
		U: v.U.CloneFunc(
			cloneT,
		),
		US: v.US.CloneFunc(
			func(v string) string {
				return v
			},
		),
		W: v.W.Clone(),
	}
}
