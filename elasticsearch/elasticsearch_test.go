package elasticsearch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientCatIndices(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Invalid HTTP Method: %s", r.Method)
		}
		if r.URL.Path != "/_cat/indices" {
			t.Errorf("incorrect path used: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("format"); got != "json" {
			t.Errorf("incorrect format querystring used: %s", got)
		}
		fmt.Fprintln(w, `[
			{"index": "filebeat-6.4.3-2018.11.21"},
			{"index": "filebeat-6.4.3-2018.11.22"},
			{"index": "filebeat-6.4.3-2018.11.23"}
		]`)
	}))
	defer ts.Close()

	es := &Client{
		Endpoint: ts.URL,
	}
	indexes, err := es.CatIndices()
	if err != nil {
		t.Errorf("failed to retrieve index list: %s", err)
	}
	if len(indexes) != 3 {
		t.Errorf("index list does not look as expected: %#v", indexes)
	}
	if indexes[0].Name != "filebeat-6.4.3-2018.11.21" {
		t.Errorf("index 0 does not look as expected: filebeat-6.4.3-2018.11.21, got: %#v", indexes[0])
	}
	if indexes[1].Name != "filebeat-6.4.3-2018.11.22" {
		t.Errorf("index 1 does not look as expected: filebeat-6.4.3-2018.11.22, got: %#v", indexes[1])
	}
	if indexes[2].Name != "filebeat-6.4.3-2018.11.23" {
		t.Errorf("index 2 does not look as expected: filebeat-6.4.3-2018.11.23, got: %#v", indexes[2])
	}
}

func TestClientCatIndicesInvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `not a json`)
	}))
	defer ts.Close()

	es := &Client{
		Endpoint: ts.URL,
	}
	indexes, err := es.CatIndices()
	if err == nil {
		t.Errorf("no error was returned")
	}
	if err.Error() != "invalid character 'o' in literal null (expecting 'u')" {
		t.Errorf("not expected error: %s", err)
	}
	if len(indexes) != 0 {
		t.Errorf("index list does not look as expected: %#v", indexes)
	}
}

func TestClientCatIndicesConnectionError(t *testing.T) {
	es := &Client{
		Endpoint: "http://127.0.0.1:0/fail/hard",
	}
	indexes, err := es.CatIndices()
	if err == nil {
		t.Errorf("no error was returned")
	}
	if err.Error() != "Get http://127.0.0.1:0/fail/hard/_cat/indices?format=json: dial tcp 127.0.0.1:0: connect: connection refused" {
		t.Errorf("not expected error: %s", err)
	}
	if len(indexes) != 0 {
		t.Errorf("index list does not look as expected: %#v", indexes)
	}
}

func TestClientDeleteIndex(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Invalid HTTP Method: %s", r.Method)
		}
		if r.URL.Path != "/filebeat-6.4.3-2018.11.21" {
			t.Errorf("incorrect path used: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	es := &Client{
		Endpoint: ts.URL,
	}
	err := es.DeleteIndex("filebeat-6.4.3-2018.11.21")
	if err != nil {
		t.Error(err)
	}
}

func TestClientDeleteIndexFailedRequest(t *testing.T) {
	es := &Client{
		Endpoint: "http://127.0.0.1:0/fail/hard",
	}
	err := es.DeleteIndex("filebeat-6.4.3-2018.11.21")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestClientDeleteIndexInvalidStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("burn"))
	}))
	defer ts.Close()

	es := &Client{
		Endpoint: ts.URL,
	}
	err := es.DeleteIndex("filebeat-6.4.3-2018.11.21")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	expected := "failed to delete index filebeat-6.4.3-2018.11.21, response code: 500, response body: burn"
	if err.Error() != expected {
		t.Errorf("unexpected error returned: %s", err)
	}
}

func TestClientRestoreSnapshot(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Invalid HTTP Method: %s", r.Method)
		}
		if r.URL.Path != "/_snapshot/repo/snapshot-2018.11.21/_restore" {
			t.Errorf("incorrect path used: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("wait_for_completion"); got != "true" {
			t.Errorf("incorrect format querystring used: %s", got)
		}
	}))
	defer ts.Close()

	es := &Client{
		Endpoint: ts.URL,
	}
	err := es.RestoreSnapshot("repo", "snapshot-2018.11.21")
	if err != nil {
		t.Error(err)
	}
}

func TestClientRestoreSnapshotFailedRequest(t *testing.T) {
	es := &Client{
		Endpoint: "http://127.0.0.1:0/fail/hard",
	}
	err := es.RestoreSnapshot("repo", "filebeat-6.4.3-2018.11.21")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestClientRestoreSnapshotInvalidStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("burn"))
	}))
	defer ts.Close()

	es := &Client{
		Endpoint: ts.URL,
	}
	err := es.RestoreSnapshot("repo", "filebeat-6.4.3-2018.11.21")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	expected := "failed to restore snapshot filebeat-6.4.3-2018.11.21, response code: 500, response body: burn"
	if err.Error() != expected {
		t.Errorf("unexpected error returned: %s", err)
	}
}

func TestClientSnapshotIndex(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Invalid HTTP Method: %s", r.Method)
		}
		name := fmt.Sprintf("index-%s-on-%s", "filebeat-6.4.3-2018.11.21", time.Now().Format(time.RFC3339))
		if r.URL.Path != "/_snapshot/repo/"+name {
			t.Errorf("incorrect path used: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("wait_for_completion"); got != "true" {
			t.Errorf("incorrect format querystring used: %s", got)
		}

		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Invalid Content-Type header: %q", got)
		}
		payload := struct {
			Indices            string `json:"indices"`
			IncludeGlobalState bool   `json:"include_global_state"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("error decoding payload: %s", err)
		}
		if payload.Indices != "filebeat-6.4.3-2018.11.21" {
			t.Errorf("invalid indices received: %s", payload.Indices)
		}
		if payload.IncludeGlobalState != false {
			t.Errorf("invalid config include_global_state: %v", payload.IncludeGlobalState)
		}
	}))
	defer ts.Close()

	es := &Client{
		Endpoint: ts.URL,
	}
	err := es.SnapshotIndex("repo", "filebeat-6.4.3-2018.11.21")
	if err != nil {
		t.Error(err)
	}
}

func TestClientSnapshotIndexFailedRequest(t *testing.T) {
	es := &Client{
		Endpoint: "http://127.0.0.1:0/fail/hard",
	}
	err := es.SnapshotIndex("repo", "filebeat-6.4.3-2018.11.21")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestClientSnapshotIndexInvalidStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("burn"))
	}))
	defer ts.Close()

	es := &Client{
		Endpoint: ts.URL,
	}
	err := es.SnapshotIndex("repo", "filebeat-6.4.3-2018.11.21")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	name := fmt.Sprintf("index-%s-on-%s", "filebeat-6.4.3-2018.11.21", time.Now().Format(time.RFC3339))
	expected := "failed to snapshot " + name + ", response code: 500, response body: burn"
	if err.Error() != expected {
		t.Errorf("unexpected error returned: %s", err)
	}
}
