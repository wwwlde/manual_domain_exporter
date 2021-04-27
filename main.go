package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Domains struct which contains an array of domains
type Domains struct {
	Domains []Domain `json:"domains"`
}

// Domain struct which contains a name & expire
type Domain struct {
	Name   string `json:"name"`
	Expire string `json:"expire"`
}

var (
	// How often to check domains
	checkRate = 12 * time.Hour
	formats   = []string{
		"2006-01-02",
	}
	configFile = kingpin.Flag("config", "Domain exporter configuration file.").Default("domains.json").String()
	httpBind   = kingpin.Flag("bind", "The address to listen on for HTTP requests.").Default(":9203").String()

	domainExpiration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "domain_expiration",
			Help: "Days until the WHOIS record states this domain will expire",
		},
		[]string{
			"domain",
			"manual",
		},
	)

	// we initialize our Domains array
	domains Domains

	version = "0.0.2"
)

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)
	_ = level.Info(logger).Log("msg", "Starting manual_domain_exporter", "version", version)

	prometheus.Register(domainExpiration)

	filename, err := filepath.Abs(*configFile)
	if err != nil {
		_ = level.Warn(logger).Log("warn", err)
	}

	// Open our jsonFile
	jsonFile, err := ioutil.ReadFile(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		_ = level.Warn(logger).Log("warn", err)
	}

	err = json.Unmarshal(jsonFile, &domains)

	if err != nil {
		_ = level.Warn(logger).Log("warn", err)
	} else {
		go func() {
			for {
				for i := 0; i < len(domains.Domains); i++ {
					err = lookup(domains.Domains[i], domainExpiration, logger)
					if err != nil {
						_ = level.Warn(logger).Log("warn", err)
					}
					continue
				}
				time.Sleep(checkRate)
			}
		}()
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(
			w, `
			<html>
			<head><title>Manual Domain Exporter</title></head>
			<body>
				<h1>Manual Domain Exporter</h1>
				<h2>Denys Lemeshko</h2>
				<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>
			`,
		)
	})
	_ = level.Info(logger).Log("msg", "Listening", "port", *httpBind)
	if err := http.ListenAndServe(*httpBind, nil); err != nil {
		_ = level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}

func lookup(domain Domain, handler *prometheus.GaugeVec, logger log.Logger) error {

	for _, format := range formats {
		if date, err := time.Parse(format, strings.TrimSpace(domain.Expire)); err == nil {
			days := math.Floor(time.Until(date).Hours() / 24)
			_ = level.Info(logger).Log("domain:", domain.Name, "days", days, "date", date)
			handler.WithLabelValues(domain.Name, "1").Set(days)
			return nil
		}
	}

	return fmt.Errorf("Unable to parse date: %s, for %s", strings.TrimSpace(domain.Expire), domain.Name)
}
