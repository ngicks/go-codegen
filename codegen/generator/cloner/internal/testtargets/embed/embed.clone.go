// Code generated by github.com/ngicks/go-codegen/codegen DO NOT EDIT.
// to regenerate the code, refer to help by invoking
// go run github.com/ngicks/go-codegen/codegen cloner --help

package embed

//codegen:generated
func (v A) Clone() A {
	return A{
		Embed: func(v Embed) Embed {
			return Embed{
				Foo: v.Foo,
				Bar: func(src []int) []int {
					if src == nil {
						return nil
					}
					dst := make([]int, len(src), cap(src))
					copy(dst, src)
					return dst
				}(v.Bar),
			}
		}(v.Embed),
	}
}

//codegen:generated
func (v B) Clone() B {
	return B{
		F1: v.F1,
		Embed: func(v Embed) Embed {
			return Embed{
				Foo: v.Foo,
				Bar: func(src []int) []int {
					if src == nil {
						return nil
					}
					dst := make([]int, len(src), cap(src))
					copy(dst, src)
					return dst
				}(v.Bar),
			}
		}(v.Embed),
		F2: v.F2,
	}
}

//codegen:generated
func (v C) Clone() C {
	return C{
		EmbedImplementor: v.EmbedImplementor.Clone(),
	}
}

//codegen:generated
func (v D) Clone() D {
	return D{
		F1:               v.F1,
		EmbedImplementor: v.EmbedImplementor.Clone(),
		F2:               v.F2,
	}
}
