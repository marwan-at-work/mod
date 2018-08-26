package main

import (
	"log"

	"github.com/marwan-at-work/mod/major"
)

func main() {
	err := major.Run()
	if err != nil {
		log.Fatal(err)
	}
}
