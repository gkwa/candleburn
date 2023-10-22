package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/taylormonacelli/candleburn/myec2"
	"github.com/taylormonacelli/candleburn/web"
	"github.com/taylormonacelli/goldbug"
)

var (
	version     = "dev"
	commit      = "none"
	date        = "unknown"
	processName = os.Args[0]
	verbose     bool
	logFormat   string
)

var (
	showVersion bool
	listen      bool
	outfile     string
)

func dostuff(w http.ResponseWriter, r *http.Request) {
	absPath, _ := filepath.Abs("hosts.yaml")

	instances, err := myec2.LoadInstancesFromYAML(absPath)
	if err != nil {
		// msg := fmt.Errorf("failed to load instances from yaml %s: %w", absPath, err)
		slog.Error("loading instances from yaml failed", "error", err)
		os.Exit(1)
	}

	results, err := myec2.GetInstancesState(instances)
	if err != nil {
		slog.Error("loading instances from yaml failed", "error", err)
		os.Exit(1)
	}
	myec2.ExportInstancesQuery(results, outfile)
}

func main() {
	flag.BoolVar(&showVersion, "version", false, "Show the application version")
	flag.BoolVar(&listen, "listen", false, "Listen for incommming commands")
	flag.StringVar(&outfile, "outfile", fmt.Sprintf("%s.json", processName), "Save query to this file or - for stdout")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output (shorthand)")

	flag.StringVar(&logFormat, "log-format", "", "Log format (text or json)")

	flag.Parse()

	if showVersion || len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("%s %s, commit %s, built at %s\n", processName, version, commit, date)
		os.Exit(0)
	}

	if verbose || logFormat != "" {
		if logFormat == "json" {
			goldbug.SetDefaultLoggerJson(slog.LevelDebug)
		} else {
			goldbug.SetDefaultLoggerText(slog.LevelDebug)
		}
	}

	if listen {
		web.Run(dostuff)
	} else {
		dostuff(nil, nil)
	}
}
