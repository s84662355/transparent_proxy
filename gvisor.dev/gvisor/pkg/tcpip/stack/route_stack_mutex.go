package stack

import (
	"reflect"

	"transparent/gvisor.dev/gvisor/pkg/sync"
	"transparent/gvisor.dev/gvisor/pkg/sync/locking"
)

// RWMutex is sync.RWMutex with the correctness validator.
type routeStackRWMutex struct {
	mu sync.RWMutex
}

// lockNames is a list of user-friendly lock names.
// Populated in init.
var routeStacklockNames []string

// lockNameIndex is used as an index passed to NestedLock and NestedUnlock,
// refering to an index within lockNames.
// Values are specified using the "consts" field of go_template_instance.
type routeStacklockNameIndex int

// DO NOT REMOVE: The following function automatically replaced with lock index constants.
// LOCK_NAME_INDEX_CONSTANTS
const ()

// Lock locks m.
// +checklocksignore
func (m *routeStackRWMutex) Lock() {
	locking.AddGLock(routeStackprefixIndex, -1)
	m.mu.Lock()
}

// NestedLock locks m knowing that another lock of the same type is held.
// +checklocksignore
func (m *routeStackRWMutex) NestedLock(i routeStacklockNameIndex) {
	locking.AddGLock(routeStackprefixIndex, int(i))
	m.mu.Lock()
}

// Unlock unlocks m.
// +checklocksignore
func (m *routeStackRWMutex) Unlock() {
	m.mu.Unlock()
	locking.DelGLock(routeStackprefixIndex, -1)
}

// NestedUnlock unlocks m knowing that another lock of the same type is held.
// +checklocksignore
func (m *routeStackRWMutex) NestedUnlock(i routeStacklockNameIndex) {
	m.mu.Unlock()
	locking.DelGLock(routeStackprefixIndex, int(i))
}

// RLock locks m for reading.
// +checklocksignore
func (m *routeStackRWMutex) RLock() {
	locking.AddGLock(routeStackprefixIndex, -1)
	m.mu.RLock()
}

// RUnlock undoes a single RLock call.
// +checklocksignore
func (m *routeStackRWMutex) RUnlock() {
	m.mu.RUnlock()
	locking.DelGLock(routeStackprefixIndex, -1)
}

// RLockBypass locks m for reading without executing the validator.
// +checklocksignore
func (m *routeStackRWMutex) RLockBypass() {
	m.mu.RLock()
}

// RUnlockBypass undoes a single RLockBypass call.
// +checklocksignore
func (m *routeStackRWMutex) RUnlockBypass() {
	m.mu.RUnlock()
}

// DowngradeLock atomically unlocks rw for writing and locks it for reading.
// +checklocksignore
func (m *routeStackRWMutex) DowngradeLock() {
	m.mu.DowngradeLock()
}

var routeStackprefixIndex *locking.MutexClass

// DO NOT REMOVE: The following function is automatically replaced.
func routeStackinitLockNames() {}

func init() {
	routeStackinitLockNames()
	routeStackprefixIndex = locking.NewMutexClass(reflect.TypeOf(routeStackRWMutex{}), routeStacklockNames)
}
