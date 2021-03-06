package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/rochacon/es-janitor/elasticsearch"
	"github.com/rochacon/es-janitor/janitor"
)

var (
	VERSION = "dev"
)

func main() {
	days := flag.Int64("days", 32, "Number of days of indexes to keep")
	endpoint := flag.String("endpoint", "", "Elasticsearch base endpoint")
	repository := flag.String("repository", "", "Elasticsearch snapshot repository. Use - to skip snapshots.")
	version := flag.Bool("version", false, "Print version information.")
	flag.Parse()

	if *version {
		fmt.Println("es-janitor", VERSION)
		return
	}

	if *endpoint == "" {
		log.Fatalf("Elasticsearch Endpoint must be provided")
	}

	if *repository == "" {
		log.Fatalf("Elasticsearch snapshot repository must be provided")
	}

	log.Printf("Elasticsearch endpoint: %s", *endpoint)
	log.Printf("Archiving indexes older than %d days", *days)
	j := &janitor.Janitor{
		Elasticsearch: &elasticsearch.Client{
			Endpoint: strings.TrimSuffix(*endpoint, "/"),
		},
		Repository: *repository,
	}
	err := j.ArchiveOlderThan(*days)
	if err != nil {
		log.Fatalf("Failed to archive old indexes: %s", err)
	}
}
