package web

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestStatic(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(staticHandler))
	defer ts.Close()

	client := ts.Client()
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/static/icons/audio.svg", nil)
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

	_, ourFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("couldn't find path to myself")
	}
	iconPath := filepath.Join(filepath.Dir(ourFile),
		"static", "icons", "audio.svg")
	f, err := os.Open(iconPath)
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

	// Test not found
	req, err = http.NewRequest(http.MethodGet, ts.URL+"/static/icons/sniodmnioewjriodsf", nil)
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
