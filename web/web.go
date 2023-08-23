package web

import (
	"flag"
	"fmt"
	"net/http"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from default endpoint!")
}

func Run(f func(http.ResponseWriter, *http.Request)) {
	port := flag.Int("port", 4444, "Port to listen on")
	flag.Parse()

	http.HandleFunc("/run", f)
	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Listening on %s\n", addr)
	http.ListenAndServe(addr, nil)
}
