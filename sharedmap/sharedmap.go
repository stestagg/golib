// Template SharedMap type

package sharedmap

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// template type SharedMap(KeyType, ValType)
type KeyType string
type ValType interface{}

type SharedMap struct {
	writeLock sync.Mutex
	current   unsafe.Pointer
}

func NewSharedMap() *SharedMap {
	newMap := new(SharedMap)
	underlyingMap := make(map[KeyType]ValType)
	newMap.current = unsafe.Pointer(&underlyingMap)
	return newMap
}

func (m *SharedMap) Get(key KeyType) (ValType, bool) {
	val, present := (*(*map[KeyType]ValType)(m.current))[key]
	return val, present
}

func (m *SharedMap) GetOrSet(key KeyType, valueMaker func() ValType) (ValType, bool) {
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

func (m *SharedMap) Copy() *map[KeyType]ValType {
	current := (*(*map[KeyType]ValType)(m.current))
	newMap := make(map[KeyType]ValType)
	for key, value := range current {
		newMap[key] = value
	}
	return &newMap
}

func (m *SharedMap) SetValue(key KeyType, value ValType) {
	m.writeLock.Lock()
	defer m.writeLock.Unlock()
	newCurrent := m.Copy()
	(*newCurrent)[key] = value
	atomic.StorePointer(&m.current, unsafe.Pointer(newCurrent))
}

func (m *SharedMap) Keys() []KeyType {
	var keys []KeyType
	for k := range (*map[KeyType]ValType)(m.current) {
		keys = append(keys, k)
	}
	return keys
}
