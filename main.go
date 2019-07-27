package main

import (
	"flag"
	"fmt"

	xmlParser "github.com/agile-work/cli/parser/xml"
)

var (
	create          = flag.String("create", "", "Create file")
	from            = flag.String("from", "", "From path file")
	to              = flag.String("to", "", "To path file")
	withTranslation = flag.String("withTranslation", "", "Use translation to create another file")
	trackJob        = flag.String("trackJob", "", "Job ID to track task execution")
)

func main() {
	fmt.Print("Horizon CLI - v1.0\n\n")
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}

	switch *create {
	case "translation":
		if err := xmlParser.CreateTranslation(*from, *to); err != nil {
			fmt.Println(err.Error())
			return
		}
	case "json-tasks":
		if err := xmlParser.GenerateJobTasks(*from, *to, *withTranslation); err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
