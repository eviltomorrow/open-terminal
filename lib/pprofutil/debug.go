package pprofutil

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"strings"
)

func Run(addr string) error {
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/debug/pprof/", pprof.Index)
	httpMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	httpMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	httpMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	httpMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	format := "%s"
	if strings.HasPrefix(addr, ":") {
		format = fmt.Sprintf("127.0.0.1%s", addr)
	}
	log.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	log.Printf("+ Sample pprof profiling will be open at: http://%s/debug/pprof        +\r\n", format)
	log.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	fmt.Println()
	return http.ListenAndServe(addr, httpMux)
}
