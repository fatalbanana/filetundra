package web

import (
	"embed"
	"io"
	"net/http"
	"strings"

	"github.com/fatalbanana/filetundra/internal/log"

	"go.uber.org/zap"
)

//go:embed static/css/tundra.css
//go:embed static/icons/archive.svg
//go:embed static/icons/audio.svg
//go:embed static/icons/back.svg
//go:embed static/icons/directory.svg
//go:embed static/icons/document.svg
//go:embed static/icons/executable.svg
//go:embed static/icons/find.svg
//go:embed static/icons/image.svg
//go:embed static/icons/pdf.svg
//go:embed static/icons/presentation.svg
//go:embed static/icons/spreadsheet.svg
//go:embed static/icons/text.svg
//go:embed static/icons/video.svg

var efs embed.FS

func staticHandler(w http.ResponseWriter, r *http.Request) {
	virtualPath := strings.TrimPrefix(r.URL.Path, "/")
	f, err := efs.Open(virtualPath)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if strings.HasPrefix(virtualPath, "static/icons/") {
		w.Header().Set("Content-type", "image/svg+xml")
	}
	_, err = io.Copy(w, f)
	if err != nil {
		log.Logger.Error("error serving embedded file",
			zap.Error(err), zap.String("path", virtualPath))
		panic(http.ErrAbortHandler)
	}
}
