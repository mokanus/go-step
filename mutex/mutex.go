package mutex

import (
	"sync"
)

type Ref struct {
	*sync.Mutex
	Count int
}

var (
	refMap       = make(map[string]*Ref)
	refMapLocker = new(sync.Mutex)
)

func Lock(id string) {
	refMapLocker.Lock()
	ref := refMap[id]
	if ref == nil {
		ref = &Ref{
			Mutex: new(sync.Mutex),
			Count: 1,
		}
		refMap[id] = ref
	} else {
		ref.Count++
	}
	refMapLocker.Unlock()
	ref.Lock()
}

func Unlock(id string) {
	refMapLocker.Lock()
	ref := refMap[id]
	if ref != nil {
		ref.Unlock()
		ref.Count--
		if ref.Count <= 0 {
			delete(refMap, id)
		}
	}
	refMapLocker.Unlock()
}
