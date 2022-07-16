package web

import (
	"html/template"
	"net/http"
	"path"
	"strings"

	"github.com/fatalbanana/filetundra/internal/env"
	"github.com/fatalbanana/filetundra/internal/idx"
	"github.com/fatalbanana/filetundra/internal/log"
	"github.com/fatalbanana/filetundra/internal/properties"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search"
	"go.uber.org/zap"
)

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "expected POST", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t, err := template.New("browse").Parse(browseTemplate)
	if err != nil {
		log.Logger.Error("error preparing browse template",
			zap.Error(err))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	reader, err := bluge.OpenReader(idx.BlugeConfig)
	if err != nil {
		log.Logger.Error("error opening reader", zap.Error(err))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	searchQ := r.Form.Get("search")
	basenameSearch := bluge.NewMatchQuery(searchQ).SetField(properties.BareBasename).SetAnalyzer(idx.BlugeAnalyzer)
	fuzzyBasenameSearch := bluge.NewFuzzyQuery(searchQ).SetField(properties.BareBasename)
	dirnameSearch := bluge.NewFuzzyQuery(searchQ).SetField(properties.Dirname)
	archiveSearch := bluge.NewFuzzyQuery(searchQ).SetField(properties.ArchiveFilename)
	query := bluge.NewBooleanQuery()
	query.AddShould(basenameSearch)
	query.AddShould(fuzzyBasenameSearch)
	query.AddShould(dirnameSearch)
	query.AddShould(archiveSearch)

	searchReq := bluge.NewAllMatches(query)
	searchResults, err := reader.Search(r.Context(), searchReq)
	if err != nil {
		log.Logger.Error("search error", zap.Error(err))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	res := DirectoryListing{
		Autoback:    true,
		SearchValue: searchQ,
	}
	var next *search.DocumentMatch
	var fi idx.FileInfo
	next, err = searchResults.Next()
	for err == nil && next != nil {
		fi, err = documentMatchToFileInfo(reader, next)
		if err != nil {
			log.Logger.Error("match to fileinfo error", zap.Error(err))
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		properBasename := fi.BareBasename + fi.Extname
		fileRes := DirectoryListingFile{
			Name:  properBasename,
			Image: getImage(fi.MimeType),
			Path:  path.Join("/download", strings.TrimPrefix(fi.Filename, env.Env.Root)),
		}
		res.Files = append(res.Files, fileRes)
		next, err = searchResults.Next()
	}

	err = t.Execute(w, res)
	if err != nil {
		log.Logger.Error("error rendering template", zap.Error(err))
		panic(http.ErrAbortHandler)
	}

}
