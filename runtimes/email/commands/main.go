package main

import (
	"os"

	email "github.com/HiIamJeff67/notegic-backend/runtimes/email"
)

func main() {
	application := email.NewApplication()
	if len(os.Args) == 1 {
		shutdown := application.Start()
		defer shutdown()
	}

	Execute()
}
