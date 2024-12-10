// Code generated by github.com/ngicks/go-codegen/codegen DO NOT EDIT.
// to regenerate the code, refer to help by invoking
// go run github.com/ngicks/go-codegen/codegen cloner --help
package nocopy

import "sync"

//codegen:generated
func (v ContainsNoCopy) Clone() ContainsNoCopy {
	return ContainsNoCopy{
		NoCopy: v.NoCopy,
		NoCopyMap: func(v map[int]*sync.Mutex) map[int]*sync.Mutex {
			out := make(map[int]*sync.Mutex, len(v))

			inner := out
			for k, v := range v {
				inner[k] = v
			}

			return out
		}(v.NoCopyMap),
	}
}
