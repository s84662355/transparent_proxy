//go:build gui
// +build gui

package server

import (
	. "transparent/server/gui"
)

func Start() error {
	return NewManager().Start()
}

func Stop() {
	NewManager().Stop()
}
