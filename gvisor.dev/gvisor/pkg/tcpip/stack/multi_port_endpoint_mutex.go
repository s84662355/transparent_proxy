package stack

import (
	"reflect"

	"transparent/gvisor.dev/gvisor/pkg/sync"
	"transparent/gvisor.dev/gvisor/pkg/sync/locking"
)

// RWMutex is sync.RWMutex with the correctness validator.
type multiPortEndpointRWMutex struct {
	mu sync.RWMutex
}

// lockNames is a list of user-friendly lock names.
// Populated in init.
var multiPortEndpointlockNames []string

// lockNameIndex is used as an index passed to NestedLock and NestedUnlock,
// refering to an index within lockNames.
// Values are specified using the "consts" field of go_template_instance.
type multiPortEndpointlockNameIndex int

// DO NOT REMOVE: The following function automatically replaced with lock index constants.
// LOCK_NAME_INDEX_CONSTANTS
const ()

// Lock locks m.
// +checklocksignore
func (m *multiPortEndpointRWMutex) Lock() {
	locking.AddGLock(multiPortEndpointprefixIndex, -1)
	m.mu.Lock()
}

// NestedLock locks m knowing that another lock of the same type is held.
// +checklocksignore
func (m *multiPortEndpointRWMutex) NestedLock(i multiPortEndpointlockNameIndex) {
	locking.AddGLock(multiPortEndpointprefixIndex, int(i))
	m.mu.Lock()
}

// Unlock unlocks m.
// +checklocksignore
func (m *multiPortEndpointRWMutex) Unlock() {
	m.mu.Unlock()
	locking.DelGLock(multiPortEndpointprefixIndex, -1)
}

// NestedUnlock unlocks m knowing that another lock of the same type is held.
// +checklocksignore
func (m *multiPortEndpointRWMutex) NestedUnlock(i multiPortEndpointlockNameIndex) {
	m.mu.Unlock()
	locking.DelGLock(multiPortEndpointprefixIndex, int(i))
}

// RLock locks m for reading.
// +checklocksignore
func (m *multiPortEndpointRWMutex) RLock() {
	locking.AddGLock(multiPortEndpointprefixIndex, -1)
	m.mu.RLock()
}

// RUnlock undoes a single RLock call.
// +checklocksignore
func (m *multiPortEndpointRWMutex) RUnlock() {
	m.mu.RUnlock()
	locking.DelGLock(multiPortEndpointprefixIndex, -1)
}

// RLockBypass locks m for reading without executing the validator.
// +checklocksignore
func (m *multiPortEndpointRWMutex) RLockBypass() {
	m.mu.RLock()
}

// RUnlockBypass undoes a single RLockBypass call.
// +checklocksignore
func (m *multiPortEndpointRWMutex) RUnlockBypass() {
	m.mu.RUnlock()
}

// DowngradeLock atomically unlocks rw for writing and locks it for reading.
// +checklocksignore
func (m *multiPortEndpointRWMutex) DowngradeLock() {
	m.mu.DowngradeLock()
}

var multiPortEndpointprefixIndex *locking.MutexClass

// DO NOT REMOVE: The following function is automatically replaced.
func multiPortEndpointinitLockNames() {}

func init() {
	multiPortEndpointinitLockNames()
	multiPortEndpointprefixIndex = locking.NewMutexClass(reflect.TypeOf(multiPortEndpointRWMutex{}), multiPortEndpointlockNames)
}
