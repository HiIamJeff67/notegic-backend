package main

import (
	"os"

	realtimegateway "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway"
)

func main() {
	application := realtimegateway.NewApplication()
	if len(os.Args) == 1 {
		shutdown := application.Start()
		defer shutdown()
	}

	Execute()
}
