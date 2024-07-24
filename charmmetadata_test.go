// Copyright 2024 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type CharmMetadataSerializationSuite struct {
	SerializationSuite
}

var _ = gc.Suite(&CharmMetadataSerializationSuite{})

func (s *CharmMetadataSerializationSuite) SetUpTest(c *gc.C) {
	s.importName = "charmMetadata"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importCharmMetadata(m)
	}
}

func (s *CharmMetadataSerializationSuite) TestNewCharmMetadata(c *gc.C) {
	args := CharmMetadataArgs{
		Name: "test-charm",
	}
	metadata := newCharmMetadata(args)

	c.Assert(metadata.Name(), gc.Equals, args.Name)
}

func minimalCharmMetadataMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version": 1,
		"name":    "test-charm",
	}
}

func minimalCharmMetadataArgs() CharmMetadataArgs {
	return CharmMetadataArgs{
		Name: "test-charm",
	}
}

func minimalCharmMetadata() *charmMetadata {
	return newCharmMetadata(minimalCharmMetadataArgs())
}

func maximalCharmMetadataMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version":          1,
		"name":             "test-charm",
		"summary":          "A test charm",
		"description":      "A test charm for testing",
		"subordinate":      true,
		"run-as":           "root",
		"assumes":          "{}",
		"min-juju-version": "4.0.0",
		"categories":       []interface{}{"test", "testing"},
		"tags":             []interface{}{"foo", "bar"},
		"terms":            []interface{}{"baz", "qux"},
		"extra-bindings": map[interface{}]interface{}{
			"db": "mysql",
		},
		"relations": map[interface{}]interface{}{
			"db": map[interface{}]interface{}{
				"name":      "db",
				"role":      "provider",
				"interface": "mysql",
				"optional":  true,
				"limit":     1,
				"scope":     "global",
			},
		},
		"storage": map[interface{}]interface{}{
			"tmp": map[interface{}]interface{}{
				"name":         "tmp",
				"description":  "Temporary storage",
				"type":         "filesystem",
				"shared":       true,
				"readonly":     true,
				"count-min":    1,
				"count-max":    2,
				"minimum-size": 1024,
				"location":     "/tmp",
				"properties":   []interface{}{"foo", "bar"},
			},
		},
		"devices": map[interface{}]interface{}{
			"gpu": map[interface{}]interface{}{
				"name":        "gpu",
				"description": "Graphics card",
				"type":        "gpu",
				"count-min":   1,
				"count-max":   2,
			},
		},
		"payloads": map[interface{}]interface{}{
			"logs": map[interface{}]interface{}{
				"name": "logs",
				"type": "log",
			},
		},
		"resources": map[interface{}]interface{}{
			"database": map[interface{}]interface{}{
				"name":        "database",
				"type":        "file",
				"description": "Database dump",
				"path":        "/var/lib/sqlite",
			},
		},
		"containers": map[interface{}]interface{}{
			"postgres": map[interface{}]interface{}{
				"resource": "database",
				"mounts": []interface{}{
					map[interface{}]interface{}{
						"storage":  "tmp",
						"location": "/var/lib/postgres",
					},
				},
			},
		},
	}
}

func maximalCharmMetadataArgs() CharmMetadataArgs {
	return CharmMetadataArgs{
		Name:           "test-charm",
		Summary:        "A test charm",
		Description:    "A test charm for testing",
		Subordinate:    true,
		RunAs:          "root",
		Assumes:        "{}",
		MinJujuVersion: "4.0.0",
		Categories:     []string{"test", "testing"},
		Tags:           []string{"foo", "bar"},
		Terms:          []string{"baz", "qux"},
		ExtraBindings: map[string]string{
			"db": "mysql",
		},
		Relations: map[string]CharmMetadataRelation{
			"db": charmMetadataRelation{
				Name_:      "db",
				Role_:      "provider",
				Interface_: "mysql",
				Optional_:  true,
				Limit_:     1,
				Scope_:     "global",
			},
		},
		Storage: map[string]CharmMetadataStorage{
			"tmp": charmMetadataStorage{
				Name_:        "tmp",
				Description_: "Temporary storage",
				Type_:        "filesystem",
				Shared_:      true,
				Readonly_:    true,
				CountMin_:    1,
				CountMax_:    2,
				MinimumSize_: 1024,
				Location_:    "/tmp",
				Properties_:  []string{"foo", "bar"},
			},
		},
		Devices: map[string]CharmMetadataDevice{
			"gpu": charmMetadataDevice{
				Name_:        "gpu",
				Description_: "Graphics card",
				Type_:        "gpu",
				CountMin_:    1,
				CountMax_:    2,
			},
		},
		Payloads: map[string]CharmMetadataPayload{
			"logs": charmMetadataPayload{
				Name_: "logs",
				Type_: "log",
			},
		},
		Resources: map[string]CharmMetadataResource{
			"database": charmMetadataResource{
				Name_:        "database",
				Type_:        "file",
				Description_: "Database dump",
				Path_:        "/var/lib/sqlite",
			},
		},
		Containers: map[string]CharmMetadataContainer{
			"postgres": charmMetadataContainer{
				Resource_: "database",
				Mounts_: []charmMetadataContainerMount{
					{
						Storage_:  "tmp",
						Location_: "/var/lib/postgres",
					},
				},
			},
		},
	}
}

func maximalCharmMetadata() *charmMetadata {
	return newCharmMetadata(maximalCharmMetadataArgs())
}

func partialCharmMetadataArgs() CharmMetadataArgs {
	return CharmMetadataArgs{
		Name:        "test-charm",
		Summary:     "A test charm",
		Description: "A test charm for testing",
		Subordinate: true,
		RunAs:       "root",
		Assumes:     "{}",
		Tags:        []string{"foo", "bar"},
		Terms:       []string{"baz", "qux"},
		ExtraBindings: map[string]string{
			"db": "mysql",
		},
		Relations: map[string]CharmMetadataRelation{
			"db": charmMetadataRelation{
				Name_:      "db",
				Role_:      "provider",
				Interface_: "mysql",
				Scope_:     "global",
			},
		},
		Storage: map[string]CharmMetadataStorage{
			"tmp": charmMetadataStorage{
				Name_:       "tmp",
				Type_:       "filesystem",
				Shared_:     true,
				Readonly_:   true,
				CountMin_:   1,
				Location_:   "/tmp",
				Properties_: []string{"foo", "bar"},
			},
		},
		Devices: map[string]CharmMetadataDevice{
			"gpu": charmMetadataDevice{
				Name_:        "gpu",
				Description_: "Graphics card",
				CountMax_:    2,
			},
		},
		Payloads: map[string]CharmMetadataPayload{
			"logs": charmMetadataPayload{
				Name_: "logs",
			},
		},
		Resources: map[string]CharmMetadataResource{
			"database": charmMetadataResource{
				Name_: "database",
				Type_: "file",
			},
		},
		Containers: map[string]CharmMetadataContainer{
			"postgres": charmMetadataContainer{
				Resource_: "database",
				Mounts_: []charmMetadataContainerMount{
					{
						Storage_: "tmp",
					},
				},
			},
		},
	}
}

func partialCharmMetadata() *charmMetadata {
	return newCharmMetadata(partialCharmMetadataArgs())
}

func (s *CharmMetadataSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalCharmMetadata())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalCharmMetadataMap())
}

func (s *CharmMetadataSerializationSuite) TestMaximalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(maximalCharmMetadata())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, maximalCharmMetadataMap())
}

func (s *CharmMetadataSerializationSuite) TestMinimalParsingSerializedData(c *gc.C) {
	initial := minimalCharmMetadata()
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importCharmMetadata(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}

func (s *CharmMetadataSerializationSuite) TestMaximalParsingSerializedData(c *gc.C) {
	initial := maximalCharmMetadata()
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importCharmMetadata(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}

func (s *CharmMetadataSerializationSuite) TestPartialParsingSerializedData(c *gc.C) {
	initial := partialCharmMetadata()
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importCharmMetadata(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}

func (s *CharmMetadataSerializationSuite) exportImportVersion(c *gc.C, origin_ *charmMetadata, version int) *charmMetadata {
	origin_.Version_ = version
	bytes, err := yaml.Marshal(origin_)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	origin, err := importCharmMetadata(source)
	c.Assert(err, jc.ErrorIsNil)
	return origin
}

func (s *CharmMetadataSerializationSuite) TestV1ParsingReturnsLatest(c *gc.C) {
	args := maximalCharmMetadataArgs()
	originV1 := newCharmMetadata(args)

	originLatest := *originV1
	originResult := s.exportImportVersion(c, originV1, 1)
	c.Assert(*originResult, jc.DeepEquals, originLatest)
}
