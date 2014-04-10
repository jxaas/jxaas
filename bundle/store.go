package bundle

import (
	"bytes"
	"fmt"
	"path"

	"os/filepath"
	"text/template"

	"github.com/justinsb/gova/log"
)

type BundleStore struct {
	basedir string
}

func (self *BundleStore) getBundleTemplate(key string) (*Bundle, error) {
	var def string

	// TODO: Check for path traversal
	path := path.Join(self.basedir, key + ".yaml")

	def, err := files.Read(path)
	if err != nil {
		return nil, err
	}

	if def == nil {
		return nil, nil
	}

	// TODO: Cache templates
	template, err := template.New("bundle-" + key).Parse(def)
	if err != nil {
		return nil, err
	}

	return template, nil
}
