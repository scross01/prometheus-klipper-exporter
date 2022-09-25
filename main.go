package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/scross01/prometheus-klipper-exporter/collector"
)

// Command line configuration options
var (
	listenAddress = flag.String("web.listen-address", ":9101", "Address on which to expose metrics and web interface.")
	debug         = flag.Bool("debug", false, "Enable debug logging")
)

func handler(w http.ResponseWriter, r *http.Request, logger log.Logger) {
	query := r.URL.Query()

	target := query.Get("target")
	if len(query["target"]) != 1 || target == "" {
		http.Error(w, "'target' parameter must be specified once", 400)
		return
	}

	log.Infof("Starting metrics collection for %s", target)

	registry := prometheus.NewRegistry()
	c := collector.New(r.Context(), target, logger)
	registry.MustRegister(c)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	var logger = log.New()
	flag.Parse()
	if *debug {
		logger.SetLevel(log.DebugLevel)
	} else {
		logger.SetLevel(log.InfoLevel)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, *logger)
	})
	logger.Infof("Beginning to serve on port %s", *listenAddress)
	logger.Fatal(http.ListenAndServe(*listenAddress, nil))
}
