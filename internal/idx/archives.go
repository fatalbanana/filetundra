package idx

import (
	"archive/zip"

	"github.com/fatalbanana/filetundra/internal/log"
	"github.com/fatalbanana/filetundra/internal/properties"

	"github.com/blugelabs/bluge"
	"github.com/h2non/filetype/matchers"
	"github.com/h2non/filetype/types"
	"go.uber.org/zap"
)

func getZipFiles(fpath string) ([]string, error) {
	z, err := zip.OpenReader(fpath)
	if err != nil {
		return []string{}, err
	}
	zipNumFiles := len(z.File)
	res := make([]string, zipNumFiles)
	for i := 0; i < zipNumFiles; i++ {
		res[i] = z.File[i].FileHeader.Name
	}
	return res, z.Close()
}

func maybeProcessArchive(fpath string, fType types.Type, doc *bluge.Document) {
	var res []string
	var err error

	switch fType {
	case matchers.TypeZip:
		res, err = getZipFiles(fpath)
		if err != nil {
			log.Logger.Error("error reading zip file contents",
				zap.String("path", fpath), zap.Error(err))
			return
		}
	default:
		return
	}

	for _, e := range res {
		doc.AddField(bluge.NewTextField(properties.ArchiveFilename, e).StoreValue())
	}
}
