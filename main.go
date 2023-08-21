package main

import (
	"flag"

	"github.com/taylormonacelli/candleburn/logging"
	"github.com/taylormonacelli/candleburn/myec2"
)

func init() {
	logging.Init()
	defer logging.Sync()
}

func main() {
	flag.Parse()
	myec2.GetInstancesState()
}
