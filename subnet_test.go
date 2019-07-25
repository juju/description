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

func testSubnet(version int) *subnet {
	args := testSubnetArgs()
	switch version {
	case 1:
		args.ProviderNetworkId = ""
		fallthrough
	case 2:
		args.AvailabilityZones = []string{"bar"}
		args.ProviderSpaceId = ""
		fallthrough
	case 3:
		args.FanLocalUnderlay = ""
		args.FanOverlay = ""
		fallthrough
	case 4:
		args.SpaceID = ""
		args.IsPublic = false
	case 5:
		args.SpaceName = ""
	}
	return newSubnet(args)
}

func testSubnetArgs() SubnetArgs {
	return SubnetArgs{
		CIDR:              "10.0.0.0/24",
		ProviderId:        "magic",
		ProviderNetworkId: "carpet",
		ProviderSpaceId:   "ride",
		VLANTag:           64,
		SpaceName:         "foo",
		SpaceID:           "7",
		IsPublic:          true,
		FanLocalUnderlay:  "1.2.3.4/24",
		FanOverlay:        "253.0.0.0/8",
		AvailabilityZones: []string{"bar", "baz"},
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
	c.Assert(subnet.ProviderSpaceId(), gc.Equals, args.ProviderSpaceId)
	c.Assert(subnet.VLANTag(), gc.Equals, args.VLANTag)
	c.Assert(subnet.SpaceName(), gc.Equals, args.SpaceName)
	c.Assert(subnet.SpaceID(), gc.Equals, args.SpaceID)
	c.Assert(subnet.IsPublic(), jc.IsTrue)
	c.Assert(subnet.AvailabilityZones(), gc.DeepEquals, args.AvailabilityZones)
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

	c.Logf("source = %#v", source)

	subnets, err := importSubnets(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(subnets, gc.HasLen, 1)
	return subnets[0]
}

func (s *SubnetSerializationSuite) importSubnet(c *gc.C, source map[string]interface{}, version int) *subnet {
	subnets, err := importSubnets(s.wrapWithVersion(source, version))
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(subnets, gc.HasLen, 1)
	return subnets[0]
}

func (s *SubnetSerializationSuite) wrapWithVersion(data map[string]interface{}, version int) map[string]interface{} {
	return map[string]interface{}{
		"version": version,
		"subnets": []interface{}{data},
	}
}

func testSubnetV1Map() map[string]interface{} {
	return map[string]interface{}{
		"cidr":                "10.0.0.0/24",
		"provider-id":         "magic",
		"vlan-tag":            64,
		"space-name":          "foo",
		"availability-zone":   "bar",
		"allocatable-ip-high": "10.0.0.255",
		"allocatable-ip-low":  "10.0.0.0",
	}
}

func (s *SubnetSerializationSuite) TestParsingV1Full(c *gc.C) {
	original := testSubnetV1Map()
	imported := s.importSubnet(c, original, 1)
	expected := testSubnet(1)
	c.Assert(imported, jc.DeepEquals, expected)
}

func (s *SubnetSerializationSuite) TestParsingV1Minimal(c *gc.C) {
	original := map[string]interface{}{
		"cidr":              "10.0.1.0/24",
		"vlan-tag":          1,
		"space-name":        "aspace",
		"availability-zone": "somewhere",
	}
	imported := s.importSubnet(c, original, 1)
	expected := newSubnet(SubnetArgs{
		CIDR:              "10.0.1.0/24",
		VLANTag:           1,
		SpaceName:         "aspace",
		AvailabilityZones: []string{"somewhere"},
	})
	c.Assert(imported, jc.DeepEquals, expected)
}

func (s *SubnetSerializationSuite) TestParsingV1IgnoresProviderNetworkId(c *gc.C) {
	original := testSubnetV1Map()
	original["provider-network-id"] = "something"
	subnet := s.importSubnet(c, original, 1)
	// ProviderNetworkId is ignored by the import because it doesn't exist in v1.
	c.Assert(subnet.ProviderNetworkId_, gc.Equals, "")
}

func testSubnetV2Map() map[string]interface{} {
	return map[string]interface{}{
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

func (s *SubnetSerializationSuite) TestParsingV2Full(c *gc.C) {
	original := testSubnetV2Map()
	imported := s.importSubnet(c, original, 2)
	expected := testSubnet(2)
	c.Assert(imported, jc.DeepEquals, expected)
}

func (s *SubnetSerializationSuite) TestParsingV2Minimal(c *gc.C) {
	original := map[string]interface{}{
		"cidr":              "10.0.1.0/24",
		"vlan-tag":          1,
		"space-name":        "aspace",
		"availability-zone": "somewhere",
	}
	imported := s.importSubnet(c, original, 2)
	expected := newSubnet(SubnetArgs{
		CIDR:              "10.0.1.0/24",
		VLANTag:           1,
		SpaceName:         "aspace",
		AvailabilityZones: []string{"somewhere"},
	})
	c.Assert(imported, jc.DeepEquals, expected)
}

func (s *SubnetSerializationSuite) TestParsingV2IgnoresNewFields(c *gc.C) {
	original := testSubnetV2Map()
	original["provider-space-id"] = "something"
	original["availability-zones"] = []string{"not", "loaded"}
	original["fan-local-underlay"] = "somethingelse"
	original["fan-overlay"] = "also ignored"
	subnet := s.importSubnet(c, original, 2)
	// The new fields are ignored by the import because they don't exist in v2.
	c.Assert(subnet.ProviderSpaceId_, gc.Equals, "")
	c.Assert(subnet.AvailabilityZones_, gc.DeepEquals, []string{"bar"})
}

func testSubnetV3Map() map[string]interface{} {
	return map[string]interface{}{
		"cidr":                "10.0.0.0/24",
		"provider-id":         "magic",
		"provider-network-id": "carpet",
		"provider-space-id":   "ride",
		"vlan-tag":            64,
		"space-name":          "foo",
		"availability-zones":  []interface{}{"bar", "baz"},
		"allocatable-ip-high": "10.0.0.255",
		"allocatable-ip-low":  "10.0.0.0",
	}
}

func (s *SubnetSerializationSuite) TestParsingV3Full(c *gc.C) {
	original := testSubnetV3Map()
	imported := s.importSubnet(c, original, 3)
	expected := testSubnet(3)
	c.Assert(imported, jc.DeepEquals, expected)
}

func (s *SubnetSerializationSuite) TestParsingV3Minimal(c *gc.C) {
	original := newSubnet(SubnetArgs{CIDR: "10.0.1.0/24"})
	subnet := s.exportImport(c, original, 3)
	c.Assert(subnet, jc.DeepEquals, original)
}

func (s *SubnetSerializationSuite) TestParsingV3IgnoresNewFields(c *gc.C) {
	original := testSubnetV3Map()
	original["fan-local-underlay"] = "somethingelse"
	original["fan-overlay"] = "also ignored"
	subnet := s.importSubnet(c, original, 3)
	// The new fields are ignored by the import because they don't exist in v3.
	c.Assert(subnet.FanLocalUnderlay_, gc.Equals, "")
	c.Assert(subnet.FanOverlay_, gc.Equals, "")
}

func (s *SubnetSerializationSuite) TestParsingV4Full(c *gc.C) {
	original := testSubnet(4)
	subnet := s.exportImport(c, original, 4)
	c.Assert(subnet, jc.DeepEquals, original)
}

func (s *SubnetSerializationSuite) TestParsingV4Minimal(c *gc.C) {
	original := newSubnet(SubnetArgs{CIDR: "10.0.1.0/24"})
	subnet := s.exportImport(c, original, 4)
	c.Assert(subnet, jc.DeepEquals, original)
}

func (s *SubnetSerializationSuite) TestParsingV4IgnoresNewFields(c *gc.C) {
	original := testSubnetV3Map()
	original["fan-local-underlay"] = "1.2.3.4/24"
	original["fan-overlay"] = "253.0.0.0/8"
	original["space-id"] = ""
	original["is-public"] = false
	subnet := s.importSubnet(c, original, 4)
	// The new fields are ignored by the import because they don't exist in v4.
	c.Assert(subnet.SpaceID_, gc.Equals, "")
	c.Assert(subnet.IsPublic_, jc.IsFalse)
}

func (s *SubnetSerializationSuite) TestParsingV5Full(c *gc.C) {
	original := testSubnet(5)
	subnet := s.exportImport(c, original, 5)
	c.Assert(subnet, jc.DeepEquals, original)
}

func (s *SubnetSerializationSuite) TestParsingV5Minimal(c *gc.C) {
	original := newSubnet(SubnetArgs{CIDR: "10.0.1.0/24"})
	subnet := s.exportImport(c, original, 5)
	c.Assert(subnet, jc.DeepEquals, original)
}
