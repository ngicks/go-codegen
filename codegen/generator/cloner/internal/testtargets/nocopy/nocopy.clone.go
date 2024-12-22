// Code generated by github.com/ngicks/go-codegen/codegen DO NOT EDIT.
// to regenerate the code, refer to help by invoking
// go run github.com/ngicks/go-codegen/codegen cloner --help

package nocopy

import (
	"sync"
)

//codegen:generated
func (v ContainsNoCopy) Clone() ContainsNoCopy {
	return ContainsNoCopy{
		NoCopy: v.NoCopy,
		NoCopyMap: func(v map[int]*sync.Mutex) map[int]*sync.Mutex {
			var out map[int]*sync.Mutex

			if v != nil {
				out = make(map[int]*sync.Mutex, len(v))
			}

			inner := out
			for k, v := range v {
				inner[k] = v
			}
			out = inner

			return out
		}(v.NoCopyMap),
		Ignored: v.Ignored,
		C:       v.C,
		CC: func(v map[string]chan int) map[string]chan int {
			var out map[string]chan int

			if v != nil {
				out = make(map[string]chan int, len(v))
			}

			inner := out
			for k, v := range v {
				inner[k] = v
			}
			out = inner

			return out
		}(v.CC),
		CS: func(v []chan int) []chan int {
			var out []chan int

			if v != nil {
				out = make([]chan int, len(v), cap(v))
			}

			inner := out
			for k, v := range v {
				inner[k] = make(chan int, cap(v))
			}
			out = inner

			return out
		}(v.CS),
		NamedFunc: v.NamedFunc,
	}
}
