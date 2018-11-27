package janitor

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/rochacon/es-janitor/elasticsearch"
)

func TestArchiveOlderThan(t *testing.T) {
	es := NewFakeES()
	if len(es.Indexes) != 7 {
		t.Errorf("test requires 7 indexes to exist")
	}
	if len(es.Deleted) != 0 {
		t.Errorf("test requires empty list of deleted")
	}
	if len(es.Snapshots) != 0 {
		t.Errorf("test requires empty list of snapshots")
	}

	j := &Janitor{
		Elasticsearch: es,
		Repository:    "ffs",
	}
	err := j.ArchiveOlderThan(3)
	if err != nil {
		t.Errorf("failed to archive indexes older than 3 days: %s", err)
	}
	if len(es.Indexes) != 3 {
		t.Errorf("incorrect remaining indexes count: %#v", es)
	}
	if len(es.Snapshots) != len(es.Deleted) {
		t.Errorf("mismatching snapshot and deleted count: %#v", es)
	}
	if len(es.Snapshots) != 4 {
		t.Errorf("incorrect snapshot count: %#v", es)
	}
	if len(es.Deleted) != 4 {
		t.Errorf("incorrect deleted count: %#v", es)
	}
}
func TestArchiveOlderThanWithRepositoryDashSkipsSnapshot(t *testing.T) {
	es := NewFakeES()
	if len(es.Indexes) != 7 {
		t.Errorf("test requires 7 indexes to exist")
	}
	if len(es.Deleted) != 0 {
		t.Errorf("test requires empty list of deleted")
	}
	if len(es.Snapshots) != 0 {
		t.Errorf("test requires empty list of snapshots")
	}

	j := &Janitor{
		Elasticsearch: es,
		Repository:    "-",
	}
	err := j.ArchiveOlderThan(3)
	if err != nil {
		t.Errorf("failed to archive indexes older than 3 days: %s", err)
	}
	if len(es.Indexes) != 3 {
		t.Errorf("incorrect remaining indexes count: %#v", es)
	}
	if len(es.Snapshots) != 0 {
		t.Errorf("incorrect snapshot count: %#v", es)
	}
	if len(es.Deleted) != 4 {
		t.Errorf("incorrect deleted count: %#v", es)
	}
}

type FakeES struct {
	Deleted   []elasticsearch.Index
	Indexes   []elasticsearch.Index
	Snapshots []string
}

func NewFakeES() *FakeES {
	return &FakeES{
		Deleted: []elasticsearch.Index{},
		Indexes: []elasticsearch.Index{
			{Name: fmt.Sprintf("filebeat-6.4.3-%s", time.Now().Add(-5*24*time.Hour).Format("2006.01.02"))},
			{Name: fmt.Sprintf("filebeat-6.4.3-%s", time.Now().Add(-4*24*time.Hour).Format("2006.01.02"))},
			{Name: fmt.Sprintf("filebeat-6.4.3-%s", time.Now().Add(-3*24*time.Hour).Format("2006.01.02"))},
			{Name: fmt.Sprintf("filebeat-6.5.0-%s", time.Now().Add(-3*24*time.Hour).Format("2006.01.02"))},
			{Name: fmt.Sprintf("filebeat-6.5.0-%s", time.Now().Add(-2*24*time.Hour).Format("2006.01.02"))},
			{Name: fmt.Sprintf("filebeat-6.5.0-%s", time.Now().Add(-1*24*time.Hour).Format("2006.01.02"))},
			{Name: fmt.Sprintf("filebeat-6.5.0-%s", time.Now().Add(-0*24*time.Hour).Format("2006.01.02"))},
		},
		Snapshots: []string{},
	}
}

func (f *FakeES) CatIndices() ([]elasticsearch.Index, error) {
	log.Printf("FakeES.CatIndices()")
	indexes := []elasticsearch.Index{}
	for _, index := range f.Indexes {
		indexes = append(indexes, index)
	}
	return indexes, nil
}

func (f *FakeES) DeleteIndex(name string) error {
	log.Printf("FakeES.DeleteIndex(%s)", name)
	for i, index := range f.Indexes {
		if index.Name == name {
			f.Deleted = append(f.Deleted, index)
			f.Indexes = append(f.Indexes[:i], f.Indexes[i+1:]...)
			break
		}
	}
	return nil
}

func (f *FakeES) SnapshotIndex(repo, name string) error {
	log.Printf("FakeES.SnapshotIndex(%s, %s)", repo, name)
	f.Snapshots = append(f.Snapshots, fmt.Sprintf("%s-%s", repo, name))
	return nil
}

func (f *FakeES) RestoreSnapshot(repo, name string) error {
	log.Printf("FakeES.RestoreSnapshot(%s, %s)", repo, name)
	return nil
}
