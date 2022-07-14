package idx

import (
	"os"
	"path/filepath"
	"unicode"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/lang/en"
	"github.com/blugelabs/bluge/analysis/token"
	"github.com/blugelabs/bluge/analysis/tokenizer"
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
