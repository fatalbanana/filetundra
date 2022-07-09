package web

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatalbanana/filetundra/internal/env"
	"github.com/fatalbanana/filetundra/internal/idx"
	"github.com/fatalbanana/filetundra/internal/log"

	"github.com/blugelabs/bluge"
	"go.uber.org/zap"
)

var (
	errNotFound = errors.New("path was not found")
)

func pathToFileInfo(ctx context.Context, searchPath string) (dud idx.FileInfo, err error) {
	var reader *bluge.Reader

	reader, err = bluge.OpenReader(idx.BlugeConfig)
	if err != nil {
		return dud, err
	}
	defer reader.Close()

	query := bluge.NewTermQuery(searchPath).SetField("_id")
	searchReq := bluge.NewAllMatches(query)
	searchResults, err := reader.Search(ctx, searchReq)
	if err != nil {
		return dud, err
	}

	next, err := searchResults.Next()
	if err != nil {
		return dud, err
	}
	if next == nil {
		return dud, errNotFound
	}

	return documentMatchToFileInfo(reader, next)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	virtualPath := strings.TrimPrefix(r.URL.Path, "/download")
	searchPath := filepath.Join(env.Env.Root, virtualPath)

	fi, err := pathToFileInfo(r.Context(), searchPath)
	if err != nil {
		if err == errNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		log.Logger.Error("error fetching path info", zap.Error(err))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	f, err := os.Open(fi.Filename)
	if err != nil {
		log.Logger.Error("failed to open file",
			zap.Error(err), zap.String("path", fi.Filename))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", fi.MimeType)
	_, err = io.Copy(w, f)
	if err != nil {
		log.Logger.Error("error serving file",
			zap.Error(err), zap.String("path", fi.Filename))
		panic(http.ErrAbortHandler)
	}
}
