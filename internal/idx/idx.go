package idx

import (
	"os"
	"path/filepath"

	"github.com/blugelabs/bluge"
)

var (
	BlugeConfig bluge.Config
)

func GetBlugeDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "filetundra", "filetundra.bluge"), nil
}

func Init(blugeDir string) {
	BlugeConfig = bluge.DefaultConfig(blugeDir)
}
