package core

import (
	"archive/zip"
	"io/ioutil"

	"github.com/justinsb/gova/log"
)

type CharmFile struct {
	path string
}

func NewCharmFile(path string) *CharmFile {
	self := &CharmFile{}
	self.path = path
	return self
}

func (self *CharmFile) read(name string) ([]byte, error) {
	r, err := zip.OpenReader(self.path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

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

		//		path := filepath.Join(dest, f.Name)
		//		if f.FileInfo().IsDir() {
		//			os.MkdirAll(path, f.Mode())
		//		} else {
		//			f, err := os.OpenFile(
		//				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		//			if err != nil {
		//				return err
		//			}
		//			defer f.Close()
		//
		//			_, err = io.Copy(f, rc)
		//			if err != nil {
		//				return err
		//			}
		//		}

		data, err := ioutil.ReadAll(rc)
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	return nil, nil
}
