package bundle

import (
	"fmt"
	"path"

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
	template TemplateBlock
	//	templateString string

	meta                 *BundleMeta
	options              *OptionsConfig
	cloudfoundryTemplate TemplateBlock
	cloudfoundryRaw      *CloudFoundryConfig
}

func (self *BundleTemplate) GetMeta() *BundleMeta {
	return self.meta
}

func (self *BundleTemplate) GetCloudFoundryPlans() []*CloudFoundryPlan {
	// Plans are not templated
	if self.cloudfoundryRaw == nil {
		return nil
	}
	return self.cloudfoundryRaw.Plans
}

func (self *BundleTemplate) GetCloudFoundryCredentials(properties map[string]string) (map[string]string, error) {
	if self.cloudfoundryTemplate == nil {
		return nil, fmt.Errorf("No cloudfoundry block configured")
	}

	credentialsTemplate := self.cloudfoundryTemplate.Get("credentials")
	if credentialsTemplate == nil {
		return nil, fmt.Errorf("No cloudfoundry credentials block configured")
	}

	rendered, err := credentialsTemplate.Render(properties)
	if err != nil {
		log.Warn("Error running cloudfoundry credentials template", err)
		return nil, err
	}

	return asStringMap(rendered), nil
}

func (self *BundleTemplate) executeTemplate(context *TemplateContext) (map[string]interface{}, error) {
	//	t, err := template.New("bundle").Parse(templateString)
	//	if err != nil {
	//		return nil, err
	//	}

	//	log.Debug("Executing bundle template: %v", serviceType)

	var err error

	result, err := self.template.Render(context)

	if err != nil {
		log.Warn("Error applying template", err)
		return nil, err
	}

	log.Debug("Applied template: %v", result)

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		log.Warn("Template did not produce map type: %T", result)
		return nil, fmt.Errorf("Unexpected result from template")
	}

	//	config := map[string]interface{}{}
	//	err := goyaml.Unmarshal([]byte(yaml), &config)
	//	if err != nil {
	//		return nil, err
	//	}

	//	var buffer bytes.Buffer
	//	err := self.template.Execute(&buffer, &templateContextCopy)
	//	if err != nil {
	//		return nil, err
	//	}

	//	yaml := buffer.String()
	//	log.Debug("Bundle is:\n%v", yaml)

	return resultMap, nil
}

func NewBundleTemplate(templateData sources.ByteSource) (*BundleTemplate, error) {
	self := &BundleTemplate{}

	templateString, err := sources.ReadToString(templateData)
	if err != nil {
		return nil, err
	}

	//	log.Debug("Reading template: %v", templateString)

	template, err := parseYamlTemplate(templateString)
	if err != nil {
		return nil, err
	}

	self.template = template

	meta := self.template.Remove("meta")
	if meta != nil {
		self.meta, err = parseMeta(meta.Raw())
		if err != nil {
			return nil, err
		}
	}
	options := self.template.Remove("options")
	if options != nil {
		self.options, err = parseOptions(options.Raw())
		if err != nil {
			return nil, err
		}
	}
	cloudfoundry := self.template.Remove("cloudfoundry")
	if cloudfoundry != nil {
		self.cloudfoundryTemplate = cloudfoundry
		self.cloudfoundryRaw, err = parseCloudFoundryConfig(cloudfoundry.Raw())
		if err != nil {
			return nil, err
		}
	}
	//	self.templateString = templateString

	return self, nil
}
