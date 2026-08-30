package main

import (
	"os"

	core "github.com/HiIamJeff67/notegic-backend/runtimes/core"
)

func main() {
	application := core.NewApplication()
	if len(os.Args) == 1 {
		shutdown := application.Start()
		defer shutdown()
	}

	Execute()
}
