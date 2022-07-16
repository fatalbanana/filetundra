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
	Autoback    bool
	Back        string
	Name        string
	Files       []DirectoryListingFile
	SearchValue string
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

	if virtualPath != "/" && virtualPath != "" {
		virtualParentDir, _ := path.Split(virtualPath)
		res.Back = path.Join("/browse", virtualParentDir)
	}

	var next *search.DocumentMatch
	var fi idx.FileInfo
	next, err = searchResults.Next()
	for err == nil && next != nil {
		fi, err = idx.DocumentMatchToFileInfo(reader, next)
		if err != nil {
			return res, err
		}
		basePath := "/download"
		if fi.MimeType == "inode/directory" {
			basePath = "/browse"
		}
		properBasename := fi.BareBasename + fi.Extname
		fileRes := DirectoryListingFile{
			Name:  properBasename,
			Image: getImage(fi.MimeType),
			Path:  path.Join(basePath, virtualPath, properBasename),
		}
		res.Files = append(res.Files, fileRes)
		next, err = searchResults.Next()
	}
	return res, err
}

func browseHandler(w http.ResponseWriter, r *http.Request) {
	virtualPath := path.Clean(strings.TrimPrefix(r.URL.Path, "/browse") + "/")
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
	case "application/gzip":
	case "application/vnd.debian.binary-package":
	case "application/vnd.rar":
	case "application/x-7z-compressed":
	case "application/x-rpm":
	case "application/x-tar":
	case "application/x-unix-archive":
	case "application/x-xz":
	case "application/zip":
	case "application/zstd":
		return "/static/icons/archive-unindexed.svg"

	case "inode/directory":
		return "/static/icons/directory.svg"

	case "application/msword":
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "/static/icons/document.svg"

	case "application/x-executable":
	case "application/x-mach-binary":
	case "application/vnd.microsoft.portable-executable":
		return "/static/icons/executable.svg"

	case "application/pdf":
		return "/static/icons/pdf.svg"

	case "application/vnd.ms-powerpoint":
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return "/static/icons/presentation.svg"

	case "application/vnd.ms-excel":
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return "/static/icons/spreadsheet.svg"
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
