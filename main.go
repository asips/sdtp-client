package main

import (
	"github.com/asips/sdtp-client/cmd"
	"github.com/asips/sdtp-client/internal/log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal("%s", err)
	}
}
