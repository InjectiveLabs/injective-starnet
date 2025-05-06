package main

import (
	"log"

	"github.com/InjectiveLabs/injective-starnet/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
