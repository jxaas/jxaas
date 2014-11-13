package core

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/justinsb/gova/log"
	"github.com/justinsb/gova/sources"
)

type CharmReader struct {
	byteSource sources.ByteSource
}

func NewCharmReader(byteSource sources.ByteSource) *CharmReader {
	self := &CharmReader{}
	self.byteSource = byteSource
	return self
}

func (self *CharmReader) read(name string) ([]byte, error) {
	inputStream, err := self.byteSource.Open()
	if err != nil {
		return nil, err
	}

	defer func() {
		closeable, ok := inputStream.(io.Closer)
		if ok {
			closeable.Close()
		}
	}()

	size, err := self.byteSource.Size()
	if err != nil {
		return nil, err
	}

	readerAt, ok := inputStream.(io.ReaderAt)
	if !ok {
		return nil, fmt.Errorf("Expected ReaderAt")
	}

	r, err := zip.NewReader(readerAt, size)
	if err != nil {
		return nil, err
	}

	for _, f := range r.File {
		log.Info("File: %v", f.Name)
	}

	for _, f := range r.File {
		if f.Name != name {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()

		data, err := ioutil.ReadAll(rc)
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	return nil, nil
}
