// Copyright 2018 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type CloudContainerSerializationSuite struct {
	SerializationSuite
}

var _ = gc.Suite(&CloudContainerSerializationSuite{})

func (s *CloudContainerSerializationSuite) SetUpTest(c *gc.C) {
	s.SerializationSuite.SetUpTest(c)
	s.importName = "cloudContainer"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importCloudContainer(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["provider-id"] = ""
		m["address"] = ""
		m["ports"] = ""
	}
}

func (*CloudContainerSerializationSuite) allArgs() CloudContainerArgs {
	return CloudContainerArgs{
		ProviderId: "some-provider",
		Address:    "10.0.0.1",
		Ports:      []string{"80", "443"},
	}
}

func (s *CloudContainerSerializationSuite) TestAllArgs(c *gc.C) {
	args := s.allArgs()
	container := newCloudContainer(args)

	c.Check(container.ProviderId(), gc.Equals, args.ProviderId)
	c.Check(container.Address(), gc.Equals, args.Address)
	c.Check(container.Ports(), jc.DeepEquals, args.Ports)
}

func (s *CloudContainerSerializationSuite) TestParsingSerializedData(c *gc.C) {
	args := s.allArgs()
	initial := newCloudContainer(args)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	imported, err := importCloudContainer(source)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(imported, jc.DeepEquals, initial)
}
