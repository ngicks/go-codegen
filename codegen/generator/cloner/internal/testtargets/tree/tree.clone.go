// Code generated by github.com/ngicks/go-codegen/codegen DO NOT EDIT.
// to regenerate the code, refer to help by invoking
// go run github.com/ngicks/go-codegen/codegen cloner --help

package tree

//codegen:generated
func (v Tree[T]) CloneFunc(cloneT func(T) T) Tree[T] {
	return Tree[T]{
		node: func(v *node[T]) *node[T] {
			var out *node[T]

			if v != nil {
				out = new(node[T])
			}

			inner := out
			if v != nil {
				v := *v
				vv := v.CloneFunc(
					cloneT,
				)
				inner = &vv
			}
			out = inner

			return out
		}(v.node),
		comparer: v.comparer,
	}
}

//codegen:generated
func (v node[T]) CloneFunc(cloneT func(T) T) node[T] {
	return node[T]{
		l: func(v *node[T]) *node[T] {
			var out *node[T]

			if v != nil {
				out = new(node[T])
			}

			inner := out
			if v != nil {
				v := *v
				vv := v.CloneFunc(
					cloneT,
				)
				inner = &vv
			}
			out = inner

			return out
		}(v.l),
		r: func(v *node[T]) *node[T] {
			var out *node[T]

			if v != nil {
				out = new(node[T])
			}

			inner := out
			if v != nil {
				v := *v
				vv := v.CloneFunc(
					cloneT,
				)
				inner = &vv
			}
			out = inner

			return out
		}(v.r),
		data: cloneT(v.data),
	}
}
