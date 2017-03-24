// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type SubnetSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&SubnetSerializationSuite{})

func (s *SubnetSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "subnets"
	s.sliceName = "subnets"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importSubnets(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["subnets"] = []interface{}{}
	}
}

func testSubnetMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"cidr":                "10.0.0.0/24",
		"provider-id":         "magic",
		"provider-network-id": "carpet",
		"vlan-tag":            64,
		"space-name":          "foo",
		"availability-zone":   "bar",
		"allocatable-ip-high": "10.0.0.255",
		"allocatable-ip-low":  "10.0.0.0",
	}
}

func testSubnet() *subnet {
	return newSubnet(testSubnetArgs())
}

func testSubnetArgs() SubnetArgs {
	return SubnetArgs{
		CIDR:              "10.0.0.0/24",
		ProviderId:        "magic",
		ProviderNetworkId: "carpet",
		VLANTag:           64,
		SpaceName:         "foo",
		AvailabilityZone:  "bar",
		AllocatableIPHigh: "10.0.0.255",
		AllocatableIPLow:  "10.0.0.0",
	}
}

func (s *SubnetSerializationSuite) TestNewSubnet(c *gc.C) {
	args := testSubnetArgs()
	subnet := newSubnet(args)
	c.Assert(subnet.CIDR(), gc.Equals, args.CIDR)
	c.Assert(subnet.ProviderId(), gc.Equals, args.ProviderId)
	c.Assert(subnet.ProviderNetworkId(), gc.Equals, args.ProviderNetworkId)
	c.Assert(subnet.VLANTag(), gc.Equals, args.VLANTag)
	c.Assert(subnet.SpaceName(), gc.Equals, args.SpaceName)
	c.Assert(subnet.AvailabilityZone(), gc.Equals, args.AvailabilityZone)
	c.Assert(subnet.AllocatableIPHigh(), gc.Equals, args.AllocatableIPHigh)
	c.Assert(subnet.AllocatableIPLow(), gc.Equals, args.AllocatableIPLow)
}

func (s *SubnetSerializationSuite) exportImport(c *gc.C, subnet_ *subnet, version int) *subnet {
	initial := subnets{
		Version:  version,
		Subnets_: []*subnet{subnet_},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	subnets, err := importSubnets(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(subnets, gc.HasLen, 1)
	return subnets[0]
}

func (s *SubnetSerializationSuite) TestParsingV1Full(c *gc.C) {
	original := testSubnet()
	original.ProviderNetworkId_ = ""
	subnet := s.exportImport(c, original, 1)
	c.Assert(subnet, jc.DeepEquals, original)
}

func (s *SubnetSerializationSuite) TestParsingV1Minimal(c *gc.C) {
	original := newSubnet(SubnetArgs{CIDR: "10.0.1.0/24"})
	subnet := s.exportImport(c, original, 1)
	c.Assert(subnet, jc.DeepEquals, original)
}

func (s *SubnetSerializationSuite) TestParsingV1IgnoresProviderNetworkId(c *gc.C) {
	original := testSubnet() // Has non-empty network id.
	subnet := s.exportImport(c, original, 1)
	expected := *original
	expected.ProviderNetworkId_ = ""
	// ProviderNetworkId is ignored by the import because it doesn't exist in v1.
	c.Assert(*subnet, jc.DeepEquals, expected)
}

func (s *SubnetSerializationSuite) TestParsingV2Full(c *gc.C) {
	original := testSubnet()
	subnet := s.exportImport(c, original, 2)
	c.Assert(subnet, jc.DeepEquals, original)
}

func (s *SubnetSerializationSuite) TestParsingV2Minimal(c *gc.C) {
	original := newSubnet(SubnetArgs{CIDR: "10.0.1.0/24"})
	subnet := s.exportImport(c, original, 2)
	c.Assert(subnet, jc.DeepEquals, original)
}
