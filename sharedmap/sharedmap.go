package sharedmap

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/joeshaw/gengen/generic"
)

type Copiable interface {
	copy() *Copiable
}

type SharedMap struct {
	writeLock sync.Mutex
	current   unsafe.Pointer
}

func newSharedMap() *SharedMap {
	newMap := new(SharedMap)
	underlyingMap := make(map[generic.T]generic.V)
	newMap.current = unsafe.Pointer(&underlyingMap)
	return newMap
}

func (m *SharedMap) Get(key generic.T) (generic.V, bool) {
	val, present := (*(*map[generic.T]generic.V)(m.current))[key]
	return val, present
}

func (m *SharedMap) GetOrSet(key generic.T, valueMaker func() generic.V) (generic.V, bool) {
	m.writeLock.Lock()
	defer m.writeLock.Unlock()
	// Have to do .Get inside the lock incase another writer snuck the entry
	// in there at the last moment.
	existing, there := m.Get(key)
	if there {
		return existing, false
	}
	newCurrent := m.Copy()
	newValue := valueMaker()
	(*newCurrent)[key] = newValue
	atomic.StorePointer(&m.current, unsafe.Pointer(newCurrent))
	return newValue, true
}

func (m *SharedMap) Copy() *map[generic.T]generic.V {
	current := (*(*map[generic.T]generic.V)(m.current))
	newMap := make(map[generic.T]generic.V)
	for key, value := range current {
		newMap[key] = value
	}
	return &newMap
}

func (m *SharedMap) SetValue(key generic.T, value generic.V) {
	m.writeLock.Lock()
	defer m.writeLock.Unlock()
	newCurrent := m.Copy()
	(*newCurrent)[key] = value
	atomic.StorePointer(&m.current, unsafe.Pointer(newCurrent))
}
