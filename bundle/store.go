package bundle

import (
	"os"
	"path"

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
