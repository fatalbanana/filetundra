package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatalbanana/filetundra/contrib/httprange"
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

	return idx.DocumentMatchToFileInfo(reader, next)
}

func writeHeaders(w http.ResponseWriter, fi idx.FileInfo) {
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", fi.MimeType)
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

	if r.Method == http.MethodHead {
		writeHeaders(w, fi)
		return
	}

	rangeHdr := r.Header.Get("Range")
	var ranges []httprange.Range
	var haveRange bool
	if rangeHdr != "" {
		haveRange = true
		ranges, err = httprange.ParseRange(rangeHdr, fi.Size)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		if len(ranges) > 1 {
			http.Error(w, "multiple ranges unimplemented", http.StatusNotImplemented)
			return
		}
	}

	f, err := os.Open(fi.Filename)
	if err != nil {
		log.Logger.Error("failed to open file",
			zap.Error(err), zap.String("path", fi.Filename))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	if haveRange {
		_, err := f.Seek(ranges[0].Start, 0)
		if err != nil {
			log.Logger.Error("seek failed", zap.Error(err))
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Range",
			fmt.Sprintf("bytes %d-%d/%d",
				ranges[0].Start,
				ranges[0].Start+ranges[0].Length,
				fi.Size))
	}
	writeHeaders(w, fi)
	if haveRange {
		w.Header().Set("Content-Length", strconv.FormatInt(ranges[0].Length, 10))
		w.WriteHeader(http.StatusPartialContent)
		_, err = io.CopyN(w, f, ranges[0].Length)
	} else {
		w.Header().Set("Content-Length", strconv.FormatInt(fi.Size, 10))
		_, err = io.Copy(w, f)
	}
	if err != nil {
		log.Logger.Error("error serving file",
			zap.Error(err), zap.String("path", fi.Filename))
		panic(http.ErrAbortHandler)
	}
}
