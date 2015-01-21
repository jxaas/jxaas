package bundle

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"launchpad.net/goyaml"

	"github.com/justinsb/gova/log"
)

type TemplateBlock interface {
	Render(context interface{}) (interface{}, error)
	Raw() interface{}

	Get(key string) TemplateBlock
	Remove(key string) TemplateBlock
}

type StringTemplateBlock struct {
	template *template.Template
	text     string
}

func newStringTemplateBlock(s string) (*StringTemplateBlock, error) {
	t, err := template.New("template").Parse(s)
	if err != nil {
		log.Warn("Error parsing template (%v)", asString, err)
		return nil, err
	}

	self := &StringTemplateBlock{}
	self.template = t
	self.text = s
	return self, nil
}

func (self *StringTemplateBlock) Remove(key string) TemplateBlock {
	return nil
}

func (self *StringTemplateBlock) Get(key string) TemplateBlock {
	return nil
}

func (self *StringTemplateBlock) Render(context interface{}) (interface{}, error) {
	// TODO: Replace the "<no value>" incorrect placeholders
	// https://code.google.com/p/go/issues/detail?id=6288

	var buffer bytes.Buffer
	err := self.template.Execute(&buffer, context)
	if err != nil {
		log.Warn("Error executing template (%v)", self.text, err)
		return nil, err
	}

	s := buffer.String()

	return s, nil
}

func (self *StringTemplateBlock) Raw() interface{} {
	return self.text
}

type ValueTemplateBlock struct {
	value interface{}
	text  string
}

func newValueTemplateBlock(value interface{}, text string) (*ValueTemplateBlock, error) {
	self := &ValueTemplateBlock{}
	self.value = value
	self.text = text
	return self, nil
}

func (self *ValueTemplateBlock) Remove(key string) TemplateBlock {
	return nil
}

func (self *ValueTemplateBlock) Get(key string) TemplateBlock {
	return nil
}

func (self *ValueTemplateBlock) Render(context interface{}) (interface{}, error) {
	return self.value, nil
}

func (self *ValueTemplateBlock) Raw() interface{} {
	return self.text
}

type MapTemplateBlock struct {
	template map[string]TemplateBlock
}

func newMapTemplateBlock(template map[string]TemplateBlock) (*MapTemplateBlock, error) {
	self := &MapTemplateBlock{}
	self.template = template
	return self, nil
}

func (self *MapTemplateBlock) Remove(key string) TemplateBlock {
	v, ok := self.template[key]
	if ok {
		delete(self.template, key)
	}
	return v
}

func (self *MapTemplateBlock) Get(key string) TemplateBlock {
	v, _ := self.template[key]
	return v
}

func (self *MapTemplateBlock) Render(context interface{}) (interface{}, error) {
	result := map[string]interface{}{}
	for k, v := range self.template {
		rendered, err := v.Render(context)
		if err != nil {
			return nil, err
		}
		result[k] = rendered
	}
	return result, nil
}

func (self *MapTemplateBlock) Raw() interface{} {
	result := map[string]interface{}{}
	for k, v := range self.template {
		result[k] = v.Raw()
	}
	return result
}

type ArrayTemplateBlock struct {
	template []TemplateBlock
}

func newArrayTemplateBlock(template []TemplateBlock) (*ArrayTemplateBlock, error) {
	self := &ArrayTemplateBlock{}
	self.template = template
	return self, nil
}

func (self *ArrayTemplateBlock) Remove(key string) TemplateBlock {
	return nil
}

func (self *ArrayTemplateBlock) Get(key string) TemplateBlock {
	return nil
}

func (self *ArrayTemplateBlock) Render(context interface{}) (interface{}, error) {
	result := []interface{}{}
	for _, v := range self.template {
		rendered, err := v.Render(context)
		if err != nil {
			return nil, err
		}
		result = append(result, rendered)
	}
	return result, nil
}

func (self *ArrayTemplateBlock) Raw() interface{} {
	result := []interface{}{}
	for _, v := range self.template {
		result = append(result, v.Raw())
	}
	return result
}

func toTemplateBlock(src interface{}) (TemplateBlock, error) {
	asMap, ok := src.(map[interface{}]interface{})
	if ok {
		blocks := map[string]TemplateBlock{}
		for k, v := range asMap {
			block, err := toTemplateBlock(v)
			if err != nil {
				return nil, err
			}
			blocks[asString(k)] = block
		}
		return newMapTemplateBlock(blocks)
	}

	asMapStringInterface, ok := src.(map[string]interface{})
	if ok {
		blocks := map[string]TemplateBlock{}
		for k, v := range asMapStringInterface {
			block, err := toTemplateBlock(v)
			if err != nil {
				return nil, err
			}
			blocks[k] = block
		}
		return newMapTemplateBlock(blocks)
	}

	asArray, ok := src.([]interface{})
	if ok {
		blocks := []TemplateBlock{}
		for _, v := range asArray {
			block, err := toTemplateBlock(v)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, block)
		}
		return newArrayTemplateBlock(blocks)
	}

	asString, ok := src.(string)
	if ok {
		s := asString

		if strings.Contains(s, "__JXAAS_BEGIN_TEMPLATE__") {
			s = strings.Replace(s, "__JXAAS_BEGIN_TEMPLATE__", "{{", -1)
			s = strings.Replace(s, "__JXAAS_END_TEMPLATE__", "}}", -1)

			return newStringTemplateBlock(s)
		} else {
			return newValueTemplateBlock(s, s)
		}
	}

	asBool, ok := src.(bool)
	if ok {
		v := asBool
		text := fmt.Sprintf("%v", v)
		return newValueTemplateBlock(v, text)
	}

	asInt, ok := src.(int)
	if ok {
		v := asInt
		text := fmt.Sprintf("%v", v)
		return newValueTemplateBlock(v, text)
	}

	return nil, fmt.Errorf("Unhandled type in template: %T", src)
}

func parseYamlTemplate(yaml string) (TemplateBlock, error) {
	replaced := strings.Replace(yaml, "{{", "__JXAAS_BEGIN_TEMPLATE__", -1)
	replaced = strings.Replace(replaced, "}}", "__JXAAS_END_TEMPLATE__", -1)

	parsed := map[string]interface{}{}
	err := goyaml.Unmarshal([]byte(replaced), &parsed)
	if err != nil {
		return nil, err
	}

	return toTemplateBlock(parsed)
}
