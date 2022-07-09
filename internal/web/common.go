package web

import (
	"github.com/fatalbanana/filetundra/internal/idx"
	"github.com/fatalbanana/filetundra/internal/properties"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search"
)

func documentMatchToFileInfo(reader *bluge.Reader, match *search.DocumentMatch) (idx.FileInfo, error) {
	fi := idx.FileInfo{}
	err := reader.VisitStoredFields(match.Number, func(field string, value []byte) bool {
		switch field {
		case "_id":
			fi.Filename = string(value)
		case properties.Basename:
			fi.Basename = string(value)
		case properties.MimeType:
			fi.MimeType = string(value)
		case properties.AudioAlbum:
			fi.AudioAlbum = string(value)
		case properties.AudioArtist:
			fi.AudioArtist = string(value)
		case properties.AudioTitle:
			fi.AudioTitle = string(value)
		}
		return true
	})
	return fi, err
}
