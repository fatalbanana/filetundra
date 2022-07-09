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

	"github.com/fatalbanana/filetundra/internal/env"
	"github.com/fatalbanana/filetundra/internal/idx"
	"github.com/fatalbanana/filetundra/internal/log"
)

func TestMain(m *testing.M) {
	log.SetupLogger()

	tempDir, err := ioutil.TempDir("", "filetundra_test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	_, ourFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("couldn't find path to myself")
	}
	dataRoot := filepath.Join(filepath.Dir(ourFile),
		"..", "..", "testdata", "fileroot")
	env.Env.Root = dataRoot

	idx.Init(tempDir)
	err = idx.Initial()
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestBrowse(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(browseHandler))
	defer ts.Close()

	client := ts.Client()
	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
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
	templateRoot := filepath.Join(filepath.Dir(ourFile),
		"..", "..", "testdata", "rendered")
	f, err := os.Open(filepath.Join(templateRoot, "browse_test.html"))
	if err != nil {
		t.Fatal(err)
	}
	fileBytes, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(responseBytes, fileBytes) {
		t.Fatal("response didn't match expected render")
	}
}
