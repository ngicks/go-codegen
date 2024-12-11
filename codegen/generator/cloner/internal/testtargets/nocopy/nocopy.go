package nocopy

import "sync"

type ContainsNoCopy struct {
	//cloner:copyptr
	NoCopy *sync.Mutex
	//cloner:copyptr
	NoCopyMap map[int]*sync.Mutex
	Ignored   *sync.Mutex
	//cloner:copyptr
	C chan int
	//cloner:copyptr
	CC map[string]chan int
	//cloner:make
	CS []chan int

	//cloner:copyptr
	NamedFunc namedFunc
}

type namedFunc func()
