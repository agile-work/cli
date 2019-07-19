package main

import (
	"flag"
	"fmt"
)

var (
	xmlPath = flag.String("xmlPath", "", "Path to the module xml job tasks")
)

func main() {
	fmt.Println("Horizon CLI - v1.0")
}
