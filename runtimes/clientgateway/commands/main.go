package main

import (
	"os"

	clientgateway "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway"
)

func main() {
	application := clientgateway.NewApplication()
	if len(os.Args) == 1 {
		shutdown := application.Start()
		defer shutdown()
	}

	Execute()
}
