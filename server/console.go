//go:build console
// +build console

package server

import (
	. "transparent/server/console"
)

func Start() error {
	return NewManager().Start()
}

func Stop() {
	NewManager().Stop()
}
