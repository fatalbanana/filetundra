package idx

import (
	"context"
	"encoding/binary"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatalbanana/filetundra/internal/env"
	"github.com/fatalbanana/filetundra/internal/properties"

	"github.com/blugelabs/bluge"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
)

var (
	ErrNeedsUpdate = errors.New("file is already indexed but needs update")
)

type FileInfo struct {
	ArchiveFilename []string
	AudioAlbum      string
	AudioArtist     string
	AudioTitle      string
	BareBasename    string
	Extname         string
	Dirname         string
	Filename        string
	MimeType        string
	ModTime         time.Time
	Size            int64
}

func FileToDocument(fpath string, d fs.DirEntry) (*bluge.Document, error) {
	doc := bluge.NewDocument(fpath)
	statInfo, err := os.Stat(fpath)
	if err != nil {
		return doc, err
	}
	encodedSize := make([]byte, binary.MaxVarintLen64)
	encodedModTime := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(encodedSize, statInfo.Size())
	encodedSize = encodedSize[:n]
	n = binary.PutVarint(encodedModTime, statInfo.ModTime().Unix())
	encodedModTime = encodedModTime[:n]
	doc.AddField(bluge.NewKeywordField(properties.Size, string(encodedSize)).StoreValue()).
		AddField(bluge.NewKeywordField(properties.ModifiedTime, string(encodedModTime)).StoreValue())
	basename := filepath.Base(fpath)
	extName := filepath.Ext(basename)
	if extName != "" {
		basename = strings.TrimSuffix(basename, extName)
		doc.AddField(bluge.NewKeywordField(properties.Extname, extName).StoreValue())
	}
	doc.AddField(bluge.NewTextField(properties.BareBasename, basename).WithAnalyzer(BlugeAnalyzer).StoreValue()).
		AddField(bluge.NewKeywordField(properties.Dirname, filepath.Dir(fpath)))
	var fType types.Type
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
	maybeProcessAudio(fpath, fType, doc)
	maybeProcessArchive(fpath, fType, doc)
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
	if next != nil {
		fi, err := DocumentMatchToFileInfo(rdr, next)
		if err != nil {
			return false, err
		}
		si, err := os.Stat(fpath)
		if err != nil {
			return false, err
		}
		if si.ModTime().After(fi.ModTime) {
			return true, ErrNeedsUpdate
		}
		return true, nil
	}
	return false, err
}

func haveExistingNone(rdr *bluge.Reader, fpath string, d fs.DirEntry) (bool, error) {
	return false, nil
}

func walk(reader *bluge.Reader, haveExisting func(*bluge.Reader, string, fs.DirEntry) (bool, error)) error {
	batch := bluge.NewBatch()
	err := filepath.WalkDir(env.Env.Root, func(fpath string, d fs.DirEntry, err error) error {
		var needsUpdate bool
		if err != nil {
			return err
		}
		if fpath == env.Env.Root {
			return nil
		}
		doHaveExisting, err := haveExisting(reader, fpath, d)
		if err != nil {
			if err == ErrNeedsUpdate {
				needsUpdate = true
			} else {
				return err
			}
		}
		if doHaveExisting && !needsUpdate {
			return nil
		}
		doc, err := FileToDocument(fpath, d)
		if err != nil {
			return err
		}
		if needsUpdate {
			batch.Update(doc.ID(), doc)
		} else {
			batch.Insert(doc)
		}
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
