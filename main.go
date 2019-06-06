package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenAddress  = flag.String("listen", ":9108", "Listen address for prometheus")
	metricsPath    = flag.String("path", "/metrics", "Path under which to expose metrics")
	updateInterval = flag.Int64("interval", 600, "Queue update interval, seconds")
	tagsAsLables   = flag.Bool("tags", true, "Add tags as labels to metrics")
	queuePrefix    = flag.String("prefix", "", "Queue prefix to fetch, will be used before filter")
	queueFilter    = flag.String("filter", ".*", "Regex to filter queue list after fetching")
)

func main() {
	flag.Parse()

	if len(os.Getenv("PREFIX")) > 0 {
		*queuePrefix = os.Getenv("PREFIX")
	}

	if len(os.Getenv("FILTER")) > 0 {
		*queueFilter = os.Getenv("FILTER")
	}

	if len(os.Getenv("INTERVAL")) > 0 {
		if i, err := strconv.ParseInt(os.Getenv("INTERVAL"), 10, 64); err == nil {
			*updateInterval = i
		}
	}

	if strings.ToLower(os.Getenv("TAGS")) == "false" {
		*tagsAsLables = false
	}

	ctx, cancel := context.WithCancel(context.Background())

	regex := regexp.MustCompile(*queueFilter)
	col := newCollector(ctx, time.Second*time.Duration(*updateInterval), queuePrefix, regex, *tagsAsLables)

	r := prometheus.NewRegistry()
	r.MustRegister(col)

	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})

	http.Handle(*metricsPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(`<html>
<head><title>AWS SQS exporter</title></head>
<body>
<h1>AWS SQS exporter</h1>
<p><a href="%s">Metrics</a></p>
</body>
</html>`, *metricsPath)))
	})

	log.Printf("Starting http server, listening on %s\n", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		cancel()
		log.Fatal(err)
	}
}
