package idx

import (
	"testing"
)

func TestGetBlugeDir(t *testing.T) {
	_, err := GetBlugeDir()
	if err != nil {
		t.Fatal(err)
	}
}
