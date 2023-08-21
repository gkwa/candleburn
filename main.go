package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/taylormonacelli/candleburn/logging"
	"github.com/taylormonacelli/candleburn/myec2"
)

var (
	version     = "dev"
	commit      = "none"
	date        = "unknown"
	processName = os.Args[0]
)

var (
	showVersion bool
	outfile     string
)

func init() {
	logging.Init()
	defer logging.Sync()
}

func main() {
	flag.BoolVar(&showVersion, "version", false, "Show the application version")
	flag.StringVar(&outfile, "outfile", fmt.Sprintf("%s.json", processName), "Save query to this file")

	flag.Parse()

	if showVersion || len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("%s %s, commit %s, built at %s\n", processName, version, commit, date)
		return
	}

	results, err := myec2.GetInstancesState()
	if err != nil {
		panic(err)
	}
	myec2.ExportInstancesQuery(results)
}
