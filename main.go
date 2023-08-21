package main

import (
	"flag"

	"github.com/taylormonacelli/candleburn/logging"
	"github.com/taylormonacelli/candleburn/myec2"
)

var (
	tagName    string
	region     string
	outputFile string
)

func init() {
	flag.StringVar(&tagName, "tag", "Name", "Specify the tag key for identifying instances")
	flag.StringVar(&region, "region", "us-east-1", "Specify the AWS region to query")
	flag.StringVar(&outputFile, "output", "goldpuppy.json", "Specify the output JSON file name")
	logging.Init()
	defer logging.Sync()
}

func main() {
	flag.Parse()
	myec2.GetInstancesState()
}
