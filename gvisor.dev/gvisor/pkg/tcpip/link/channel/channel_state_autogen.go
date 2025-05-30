// automatically generated by stateify.

package channel

import (
	"transparent/gvisor.dev/gvisor/pkg/state"
)

func (n *NotificationHandle) StateTypeName() string {
	return "pkg/tcpip/link/channel.NotificationHandle"
}

func (n *NotificationHandle) StateFields() []string {
	return []string{
		"n",
	}
}

func (n *NotificationHandle) beforeSave() {}

// +checklocksignore
func (n *NotificationHandle) StateSave(stateSinkObject state.Sink) {
	n.beforeSave()
	stateSinkObject.Save(0, &n.n)
}

func (n *NotificationHandle) afterLoad() {}

// +checklocksignore
func (n *NotificationHandle) StateLoad(stateSourceObject state.Source) {
	stateSourceObject.Load(0, &n.n)
}

func init() {
	state.Register((*NotificationHandle)(nil))
}
