// automatically generated by stateify.

package cpuid

import (
	"transparent/gvisor.dev/gvisor/pkg/state"
)

func (h *hwCap) StateTypeName() string {
	return "pkg/cpuid.hwCap"
}

func (h *hwCap) StateFields() []string {
	return []string{
		"hwCap1",
		"hwCap2",
	}
}

func (h *hwCap) beforeSave() {}

// +checklocksignore
func (h *hwCap) StateSave(stateSinkObject state.Sink) {
	h.beforeSave()
	stateSinkObject.Save(0, &h.hwCap1)
	stateSinkObject.Save(1, &h.hwCap2)
}

func (h *hwCap) afterLoad() {}

// +checklocksignore
func (h *hwCap) StateLoad(stateSourceObject state.Source) {
	stateSourceObject.Load(0, &h.hwCap1)
	stateSourceObject.Load(1, &h.hwCap2)
}

func init() {
	state.Register((*hwCap)(nil))
}
