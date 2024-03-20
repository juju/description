// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type StorageDirectiveSerializationSuite struct {
	SerializationSuite
}

var _ = gc.Suite(&StorageDirectiveSerializationSuite{})

func (s *StorageDirectiveSerializationSuite) SetUpTest(c *gc.C) {
	s.SerializationSuite.SetUpTest(c)
	s.importName = "storage directive"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importStorageDirective(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["pool"] = ""
		m["size"] = 0
		m["count"] = 0
	}
}

func (s *StorageDirectiveSerializationSuite) TestMissingValue(c *gc.C) {
	testMap := s.makeMap(1)
	delete(testMap, "pool")
	_, err := importStorageDirective(testMap)
	c.Check(err.Error(), gc.Equals, "storage directive v1 schema check failed: pool: expected string, got nothing")
}

func (*StorageDirectiveSerializationSuite) TestParsing(c *gc.C) {
	addr, err := importStorageDirective(map[string]interface{}{
		"version": 1,
		"pool":    "olympic",
		"size":    50,
		"count":   2,
	})
	c.Assert(err, jc.ErrorIsNil)
	expected := &storageDirective{
		Version: 1,
		Pool_:   "olympic",
		Size_:   50,
		Count_:  2,
	}
	c.Assert(addr, jc.DeepEquals, expected)
}

func (*StorageDirectiveSerializationSuite) TestParsingSerializedData(c *gc.C) {
	initial := &storageDirective{
		Version: 1,
		Pool_:   "olympic",
		Size_:   50,
		Count_:  2,
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	directives, err := importStorageDirective(source)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(directives, jc.DeepEquals, initial)
}
