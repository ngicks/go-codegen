package nocopy

import "sync"

type ContainsNoCopy struct {
	//cloner:copyptr
	NoCopy *sync.Mutex
	//cloner:copyptr
	NoCopyMap map[int]*sync.Mutex
	Ignored   *sync.Mutex
}
