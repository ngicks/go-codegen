// Code generated by github.com/ngicks/go-codegen/codegen DO NOT EDIT.
// to regenerate the code, refer to help by invoking
// go run github.com/ngicks/go-codegen/codegen cloner --help
package customcloner

import (
	"time"

	"maps"

	"github.com/ngicks/und"
)

//codegen:generated
func (v Custom) Clone() Custom {
	return Custom{
		T: func(t time.Time) time.Time {
			return time.Date(
				t.Year(),
				t.Month(),
				t.Day(),
				t.Hour(),
				t.Minute(),
				t.Second(),
				t.Nanosecond(),
				t.Location(),
			)
		}(v.T),
		TM: func(v map[string]time.Time) map[string]time.Time {
			out := make(map[string]time.Time, len(v))

			inner := out
			for k, v := range v {
				inner[k] = func(t time.Time) time.Time {
					return time.Date(
						t.Year(),
						t.Month(),
						t.Day(),
						t.Hour(),
						t.Minute(),
						t.Second(),
						t.Nanosecond(),
						t.Location(),
					)
				}(v)
			}
			out = inner

			return out
		}(v.TM),
		M: maps.Clone(v.M),
		B: func(src []byte) []byte {
			if src == nil {
				return nil
			}
			dst := make([]byte, len(src), cap(src))
			copy(dst, src)
			return dst
		}(v.B),
		Implementor: func(v [][]und.Und[string]) [][]und.Und[string] {
			out := make([][]und.Und[string], len(v))

			inner := out
			for k, v := range v {
				outer := &inner
				inner := make([]und.Und[string], len(v))
				for k, v := range v {
					inner[k] = v.CloneFunc(
						func(v string) string {
							return v
						},
					)
				}
				(*outer)[k] = inner
			}
			out = inner

			return out
		}(v.Implementor),
		TypeParam: v.TypeParam.CloneFunc(
			func(src []string) []string {
				if src == nil {
					return nil
				}
				dst := make([]string, len(src), cap(src))
				copy(dst, src)
				return dst
			},
		),
		H: v.H,
	}
}
