// Copyright 2024 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type CharmManifestSerializationSuite struct {
	SerializationSuite
}

var _ = gc.Suite(&CharmManifestSerializationSuite{})

func (s *CharmManifestSerializationSuite) SetUpTest(c *gc.C) {
	s.importName = "charmManifest"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importCharmManifest(m)
	}
}

func (s *CharmManifestSerializationSuite) TestNewCharmManifest(c *gc.C) {
	args := CharmManifestArgs{
		Bases: []CharmManifestBase{
			charmManifestBase{
				Name_:          "ubuntu",
				Channel_:       "22.04",
				Architectures_: []string{"amd64"},
			},
		},
	}
	metadata := newCharmManifest(args)

	c.Assert(metadata.Bases(), gc.DeepEquals, []CharmManifestBase{
		charmManifestBase{
			Name_:          "ubuntu",
			Channel_:       "22.04",
			Architectures_: []string{"amd64"},
		},
	})
}

func minimalCharmManifestMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version": 1,
		"bases":   []interface{}{},
	}
}

func minimalCharmManifestArgs() CharmManifestArgs {
	return CharmManifestArgs{}
}

func minimalCharmManifest() *charmManifest {
	return newCharmManifest(minimalCharmManifestArgs())
}

func maximalCharmManifestMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version": 1,
		"bases": []interface{}{
			map[interface{}]interface{}{
				"name":          "ubuntu",
				"channel":       "22.04",
				"architectures": []interface{}{"amd64"},
			},
		},
	}
}

func maximalCharmManifestArgs() CharmManifestArgs {
	return CharmManifestArgs{
		Bases: []CharmManifestBase{
			charmManifestBase{
				Name_:          "ubuntu",
				Channel_:       "22.04",
				Architectures_: []string{"amd64"},
			},
		},
	}
}

func maximalCharmManifest() *charmManifest {
	return newCharmManifest(maximalCharmManifestArgs())
}

func (s *CharmManifestSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalCharmManifest())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalCharmManifestMap())
}

func (s *CharmManifestSerializationSuite) TestMaximalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(maximalCharmManifest())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, maximalCharmManifestMap())
}

func (s *CharmManifestSerializationSuite) TestMinimalParsingSerializedData(c *gc.C) {
	initial := minimalCharmManifest()
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importCharmManifest(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}

func (s *CharmManifestSerializationSuite) TestMaximalParsingSerializedData(c *gc.C) {
	initial := maximalCharmManifest()
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importCharmManifest(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}

func (s *CharmManifestSerializationSuite) exportImportVersion(c *gc.C, origin_ *charmManifest, version int) *charmManifest {
	origin_.Version_ = version
	bytes, err := yaml.Marshal(origin_)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	origin, err := importCharmManifest(source)
	c.Assert(err, jc.ErrorIsNil)
	return origin
}

func (s *CharmManifestSerializationSuite) TestV1ParsingReturnsLatest(c *gc.C) {
	args := maximalCharmManifestArgs()
	originV1 := newCharmManifest(args)

	originLatest := *originV1
	originResult := s.exportImportVersion(c, originV1, 1)
	c.Assert(*originResult, jc.DeepEquals, originLatest)
}
