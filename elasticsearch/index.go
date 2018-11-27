package elasticsearch

import (
	"fmt"
	"regexp"
	"time"
)

// indexDateFromNameRE holds the regex to match
var indexDateFromNameRE = regexp.MustCompile("[0-9]+.[0-9]+.[0-9]+$")

// Index represents an Elasticsearch index
type Index struct {
	ID     string `json:"uuid"`
	Name   string `json:"index"`
	State  string `json:"state"`
	Health string `json:"health"`
	Size   string `json:"store.size"`
}

// DateFromName extracts the index date from its name
func (i Index) DateFromName() (time.Time, error) {
	date := indexDateFromNameRE.FindString(i.Name)
	if date == "" {
		return time.Time{}, fmt.Errorf("unable to find date from index name: %s", i.Name)
	}
	return time.Parse("2006.01.02", date)
}
