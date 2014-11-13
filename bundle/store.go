package bundle

import (
	"fmt"
	"path"
	"text/template"

	"github.com/justinsb/gova/log"
	"github.com/justinsb/gova/sources"
)

type BundleStore struct {
	basedir string
}

func NewBundleStore(basedir string) *BundleStore {
	self := &BundleStore{}
	self.basedir = basedir
	return self
}

func (self *BundleStore) GetBundleTemplate(key string) (*BundleTemplate, error) {
	// TODO: Check for path traversal
	path := path.Join(self.basedir, key+".yaml")

	bytes := sources.NewFileByteSource(path)
	exists, err := bytes.Exists()
	if err != nil {
		return nil, fmt.Errorf("Error checking for template", err)
	}
	if !exists {
		log.Warn("Service bundle not found: %v", path)
		return nil, nil
	}

	return NewBundleTemplate(bytes)
}

type BundleTemplate struct {
	template *template.Template
}

func NewBundleTemplate(templateData sources.ByteSource) (*BundleTemplate, error) {
	self := &BundleTemplate{}

	templateString, err := sources.ReadToString(templateData)
	if err != nil {
		return nil, err
	}

	t, err := template.New("bundle").Parse(templateString)
	if err != nil {
		return nil, err
	}

	self.template = t
	return self, nil
}
