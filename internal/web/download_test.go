package web

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/fatalbanana/filetundra/internal/env"
)

func TestDownload(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(downloadHandler))
	defer ts.Close()

	client := ts.Client()
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/download/aaa/bbb", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(filepath.Join(env.Env.Root, "aaa", "bbb"))
	if err != nil {
		t.Fatal(err)
	}
	fileBytes, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(responseBytes, fileBytes) {
		t.Fatal("response didn't match file on disk")
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected HTTP status: got %d expected %d",
			resp.StatusCode, http.StatusOK)
	}

	// Test 404
	req, err = http.NewRequest(http.MethodGet, ts.URL+"/download/sniodmnioewjriodsf", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected HTTP status: got %d expected %d",
			resp.StatusCode, http.StatusNotFound)
	}
}
