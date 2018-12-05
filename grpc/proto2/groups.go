package proto2

import (
	"sync/atomic"
	"unsafe"
)

// サーバが所属するグループを指定。
var groupsPtr unsafe.Pointer

func GetGroups() []string {
	p := (*[]string)(atomic.LoadPointer(&groupsPtr))
	if p == nil {
		return nil
	}
	return *p
}

// groupsを設定する。
func SetGroups(groups []string) bool {
	// Copy all items from groups to newSlice.
	newSlice := new([]string)
	for i := range groups {
		*newSlice = append(*newSlice, groups[i])
	}

	// Store newSlice atomically.
	before := atomic.LoadPointer(&groupsPtr)
	after := unsafe.Pointer(&newSlice)
	return atomic.CompareAndSwapPointer(&groupsPtr, before, after)
}
