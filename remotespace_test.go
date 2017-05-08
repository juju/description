// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type RemoteSpaceSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&RemoteSpaceSerializationSuite{})

func (s *RemoteSpaceSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "remote spaces"
	s.sliceName = "spaces"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importRemoteSpaces(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["spaces"] = []interface{}{}
	}
}

func minimalRemoteSpaceMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"cloud-type":  "gce",
		"name":        "private",
		"provider-id": "juju-space-private",
		"provider-attributes": map[interface{}]interface{}{
			"project": "gothic",
		},
		"subnets": map[interface{}]interface{}{
			"version": 3,
			"subnets": []interface{}{map[interface{}]interface{}{
				"cidr":                "2.3.4.0/24",
				"space-name":          "a-space",
				"vlan-tag":            64,
				"provider-id":         "juju-subnet-1",
				"availability-zones":  []interface{}{"az1", "az2"},
				"provider-space-id":   "juju-space-private",
				"provider-network-id": "network-1",
			}},
		},
	}
}

func minimalRemoteSpace() *remoteSpace {
	space := newRemoteSpace(RemoteSpaceArgs{
		CloudType:  "gce",
		Name:       "private",
		ProviderId: "juju-space-private",
		ProviderAttributes: map[string]interface{}{
			"project": "gothic",
		},
	})
	space.AddSubnet(SubnetArgs{
		CIDR:              "2.3.4.0/24",
		SpaceName:         "a-space",
		VLANTag:           64,
		ProviderId:        "juju-subnet-1",
		AvailabilityZones: []string{"az1", "az2"},
		ProviderSpaceId:   "juju-space-private",
		ProviderNetworkId: "network-1",
	})
	return space
}

func (*RemoteSpaceSerializationSuite) TestNew(c *gc.C) {
	r := minimalRemoteSpace()
	c.Check(r.CloudType(), gc.Equals, "gce")
	c.Check(r.Name(), gc.Equals, "private")
	c.Check(r.ProviderId(), gc.Equals, "juju-space-private")
	c.Check(r.ProviderAttributes(), gc.DeepEquals, map[string]interface{}{
		"project": "gothic",
	})
	c.Check(r.Subnets(), gc.DeepEquals, []Subnet{
		newSubnet(SubnetArgs{
			CIDR:              "2.3.4.0/24",
			SpaceName:         "a-space",
			VLANTag:           64,
			ProviderId:        "juju-subnet-1",
			AvailabilityZones: []string{"az1", "az2"},
			ProviderSpaceId:   "juju-space-private",
			ProviderNetworkId: "network-1",
		}),
	})
}

func (*RemoteSpaceSerializationSuite) TestBadSchema1(c *gc.C) {
	container := map[string]interface{}{
		"version": 1,
		"spaces":  []interface{}{1234},
	}
	_, err := importRemoteSpaces(container)
	c.Assert(err, gc.ErrorMatches, `remote spaces version schema check failed: spaces\[0\]: expected map, got int\(1234\)`)
}

func (*RemoteSpaceSerializationSuite) TestBadSchema2(c *gc.C) {
	m := minimalRemoteSpaceMap()
	m["provider-attributes"] = "blah"
	container := map[string]interface{}{
		"version": 1,
		"spaces":  []interface{}{m},
	}
	_, err := importRemoteSpaces(container)
	c.Assert(err, gc.ErrorMatches, `remote space 0 v1 schema check failed: provider-attributes: expected map, got string\("blah"\)`)
}

func (*RemoteSpaceSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalRemoteSpace())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalRemoteSpaceMap())
}

func (s *RemoteSpaceSerializationSuite) TestRoundTrip(c *gc.C) {
	rIn := minimalRemoteSpace()
	rOut := s.exportImport(c, rIn)
	c.Assert(rOut, jc.DeepEquals, rIn)
}

func (s *RemoteSpaceSerializationSuite) exportImport(c *gc.C, spaceIn *remoteSpace) *remoteSpace {
	spacesIn := &remoteSpaces{
		Version: 1,
		Spaces:  []*remoteSpace{spaceIn},
	}
	bytes, err := yaml.Marshal(spacesIn)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	spacesOut, err := importRemoteSpaces(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(spacesOut, gc.HasLen, 1)
	return spacesOut[0]
}
