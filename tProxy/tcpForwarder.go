package tProxy

import (
	"transparent/gvisor.dev/gvisor/pkg/tcpip/stack"
)

func (m *manager) HandleTcpPacket(id stack.TransportEndpointID, pkt stack.PacketBufferPtr) bool {
	return m.tcpForwarder.HandlePacket(id, pkt)
}
