package idx

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fatalbanana/filetundra/internal/env"
	"github.com/fatalbanana/filetundra/internal/log"
)

func TestMain(m *testing.M) {
	log.SetupLogger()
	os.Exit(m.Run())
}

func TestMakeUpdateIndex(t *testing.T) {
	_, ourFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("couldn't find path to myself")
	}

	dataRoot := filepath.Join(filepath.Dir(ourFile),
		"..", "..", "testdata", "fileroot")
	env.Env.Root = dataRoot

	tempDir, err := ioutil.TempDir("", "filetundra_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	Init(tempDir)
	err = Initial()
	if err != nil {
		t.Fatal(err)
	}

	err = Update()
	if err != nil {
		t.Fatal(err)
	}
}
