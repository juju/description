// Copyright 2020 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type CharmOriginSerializationSuite struct {
	SerializationSuite
}

var _ = gc.Suite(&CharmOriginSerializationSuite{})

func (s *CharmOriginSerializationSuite) SetUpTest(c *gc.C) {
	s.importName = "charmOrigin"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importCharmOrigin(m)
	}
}

func (s *CharmOriginSerializationSuite) TestNewCharmOrigin(c *gc.C) {
	args := CharmOriginArgs{
		Source: "local",
	}
	instance := newCharmOrigin(args)

	c.Assert(instance.Source(), gc.Equals, args.Source)
}

func minimalCharmOriginMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version":  1,
		"source":   "local",
		"id":       "",
		"hash":     "",
		"revision": 0,
		"channel":  "",
	}
}

func minimalCharmOriginArgs() CharmOriginArgs {
	return CharmOriginArgs{
		Source: "local",
	}
}

func minimalCharmOrigin() *charmOrigin {
	return newCharmOrigin(minimalCharmOriginArgs())
}

func maximalCharmOriginMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version":  1,
		"source":   "charmhub",
		"id":       "random-id",
		"hash":     "c553eee8dc77f2cce29a1c7090d1e3c81e76c6e12346d09936048ed12305fd35",
		"revision": 0,
		"channel":  "foo/stable",
	}
}

func maximalCharmOriginArgs() CharmOriginArgs {
	return CharmOriginArgs{
		Source:  "charmhub",
		ID:      "random-id",
		Hash:    "c553eee8dc77f2cce29a1c7090d1e3c81e76c6e12346d09936048ed12305fd35",
		Channel: "foo/stable",
	}
}

func maximalCharmOrigin() *charmOrigin {
	return newCharmOrigin(maximalCharmOriginArgs())
}

func (s *CharmOriginSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalCharmOrigin())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalCharmOriginMap())
}

func (s *CharmOriginSerializationSuite) TestMaximalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(maximalCharmOrigin())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, maximalCharmOriginMap())
}

func (s *CharmOriginSerializationSuite) TestParsingSerializedData(c *gc.C) {
	initial := newCharmOrigin(CharmOriginArgs{
		Source: "local",
	})
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importCharmOrigin(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}
