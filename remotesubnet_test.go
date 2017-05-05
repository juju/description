// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type RemoteSubnetSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&RemoteSubnetSerializationSuite{})

func (s *RemoteSubnetSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "remote subnets"
	s.sliceName = "subnets"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importRemoteSubnets(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["subnets"] = []interface{}{}
	}
}

func minimalRemoteSubnetMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"cidr":                "2.3.4.0/24",
		"provider-id":         "juju-subnet-1",
		"vlan-tag":            23,
		"availability-zones":  []interface{}{"az1", "az2"},
		"provider-space-id":   "juju-space-private",
		"provider-network-id": "network-1",
	}
}

func minimalRemoteSubnet() *remoteSubnet {
	return newRemoteSubnet(RemoteSubnetArgs{
		CIDR:              "2.3.4.0/24",
		ProviderId:        "juju-subnet-1",
		VLANTag:           23,
		AvailabilityZones: []string{"az1", "az2"},
		ProviderSpaceId:   "juju-space-private",
		ProviderNetworkId: "network-1",
	})
}

func (*RemoteSubnetSerializationSuite) TestNew(c *gc.C) {
	r := minimalRemoteSubnet()
	c.Check(r.CIDR(), gc.Equals, "2.3.4.0/24")
	c.Check(r.ProviderId(), gc.Equals, "juju-subnet-1")
	c.Check(r.VLANTag(), gc.Equals, 23)
	c.Check(r.AvailabilityZones(), gc.DeepEquals, []string{"az1", "az2"})
	c.Check(r.ProviderSpaceId(), gc.Equals, "juju-space-private")
	c.Check(r.ProviderNetworkId(), gc.Equals, "network-1")
}

func (*RemoteSubnetSerializationSuite) TestBadSchema1(c *gc.C) {
	container := map[string]interface{}{
		"version": 1,
		"subnets": []interface{}{1234},
	}
	_, err := importRemoteSubnets(container)
	c.Assert(err, gc.ErrorMatches, `remote subnets version schema check failed: subnets\[0\]: expected map, got int\(1234\)`)
}

func (*RemoteSubnetSerializationSuite) TestBadSchema2(c *gc.C) {
	m := minimalRemoteSubnetMap()
	m["vlan-tag"] = "blah"
	container := map[string]interface{}{
		"version": 1,
		"subnets": []interface{}{m},
	}
	_, err := importRemoteSubnets(container)
	c.Assert(err, gc.ErrorMatches, `remote subnet 0 v1 schema check failed: vlan-tag: expected int, got string\("blah"\)`)
}

func (*RemoteSubnetSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalRemoteSubnet())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalRemoteSubnetMap())
}

func (s *RemoteSubnetSerializationSuite) TestRoundTrip(c *gc.C) {
	rIn := minimalRemoteSubnet()
	rOut := s.exportImport(c, rIn)
	c.Assert(rOut, jc.DeepEquals, rIn)
}

func (s *RemoteSubnetSerializationSuite) exportImport(c *gc.C, subnetIn *remoteSubnet) *remoteSubnet {
	subnetsIn := &remoteSubnets{
		Version: 1,
		Subnets: []*remoteSubnet{subnetIn},
	}
	bytes, err := yaml.Marshal(subnetsIn)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	subnetsOut, err := importRemoteSubnets(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(subnetsOut, gc.HasLen, 1)
	return subnetsOut[0]
}
