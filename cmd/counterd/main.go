package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	readTimeout  = time.Second * 5
	writeTimeout = time.Second * 5

	fwdForHeader = "X-Forwarded-For"

	namespace = "counterd"
)

var (
	addr = flag.String("addr", "0.0.0.0:8000", "HTTP addr")

	count uint64
	host  string

	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "http_requests_total",
		Namespace: namespace,
		Help:      "Total number of HTTP requests",
	}, []string{"endpoint"})
)

func main() {
	flag.Parse()

	var err error
	host, err = os.Hostname()
	if err != nil {
		log.Fatalf("unable to get hostname: %v", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", Count)
	r.HandleFunc("/dump", Dump)
	r.HandleFunc("/ping", Ping)
	r.Handle("/metrics", promhttp.Handler())
	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	srv := &http.Server{
		Handler:      r,
		Addr:         *addr,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	log.Printf("Starting service at %s", srv.Addr)
	err = gracehttp.Serve(srv)
	if err != nil {
		log.Fatalf("Failure in gracehttp.Serve(), err: %s", err)
	}
}

func Count(w http.ResponseWriter, r *http.Request) {
	httpRequestsTotal.WithLabelValues("/").Inc()
	atomic.AddUint64(&count, 1)
	lines := []string{
		fmt.Sprintf("host: %v", host),
		fmt.Sprintf("count: %v", atomic.LoadUint64(&count)),
		fmt.Sprintf("remote: %v", r.RemoteAddr),
	}

	fwdFor := r.Header.Get(fwdForHeader)
	if fwdFor != "" {
		f := fmt.Sprintf("X-Forwarded-For: %v", fwdFor)
		lines = append(lines, f)
	}

	msg := strings.Join(lines, "\n")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msg + "\n"))
}

func Dump(w http.ResponseWriter, r *http.Request) {
	httpRequestsTotal.WithLabelValues("/dump").Inc()
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		http.Error(w, "failed to dump request", http.StatusInternalServerError)
		return
	}

	fmt.Print(string(dump))
	w.WriteHeader(http.StatusOK)
}

func Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}
