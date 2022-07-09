package web

import (
	"context"
	_ "embed"
	"html/template"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/fatalbanana/filetundra/internal/env"
	"github.com/fatalbanana/filetundra/internal/idx"
	"github.com/fatalbanana/filetundra/internal/log"
	"github.com/fatalbanana/filetundra/internal/properties"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search"
	"go.uber.org/zap"
)

//go:embed templates/browse.html
var browseTemplate string

type DirectoryListing struct {
	Name  string
	Files []DirectoryListingFile
}

type DirectoryListingFile struct {
	Name  string
	Image string
	Path  string
}

func pathToDirectoryListing(ctx context.Context, searchPath string, virtualPath string) (res DirectoryListing, err error) {
	reader, err := bluge.OpenReader(idx.BlugeConfig)
	if err != nil {
		return res, err
	}
	defer reader.Close()

	query := bluge.NewTermQuery(searchPath).SetField(properties.Dirname)
	searchReq := bluge.NewAllMatches(query)
	searchResults, err := reader.Search(ctx, searchReq)
	if err != nil {
		return res, err
	}

	res.Name = virtualPath
	res.Files = make([]DirectoryListingFile, 0)

	var next *search.DocumentMatch
	var fi idx.FileInfo
	next, err = searchResults.Next()
	for err == nil && next != nil {
		fi, err = documentMatchToFileInfo(reader, next)
		if err != nil {
			return res, err
		}
		basePath := "/download"
		if fi.MimeType == "inode/directory" {
			basePath = "/browse"
		}
		fileRes := DirectoryListingFile{
			Name:  fi.Basename,
			Image: getImage(fi.MimeType),
			Path:  path.Join(basePath, virtualPath, fi.Basename),
		}
		res.Files = append(res.Files, fileRes)
		next, err = searchResults.Next()
	}
	return res, err
}

func browseHandler(w http.ResponseWriter, r *http.Request) {
	virtualPath := strings.TrimPrefix(r.URL.Path, "/browse")
	searchPath := filepath.Join(env.Env.Root, virtualPath)

	t, err := template.New("browse").Parse(browseTemplate)
	if err != nil {
		log.Logger.Error("error preparing browse template",
			zap.String("directory", searchPath),
			zap.Error(err))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	var res DirectoryListing
	res, err = pathToDirectoryListing(r.Context(), searchPath, virtualPath)
	if err != nil {
		log.Logger.Error("error fetching directory listing",
			zap.String("directory", searchPath),
			zap.Error(err))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, res)
	if err != nil {
		log.Logger.Error("error rendering template", zap.Error(err))
		panic(http.ErrAbortHandler)
	}
}

func getImage(mType string) string {
	switch mType {
	case "inode/directory":
		return "/static/icons/directory.svg"
	}
	majorVersion := strings.Split(mType, "/")[0]
	switch majorVersion {
	case "image":
		return "/static/icons/image.svg"
	case "audio":
		return "/static/icons/audio.svg"
	case "video":
		return "/static/icons/video.svg"
	}
	return "/static/icons/text.svg"
}
