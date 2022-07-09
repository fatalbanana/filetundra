package idx

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fatalbanana/filetundra/internal/env"
	"github.com/fatalbanana/filetundra/internal/log"
	"github.com/fatalbanana/filetundra/internal/properties"

	"github.com/blugelabs/bluge"
	"github.com/dhowden/tag"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers"
	"github.com/h2non/filetype/types"
	"go.uber.org/zap"
)

var (
	audioMetaMimeMap = map[types.Type]struct{}{
		matchers.TypeFlac: struct{}{},
		matchers.TypeM4a:  struct{}{},
		matchers.TypeMp3:  struct{}{},
		matchers.TypeOgg:  struct{}{},
	}

	errUnhandledAudioFormat = errors.New("unrecognised audio format")
)

type FileInfo struct {
	AudioAlbum  string
	AudioArtist string
	AudioTitle  string
	Basename    string
	Dirname     string
	Filename    string
	MimeType    string
}

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

func FileToDocument(fpath string, d fs.DirEntry) (*bluge.Document, error) {
	doc := bluge.NewDocument(fpath)
	doc.AddField(bluge.NewTextField(properties.Basename, filepath.Base(fpath)).StoreValue())
	doc.AddField(bluge.NewKeywordField(properties.Dirname, filepath.Dir(fpath)))
	var fType types.Type
	var err error
	if d.IsDir() {
		doc.AddField(bluge.NewTextField(properties.MimeType, "inode/directory").StoreValue())
	} else {
		fType, err = filetype.MatchFile(fpath)
		if err != nil {
			return doc, err
		} else {
			if fType.MIME.Value == "" {
				doc.AddField(bluge.NewTextField(properties.MimeType, "application/octet-stream").StoreValue()) // XXX: text?
			} else {
				doc.AddField(bluge.NewTextField(properties.MimeType, fType.MIME.Value).StoreValue())
			}
		}
	}
	_, ok := audioMetaMimeMap[fType]
	if ok {
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
	return doc, nil
}

func haveExisting(rdr *bluge.Reader, fpath string, d fs.DirEntry) (bool, error) {
	query := bluge.NewTermQuery(fpath).SetField("_id")
	search := bluge.NewAllMatches(query)
	searchResults, err := rdr.Search(context.TODO(), search)
	if err != nil {
		return false, err
	}
	next, err := searchResults.Next()
	return (next != nil), err
}

func haveExistingNone(rdr *bluge.Reader, fpath string, d fs.DirEntry) (bool, error) {
	return false, nil
}

func walk(reader *bluge.Reader, haveExisting func(*bluge.Reader, string, fs.DirEntry) (bool, error)) error {
	batch := bluge.NewBatch()
	err := filepath.WalkDir(env.Env.Root, func(fpath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if fpath == env.Env.Root {
			return nil
		}
		doHaveExisting, err := haveExisting(reader, fpath, d)
		if err != nil {
			return err
		}
		if doHaveExisting {
			return nil
		}
		doc, err := FileToDocument(fpath, d)
		if err != nil {
			return err
		}
		batch.Insert(doc)
		return nil
	})
	if err != nil {
		return err
	}
	writer, err := bluge.OpenWriter(BlugeConfig)
	if err != nil {
		return err
	}
	defer writer.Close()
	return writer.Batch(batch)
}

func Initial() error {
	return walk(nil, haveExistingNone)
}

func Update() error {
	reader, err := bluge.OpenReader(BlugeConfig)
	if err != nil {
		return err
	}
	defer reader.Close()
	return walk(reader, haveExisting)
}
