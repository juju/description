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
		"version":  2,
		"source":   "local",
		"id":       "",
		"hash":     "",
		"revision": 0,
		"channel":  "",
		"platform": "amd64/ubuntu/22.04/stable",
	}
}

func minimalCharmOriginArgs() CharmOriginArgs {
	return CharmOriginArgs{
		Source:   "local",
		Platform: "amd64/ubuntu/22.04/stable",
	}
}

func minimalCharmOrigin() *charmOrigin {
	return newCharmOrigin(minimalCharmOriginArgs())
}

func maximalCharmOriginMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version":  2,
		"source":   "charmhub",
		"id":       "random-id",
		"hash":     "c553eee8dc77f2cce29a1c7090d1e3c81e76c6e12346d09936048ed12305fd35",
		"revision": 0,
		"channel":  "foo/stable",
		"platform": "amd64/ubuntu/22.04/stable",
	}
}

func maximalCharmOriginArgs() CharmOriginArgs {
	return CharmOriginArgs{
		Source:   "charmhub",
		ID:       "random-id",
		Hash:     "c553eee8dc77f2cce29a1c7090d1e3c81e76c6e12346d09936048ed12305fd35",
		Channel:  "foo/stable",
		Platform: "amd64/ubuntu/22.04/stable",
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
	initial := maximalCharmOrigin()
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importCharmOrigin(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}

func (s *CharmOriginSerializationSuite) exportImportVersion(c *gc.C, origin_ *charmOrigin, version int) *charmOrigin {
	origin_.Version_ = version
	bytes, err := yaml.Marshal(origin_)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	origin, err := importCharmOrigin(source)
	c.Assert(err, jc.ErrorIsNil)
	return origin
}

func (s *CharmOriginSerializationSuite) TestV1ParsingReturnsLatest(c *gc.C) {
	args := maximalCharmOriginArgs()
	args.Platform = "amd64/ubuntu/focal"
	originV1 := newCharmOrigin(args)

	originLatest := *originV1
	originLatest.Platform_ = "amd64/ubuntu/20.04/stable"
	originResult := s.exportImportVersion(c, originV1, 1)
	c.Assert(*originResult, jc.DeepEquals, originLatest)
}

func (s *CharmOriginSerializationSuite) TestPlatformFromSeriesErrorsOnEmpty(c *gc.C) {
	rval, err := platformFromSeries("")
	c.Assert(err, gc.NotNil)
	c.Assert(rval, gc.Equals, "")
	c.Assert(err.Error(), gc.Equals, "cannot convert empty series to a platform")
}

func (s *CharmOriginSerializationSuite) TestPlatformFromSeries(c *gc.C) {
	tests := []struct {
		expectedVal string
		series      string
	}{
		{
			expectedVal: "amd64/ubuntu/20.04",
			series:      "focal",
		},
		{
			expectedVal: "amd64/ubuntu/22.04",
			series:      "jammy",
		},
		{
			expectedVal: "amd64/ubuntu/16.04",
			series:      "xenial",
		},
		{
			expectedVal: "amd64/windows/win2008r2",
			series:      "win2008r2",
		},
	}

	for _, test := range tests {
		platform, err := platformFromSeries(test.series)
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(platform, gc.Equals, test.expectedVal)
	}
}
