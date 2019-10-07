// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type RelationSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&RelationSerializationSuite{})

func (s *RelationSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "relations"
	s.sliceName = "relations"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importRelations(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["relations"] = []interface{}{}
	}
}

func (s *RelationSerializationSuite) completeRelation() *relation {
	relation := newRelation(RelationArgs{
		Id:              42,
		Key:             "special",
		Suspended:       true,
		SuspendedReason: "reason",
	})
	relation.SetStatus(minimalStatusArgs())

	endpoint := relation.AddEndpoint(minimalEndpointArgs())
	u1Settings := map[string]interface{}{
		"name": "unit one",
		"key":  42,
	}
	u2Settings := map[string]interface{}{
		"name": "unit two",
		"foo":  "bar",
	}
	endpoint.SetUnitSettings("ubuntu/0", u1Settings)
	endpoint.SetUnitSettings("ubuntu/1", u2Settings)

	return relation
}

func (s *RelationSerializationSuite) TestNewRelation(c *gc.C) {
	relation := newRelation(RelationArgs{
		Id:              42,
		Key:             "special",
		Suspended:       true,
		SuspendedReason: "reason",
	})

	c.Assert(relation.Id(), gc.Equals, 42)
	c.Assert(relation.Key(), gc.Equals, "special")
	c.Assert(relation.Suspended(), jc.IsTrue)
	c.Assert(relation.SuspendedReason(), gc.Equals, "reason")
	c.Assert(relation.Endpoints(), gc.HasLen, 0)
}

func (s *RelationSerializationSuite) TestRelationEndpoints(c *gc.C) {
	relation := s.completeRelation()

	endpoints := relation.Endpoints()
	c.Assert(endpoints, gc.HasLen, 1)

	ep := endpoints[0]
	c.Assert(ep.ApplicationName(), gc.Equals, "ubuntu")
	// Not going to check the exact contents, we expect that there
	// should be two entries.
	c.Assert(ep.Settings("ubuntu/0"), gc.HasLen, 2)
}

func (s *RelationSerializationSuite) TestParsingSerializedData(c *gc.C) {
	initial := relations{
		Version:    3,
		Relations_: []*relation{s.completeRelation()},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	relations, err := importRelations(source)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(relations, jc.DeepEquals, initial.Relations_)
}

func (s *RelationSerializationSuite) TestParsingSerializedDataNoStatus(c *gc.C) {
	initial := relations{
		Version:    3,
		Relations_: []*relation{s.completeRelation()},
	}
	initial.Relations_[0].Status_ = nil

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	relations, err := importRelations(source)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(relations, jc.DeepEquals, initial.Relations_)
}

func (s *RelationSerializationSuite) TestVersion1Works(c *gc.C) {
	initial := relations{
		Version:    3,
		Relations_: []*relation{s.completeRelation()},
	}
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)
	var data map[string]interface{}
	err = yaml.Unmarshal(bytes, &data)
	c.Assert(err, jc.ErrorIsNil)
	data["version"] = 1

	relations, err := importRelations(data)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(relations, gc.HasLen, 1)
	// V1 doesn't have status.
	c.Assert(relations[0].Status(), gc.IsNil)
}

func (s *RelationSerializationSuite) TestVersion2Works(c *gc.C) {
	initial := relations{
		Version:    3,
		Relations_: []*relation{s.completeRelation()},
	}
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)
	var data map[string]interface{}
	err = yaml.Unmarshal(bytes, &data)
	c.Assert(err, jc.ErrorIsNil)
	data["version"] = 2

	relations, err := importRelations(data)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(relations, gc.HasLen, 1)
	// V2 suspended is always false.
	c.Assert(relations[0].Suspended(), jc.IsFalse)
	c.Assert(relations[0].SuspendedReason(), gc.Equals, "")
}
