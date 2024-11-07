// Copyright 2018 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type CloudServiceSerializationSuite struct {
	SerializationSuite
}

var _ = gc.Suite(&CloudServiceSerializationSuite{})

func (s *CloudServiceSerializationSuite) SetUpTest(c *gc.C) {
	s.SerializationSuite.SetUpTest(c)
	s.importName = "cloudService"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importCloudService(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["provider-id"] = ""
		m["addresses"] = []interface{}{}
	}
}

func (*CloudServiceSerializationSuite) allArgs() *CloudServiceArgs {
	return &CloudServiceArgs{
		ProviderId: "provider-id",
		Addresses: []AddressArgs{
			{Value: "10.0.0.1", Type: "special"},
			{Value: "10.0.0.2", Type: "other"},
		},
	}
}

func (s *CloudServiceSerializationSuite) TestAllArgs(c *gc.C) {
	args := s.allArgs()
	container := newCloudService(args)

	c.Check(container.ProviderId(), gc.Equals, args.ProviderId)
	c.Check(container.Addresses(), jc.DeepEquals, []Address{
		&address{Version: 3, Value_: "10.0.0.1", Type_: "special"},
		&address{Version: 3, Value_: "10.0.0.2", Type_: "other"},
	})
}

func (s *CloudServiceSerializationSuite) TestParsingSerializedData(c *gc.C) {
	args := s.allArgs()
	initial := newCloudService(args)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	imported, err := importCloudService(source)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(imported, jc.DeepEquals, initial)
}

func (s *CloudServiceSerializationSuite) TestParsingMinimalSerializedData(c *gc.C) {
	initial := newCloudService(&CloudServiceArgs{})

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	imported, err := importCloudService(source)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(imported, jc.DeepEquals, initial)
}
