package main

import (
	"flag"
	"fmt"

	xmlParser "github.com/agile-work/cli/parser/xml"
)

var (
	xmlPath             = flag.String("xmlPath", "", "Path to xml file")
	translationPath     = flag.String("translationPath", "", "Path to translation file")
	generateTranslation = flag.Bool("generateTranslation", false, "Generate a translation file")
	trackJob            = flag.String("trackJob", "", "Job ID to track task execution")
)

func main() {
	fmt.Print("Horizon CLI - v1.0\n\n")
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}

	if isFlagPassed("xmlPath") {
		if err := xmlParser.GenerateJobTasks(*xmlPath, *translationPath, *generateTranslation); err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	if isFlagPassed("trackJob") {

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
