package main

import (
	"flag"
	"fmt"

	"github.com/agile-work/cli/parser"
)

var (
	xmlPath = flag.String("xmlPath", "", "Path to the module xml job tasks")
)

func main() {
	fmt.Println("Horizon CLI - v1.0")
	if err := parser.GenerateJobTasks("module_task_1.0.xml"); err != nil {
		fmt.Println(err.Error())
		return
	}
}
