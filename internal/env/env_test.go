package env

import (
	"os"
	"testing"
)

func init() {
	os.Setenv("FILETUNDRA_ROOT", "/lol")
}

func TestEnv(t *testing.T) {
	err := Process()
	if err != nil {
		t.Fatal(err)
	}
}
