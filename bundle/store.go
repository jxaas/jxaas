package bundle

import (
	"os"
	"path"

	"github.com/justinsb/gova/log"

	"io/ioutil"
	"text/template"
)

type BundleStore struct {
	basedir string
}

func NewBundleStore(basedir string) *BundleStore {
	self := &BundleStore{}
	self.basedir = basedir
	return self
}

func (self *BundleStore) getBundleTemplate(key string) (*template.Template, error) {
	// TODO: Check for path traversal
	path := path.Join(self.basedir, key+".yaml")

	def, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn("Service bundle not found: %v", path)
			return nil, nil
		} else {
			return nil, err
		}
	}

	// TODO: Cache templates
	template, err := template.New("bundle-" + key).Parse(string(def))
	if err != nil {
		return nil, err
	}

	return template, nil
}
