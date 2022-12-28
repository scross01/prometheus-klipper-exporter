package main

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/scross01/prometheus-klipper-exporter/collector"
)

// Command line configuration options
var (
	listenAddress = flag.String("web.listen-address", ":9101", "Address on which to expose metrics and web interface.")
	klipperApiKey = flag.String("moonraker.apikey", "", "API Key to authenticate with the Klipper APIs.")
	debug         = flag.Bool("debug", false, "Enable debug logging.")
	verbose       = flag.Bool("verbose", false, "Enable verbose trace level logging")
)

func handler(w http.ResponseWriter, r *http.Request, logger log.Logger) {
	query := r.URL.Query()

	target := query.Get("target")
	if len(query["target"]) != 1 || target == "" {
		http.Error(w, "'target' parameter must be specified once", 400)
		return
	}

	// Set default modules
	modules := []string{"process_stats", "job_queue", "system_info"}
	// get `modules` configuration passed from the prometheus.yml
	if len(query["modules"]) > 0 {
		modules = query["modules"]
	}
	logger.Infof("Starting metrics collection of %s for %s", modules, target)

	// set api key. prometheus.yml > command line arg > environment variable
	apiKey := ""
	auth := r.Header.Get("Authorization")
	if auth != "" && strings.HasPrefix(auth, "APIKEY") {
		apiKey = strings.Replace(auth, "APIKEY ", "", 1)
		logger.Trace("Using API key from prometheus.yml authorization configuration")
	} else if *klipperApiKey != "" {
		apiKey = *klipperApiKey
		logger.Trace("Using API key from -moonraker.apikey command line argument")
	} else if apiKey = os.Getenv("MOONRAKER_APIKEY"); apiKey != "" {
		logger.Trace("Using API key from MOONRAKER_APIKEY environment variable")
	} else {
		logger.Trace("API key not set")
	}

	registry := prometheus.NewRegistry()
	c := collector.New(r.Context(), target, modules, apiKey, logger)
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
	if *verbose {
		logger.SetLevel(log.TraceLevel)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, *logger)
	})
	logger.Infof("Beginning to serve on port %s", *listenAddress)
	logger.Fatal(http.ListenAndServe(*listenAddress, nil))
}
