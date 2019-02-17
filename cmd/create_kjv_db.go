package main

import (
	"flag"
	"fmt"

	"github.com/r4wm/kjvapi"
)

// main: Create the kjv database at desired path for mintz5/deploy.go
func main() {
	var dbPath = flag.String("dbpath", "/tmp/kjv.db", "Path where DB should be created.")
	flag.Parse()

	fmt.Printf("Creating kjv db to %s\n", *dbPath)
	kjvapi.CreateKJVDB(*dbPath)

}
