package main

import (
	"github.com/taylormonacelli/candleburn/logging"
	"github.com/taylormonacelli/candleburn/myec2"
)

func init() {
	logging.Init()
	defer logging.Sync()
}

func main() {
	myec2.GetInstancesState()
}
