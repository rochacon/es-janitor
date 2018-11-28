package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Elasticsearch client interface
type Elasticsearch interface {
	CatIndices() ([]Index, error)
	DeleteIndex(string) error
	RestoreSnapshot(string, string) error
	SnapshotIndex(string, string) error
}

// Client implements the Elasticsearch client interface
type Client struct {
	Endpoint string
}

// CatIndices retrieve all indexes information
func (c *Client) CatIndices() ([]Index, error) {
	resp, err := http.Get(c.Endpoint + "/_cat/indices?format=json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	indexes := []Index{}
	err = json.NewDecoder(resp.Body).Decode(&indexes)
	if err != nil {
		return nil, err
	}
	return indexes, nil
}

// DeleteIndex delete an index
func (c *Client) DeleteIndex(name string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", c.Endpoint, name), nil)
	if err != nil {
		return fmt.Errorf("cannot create request to delete index: %s", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request to delete index failed: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete index %s, response code: %d, response body: %s", name, resp.StatusCode, body)
	}
	return nil
}

// RestoreSnapshot requests the restore of snapshots.
// Multiple snapshots may be restored by providing a comma separated name.
func (c *Client) RestoreSnapshot(repository, name string) error {
	log.Println("Restoring snapshot", name)
	url := fmt.Sprintf("%s/_snapshot/%s/%s/_restore?wait_for_completion=true", c.Endpoint, repository, name)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("cannot create request to restore snapshot: %s", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request to restore snapshot failed: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to restore snapshot %s, response code: %d, response body: %s", name, resp.StatusCode, body)
	}
	return nil
}

// RestoreSnapshot requests the restore of snapshots.
// Multiple indexes may be snapshotted at once, by providing a comma separated name.
func (c *Client) SnapshotIndex(repository, indexName string) error {
	name := fmt.Sprintf("index-%s-on-%s", indexName, time.Now().Format(time.RFC3339))
	log.Println("Creating snapshot", name)
	url := fmt.Sprintf("%s/_snapshot/%s/%s?wait_for_completion=true", c.Endpoint, repository, name)
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, `{"indices": "%s", "include_global_state": false}`, indexName)
	req, err := http.NewRequest("PUT", url, buf)
	if err != nil {
		return fmt.Errorf("cannot create request to snapshot index: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request to snapshot index failed: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to snapshot %s, response code: %d, response body: %s", name, resp.StatusCode, body)
	}
	return nil
}
