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
	loggingLevel  = flag.String("logging.level", "Info", "Logging output level. Set to one of Trace, Debug, Info, Warning, Error, Fatal, or Panic")
	klipperApiKey = flag.String("moonraker.apikey", "", "API Key to authenticate with the Klipper APIs.")
	listenAddress = flag.String("web.listen-address", ":9101", "Address on which to expose metrics and web interface.")
	// TODO deprecated, to be removed.
	debug   = flag.Bool("debug", false, "(Deprecated) Enable debug logging. Use -logging.level instead.")
	verbose = flag.Bool("verbose", false, "(Deprecated) Enable verbose trace level logging. Use -logging.level instead.")
)

func handler(w http.ResponseWriter, r *http.Request) {
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
	log.Infof("Starting metrics collection of %s for %s", modules, target)

	// set api key. prometheus.yml > command line arg > environment variable
	apiKey := ""
	auth := r.Header.Get("Authorization")
	if auth != "" && strings.HasPrefix(auth, "APIKEY") {
		apiKey = strings.Replace(auth, "APIKEY ", "", 1)
		log.Debug("Using API key from prometheus.yml authorization configuration")
	} else if *klipperApiKey != "" {
		apiKey = *klipperApiKey
		log.Debug("Using API key from -moonraker.apikey command line argument")
	} else if apiKey = os.Getenv("MOONRAKER_APIKEY"); apiKey != "" {
		log.Debug("Using API key from MOONRAKER_APIKEY environment variable")
	} else {
		log.Debug("API key not set")
	}

	registry := prometheus.NewRegistry()
	c := collector.New(r.Context(), target, modules, apiKey)
	registry.MustRegister(c)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	flag.Parse()

	level, err := log.ParseLevel(strings.ToLower(*loggingLevel))
	if err != nil {
		log.Fatalf("Invalid logging level '%s'", *loggingLevel)
	}
	log.SetLevel(level)

	// TODO remove when -debug and -verbose options are removed
	if *debug {
		log.Warn("-debug option is deprecated, change to using '-logging.level debug'")
		log.SetLevel(log.DebugLevel)
	}
	if *verbose {
		log.Warn("-verbose option is deprecated, change to using '-logging.level trace'")
		log.SetLevel(log.TraceLevel)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	})
	log.Infof("Beginning to serve on port %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
