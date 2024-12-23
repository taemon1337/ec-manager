package main

import (
	"log"

	"github.com/taemon1337/ec-manager/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
