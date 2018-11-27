package janitor

import (
	"fmt"
	"log"
	"time"

	"github.com/rochacon/es-janitor/elasticsearch"
)

// Janitor holds logic on how to keep things clean
type Janitor struct {
	Elasticsearch elasticsearch.Elasticsearch
	Repository    string
}

// ArchiveOlderThan snapshots and delete Elasticsearch indexes older
// than a given number of days.
func (j *Janitor) ArchiveOlderThan(days int64) error {
	indexes, err := j.getArchiveElibigleIndexes(days)
	if err != nil {
		return err
	}
	for _, index := range *indexes {
		log.Println("Archiving index", index.Name)
		if j.Repository != "-" {
			err = j.Elasticsearch.SnapshotIndex(j.Repository, index.Name)
			if err != nil {
				return fmt.Errorf("failed to snapshot index: %s", err)
			}
		}
		err = j.Elasticsearch.DeleteIndex(index.Name)
		if err != nil {
			return fmt.Errorf("failed to delete index: %s", err)
		}
	}
	return nil
}

// getArchiveElibigleIndexes filter Elasticsearch filter that may be
// archived given its name date.
func (j *Janitor) getArchiveElibigleIndexes(days int64) (*[]elasticsearch.Index, error) {
	indexes, err := j.Elasticsearch.CatIndices()
	if err != nil {
		return nil, err
	}
	diff := time.Duration(time.Hour*24) * time.Duration(days)
	filtered := []elasticsearch.Index{}
	for _, index := range indexes {
		created, err := index.DateFromName()
		if err != nil {
			continue
		}
		if time.Since(created) > diff {
			filtered = append(filtered, index)
		}
	}
	return &filtered, nil
}
