package main

import (
	"log"

	"github.com/cryskram/cogniq/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}