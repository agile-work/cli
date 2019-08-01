package main

import (
	"flag"
	"fmt"
	"os"

	xmlParser "github.com/agile-work/cli/parser/xml"
)

func main() {
	fmt.Print("Horizon CLI - v1.0\n\n")

	jobCommand := flag.NewFlagSet("job", flag.ExitOnError)
	parse := jobCommand.String("parse", "", "XML file to parse.")
	translation := jobCommand.String("translation", "", "CSV file to make translation.")
	jsonTasks := jobCommand.String("json", "", "JSON file to save the xml parse.")

	if len(os.Args) < 2 {
		fmt.Println("job subcommand is required")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "job":
		jobCommand.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if jobCommand.Parsed() {
		if *parse == "" || (*translation == "" && *jsonTasks == "") {
			jobCommand.PrintDefaults()
			os.Exit(1)
		}
		if err := xmlParser.Process(*parse, *translation, *jsonTasks); err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}
