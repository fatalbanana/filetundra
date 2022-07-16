package idx

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"time"
	"unicode"

	"github.com/fatalbanana/filetundra/internal/properties"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/lang/en"
	"github.com/blugelabs/bluge/analysis/token"
	"github.com/blugelabs/bluge/analysis/tokenizer"
	"github.com/blugelabs/bluge/search"
)

var (
	BlugeAnalyzer *analysis.Analyzer
	BlugeConfig   bluge.Config
)

func GetBlugeDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "filetundra", "filetundra.bluge"), nil
}

func Init(blugeDir string) {
	BlugeAnalyzer = newAnalyzer()
	BlugeConfig = bluge.DefaultConfig(blugeDir)
}

func newAnalyzer() *analysis.Analyzer {
	// english analyzer with only alphanumerics tokenized
	return &analysis.Analyzer{
		Tokenizer: tokenizer.NewCharacterTokenizer(func(r rune) bool {
			return unicode.IsLetter(r) || unicode.IsNumber(r)
		}),
		TokenFilters: []analysis.TokenFilter{
			en.NewPossessiveFilter(),
			token.NewLowerCaseFilter(),
			en.StopWordsFilter(),
			en.StemmerFilter(),
		},
	}

}

func DocumentMatchToFileInfo(reader *bluge.Reader, match *search.DocumentMatch) (FileInfo, error) {
	fi := FileInfo{}
	err := reader.VisitStoredFields(match.Number, func(field string, value []byte) bool {
		switch field {
		case "_id":
			fi.Filename = string(value)
		case properties.BareBasename:
			fi.BareBasename = string(value)
		case properties.Extname:
			fi.Extname = string(value)
		case properties.MimeType:
			fi.MimeType = string(value)
		case properties.AudioAlbum:
			fi.AudioAlbum = string(value)
		case properties.AudioArtist:
			fi.AudioArtist = string(value)
		case properties.AudioTitle:
			fi.AudioTitle = string(value)
		case properties.Size:
			sz, bytesRead := binary.Varint(value)
			if bytesRead != 0 {
				fi.Size = sz
			}
		case properties.ModifiedTime:
			mt, bytesRead := binary.Varint(value)
			if bytesRead != 0 {
				fi.ModTime = time.Unix(mt, 0)
			}
		}
		return true
	})
	return fi, err
}
