// Copyright 2024 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type CharmConfigsSerializationSuite struct {
	SerializationSuite
}

var _ = gc.Suite(&CharmConfigsSerializationSuite{})

func (s *CharmConfigsSerializationSuite) SetUpTest(c *gc.C) {
	s.importName = "charmConfigs"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importCharmConfigs(m)
	}
}

func (s *CharmConfigsSerializationSuite) TestNewCharmConfigs(c *gc.C) {
	args := CharmConfigsArgs{
		Configs: map[string]CharmConfig{
			"foo": charmConfig{
				Description_: "description",
				Type_:        "string",
				Default_:     "default",
			},
		},
	}
	metadata := newCharmConfigs(args)

	c.Assert(metadata.Configs(), gc.DeepEquals, map[string]CharmConfig{
		"foo": charmConfig{
			Description_: "description",
			Type_:        "string",
			Default_:     "default",
		},
	})
}

func minimalCharmConfigsMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version": 1,
		"configs": map[interface{}]interface{}{},
	}
}

func minimalCharmConfigsArgs() CharmConfigsArgs {
	return CharmConfigsArgs{}
}

func minimalCharmConfigs() *charmConfigs {
	return newCharmConfigs(minimalCharmConfigsArgs())
}

func maximalCharmConfigsMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version": 1,
		"configs": map[interface{}]interface{}{
			"foo": map[interface{}]interface{}{
				"description": "description",
				"type":        "string",
				"default":     "default",
			},
		},
	}
}

func maximalCharmConfigsArgs() CharmConfigsArgs {
	return CharmConfigsArgs{
		Configs: map[string]CharmConfig{
			"foo": charmConfig{
				Description_: "description",
				Type_:        "string",
				Default_:     "default",
			},
		},
	}
}

func maximalCharmConfigs() *charmConfigs {
	return newCharmConfigs(maximalCharmConfigsArgs())
}

func (s *CharmConfigsSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalCharmConfigs())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalCharmConfigsMap())
}

func (s *CharmConfigsSerializationSuite) TestMaximalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(maximalCharmConfigs())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, maximalCharmConfigsMap())
}

func (s *CharmConfigsSerializationSuite) TestMinimalParsingSerializedData(c *gc.C) {
	initial := minimalCharmConfigs()
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importCharmConfigs(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}

func (s *CharmConfigsSerializationSuite) TestMaximalParsingSerializedData(c *gc.C) {
	initial := maximalCharmConfigs()
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importCharmConfigs(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}

func (s *CharmConfigsSerializationSuite) exportImportVersion(c *gc.C, origin_ *charmConfigs, version int) *charmConfigs {
	origin_.Version_ = version
	bytes, err := yaml.Marshal(origin_)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	origin, err := importCharmConfigs(source)
	c.Assert(err, jc.ErrorIsNil)
	return origin
}

func (s *CharmConfigsSerializationSuite) TestV1ParsingReturnsLatest(c *gc.C) {
	args := maximalCharmConfigsArgs()
	originV1 := newCharmConfigs(args)

	originLatest := *originV1
	originResult := s.exportImportVersion(c, originV1, 1)
	c.Assert(*originResult, jc.DeepEquals, originLatest)
}
