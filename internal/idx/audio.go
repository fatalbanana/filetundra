package idx

import (
	"errors"
	"os"

	"github.com/fatalbanana/filetundra/internal/log"
	"github.com/fatalbanana/filetundra/internal/properties"

	"github.com/blugelabs/bluge"
	"github.com/dhowden/tag"
	"github.com/h2non/filetype/matchers"
	"github.com/h2non/filetype/types"
	"go.uber.org/zap"
)

var (
	audioMetaMimeMap = map[types.Type]struct{}{
		matchers.TypeFlac: {},
		matchers.TypeM4a:  {},
		matchers.TypeMp3:  {},
		matchers.TypeOgg:  {},
	}

	errUnhandledAudioFormat = errors.New("unhandled audio format")
)

func getAudioMetadata(fpath string, fType types.Type) (tag.Metadata, error) {
	var meta tag.Metadata

	f, err := os.Open(fpath) // #nosec: shut up
	if err != nil {
		return meta, err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Logger.Error("error closing file",
				zap.String("path", fpath), zap.Error(err))
		}
	}()

	switch fType {
	case matchers.TypeFlac:
		meta, err = tag.ReadFLACTags(f)
	case matchers.TypeMp3:
		meta, err = tag.ReadID3v2Tags(f)
		if err != nil {
			meta, err = tag.ReadID3v1Tags(f)
		}
	case matchers.TypeOgg:
		meta, err = tag.ReadOGGTags(f)
	case matchers.TypeM4a:
		meta, err = tag.ReadAtoms(f)
	default:
		err = errUnhandledAudioFormat
	}
	return meta, err
}

func maybeProcessAudio(fpath string, fType types.Type, doc *bluge.Document) {
	_, ok := audioMetaMimeMap[fType]
	if !ok {
		return
	}
	meta, err := getAudioMetadata(fpath, fType)
	if err != nil {
		log.Logger.Error("error getting audio metadata",
			zap.String("path", fpath), zap.Error(err))
	} else {
		artist := meta.Artist()
		if artist != "" {
			doc.AddField(bluge.NewTextField(properties.AudioArtist, artist).StoreValue())
		}
		album := meta.Album()
		if album != "" {
			doc.AddField(bluge.NewTextField(properties.AudioAlbum, album).StoreValue())
		}
		title := meta.Title()
		if title != "" {
			doc.AddField(bluge.NewTextField(properties.AudioTitle, title).StoreValue())
		}
	}
}
