// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type RelationNetworkSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&RelationNetworkSerializationSuite{})

func (s *RelationNetworkSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "relation networks"
	s.sliceName = "relation-networks"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importRelationNetworks(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["relation-networks"] = []interface{}{}
	}
}

func minimalRelationNetworkMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"id":           "rel-netw-id",
		"relation-key": "keys-to-the-city",
		"cidrs": []interface{}{
			"1.2.3.4/24",
			"0.0.0.1",
		},
		"type": "ingress",
	}
}

func minimalRelationNetwork() *relationNetwork {
	c := newRelationNetwork(RelationNetworkArgs{
		ID:          "rel-netw-id",
		RelationKey: "keys-to-the-city",
		CIDRS: []string{
			"1.2.3.4/24",
			"0.0.0.1",
		},
	})
	return c
}

func (*RelationNetworkSerializationSuite) TestNew(c *gc.C) {
	e := minimalRelationNetwork()
	c.Check(e.ID(), gc.Equals, "rel-netw-id")
	c.Check(e.RelationKey(), gc.Equals, "keys-to-the-city")
	c.Check(e.CIDRS(), gc.DeepEquals, []string{
		"1.2.3.4/24",
		"0.0.0.1",
	})
}

func (*RelationNetworkSerializationSuite) TestBadSchema1(c *gc.C) {
	container := map[string]interface{}{
		"version":           1,
		"relation-networks": []interface{}{1234},
	}
	_, err := importRelationNetworks(container)
	c.Assert(err, gc.ErrorMatches, `relation networks version schema check failed: relation-networks\[0\]: expected map, got int\(1234\)`)
}

func (*RelationNetworkSerializationSuite) TestBadSchema2(c *gc.C) {
	m := minimalRelationNetworkMap()
	m["id"] = true
	container := map[string]interface{}{
		"version":           1,
		"relation-networks": []interface{}{m},
	}
	_, err := importRelationNetworks(container)
	c.Assert(err, gc.ErrorMatches, `relation network 0 v1 schema check failed: id: expected string, got bool\(true\)`)
}

func (*RelationNetworkSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalRelationNetworkMap())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalRelationNetworkMap())
}

func (s *RelationNetworkSerializationSuite) TestRoundTrip(c *gc.C) {
	rIn := minimalRelationNetwork()
	rOut := s.exportImport(c, rIn)
	c.Assert(rOut, jc.DeepEquals, rIn)
}

func (s *RelationNetworkSerializationSuite) exportImport(c *gc.C, relationNetworkIn *relationNetwork) *relationNetwork {
	relationNetworksIn := &relationNetworks{
		Version:          1,
		RelationNetworks: []*relationNetwork{relationNetworkIn},
	}
	bytes, err := yaml.Marshal(relationNetworksIn)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	relationNetworksOut, err := importRelationNetworks(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(relationNetworksOut, gc.HasLen, 1)
	return relationNetworksOut[0]
}
