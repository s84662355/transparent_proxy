package server

func Start() error {
	return newManager().Start()
}

func Stop() {
	newManager().Stop()
}
