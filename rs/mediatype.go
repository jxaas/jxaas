package rs

import (
	"fmt"
	"mime"
	"strings"
)

type MediaType struct {
	Type       string
	Subtype    string
	Parameters map[string]string

	match string
}

// TODO: Make immutable, use cache / well-known values?

func ParseMediaType(t string) (*MediaType, error) {
	m, params, err := mime.ParseMediaType(t)
	if err != nil {
		return nil, err
	}

	tokens := strings.Split(m, "/")
	if len(tokens) != 2 {
		return nil, fmt.Errorf("Cannot parse media type: %v", m)
	}

	self := &MediaType{}
	self.Type = tokens[0]
	self.Subtype = tokens[1]
	self.match = m
	self.Parameters = params

	return self, nil
}

func (self *MediaType) Is(match string) bool {
	return self.match == match
}
