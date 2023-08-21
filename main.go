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

func init() {
	logging.Init()
	defer logging.Sync()
}

func main() {
	showVersion := flag.Bool("version", false, "Show the application version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s %s, commit %s, built at %s\n", processName, version, commit, date)
		os.Exit(0)
	}

	myec2.GetInstancesState()
}
