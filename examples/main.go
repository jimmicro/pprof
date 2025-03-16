package main

import (
	"log"

	_ "github.com/jimmicro/pprof"
)

func main() {
	log.Println("Hello world")
	select {}
}
