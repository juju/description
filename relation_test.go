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

type EndpointSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&EndpointSerializationSuite{})

func (s *EndpointSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "endpoints"
	s.sliceName = "endpoints"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importEndpoints(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["endpoints"] = []interface{}{}
	}
}

func minimalEndpointMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"application-name":     "ubuntu",
		"name":                 "juju-meta",
		"role":                 "peer",
		"interface":            "something",
		"optional":             true,
		"limit":                1,
		"scope":                "container",
		"unit-settings":        map[interface{}]interface{}{},
		"application-settings": map[interface{}]interface{}{},
	}
}

func minimalEndpoint() *endpoint {
	return newEndpoint(minimalEndpointArgs())
}

func minimalEndpointArgs() EndpointArgs {
	return EndpointArgs{
		ApplicationName: "ubuntu",
		Name:            "juju-meta",
		Role:            "peer",
		Interface:       "something",
		Optional:        true,
		Limit:           1,
		Scope:           "container",
	}
}

func endpointWithSettings() *endpoint {
	endpoint := minimalEndpoint()
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
	appSettings := map[string]interface{}{
		"venusian": "superbug",
	}
	endpoint.SetApplicationSettings(appSettings)
	return endpoint
}

func (s *EndpointSerializationSuite) TestNewEndpoint(c *gc.C) {
	endpoint := endpointWithSettings()

	c.Assert(endpoint.ApplicationName(), gc.Equals, "ubuntu")
	c.Assert(endpoint.Name(), gc.Equals, "juju-meta")
	c.Assert(endpoint.Role(), gc.Equals, "peer")
	c.Assert(endpoint.Interface(), gc.Equals, "something")
	c.Assert(endpoint.Optional(), jc.IsTrue)
	c.Assert(endpoint.Limit(), gc.Equals, 1)
	c.Assert(endpoint.Scope(), gc.Equals, "container")
	c.Assert(endpoint.UnitCount(), gc.Equals, 2)
	c.Assert(endpoint.Settings("ubuntu/0"), jc.DeepEquals, map[string]interface{}{
		"name": "unit one",
		"key":  42,
	})
	c.Assert(endpoint.Settings("ubuntu/1"), jc.DeepEquals, map[string]interface{}{
		"name": "unit two",
		"foo":  "bar",
	})
	c.Assert(endpoint.AllSettings(), jc.DeepEquals, map[string]map[string]interface{}{
		"ubuntu/0": {
			"name": "unit one",
			"key":  42,
		},
		"ubuntu/1": {
			"name": "unit two",
			"foo":  "bar",
		},
	})
	c.Assert(endpoint.ApplicationSettings(), gc.DeepEquals, map[string]interface{}{
		"venusian": "superbug",
	})
}

func (s *EndpointSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalEndpoint())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalEndpointMap())
}

func (s *EndpointSerializationSuite) TestParsingSerializedData(c *gc.C) {
	initial := endpoints{
		Version:    2,
		Endpoints_: []*endpoint{endpointWithSettings()},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	endpoints, err := importEndpoints(source)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(endpoints, jc.DeepEquals, initial.Endpoints_)
}

func (s *EndpointSerializationSuite) TestParsingV1IgnoresAppSettings(c *gc.C) {
	initial := endpoints{
		Version:    2,
		Endpoints_: []*endpoint{endpointWithSettings()},
	}
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)
	var data map[string]interface{}
	err = yaml.Unmarshal(bytes, &data)
	c.Assert(err, jc.ErrorIsNil)
	data["version"] = 1

	endpoints, err := importEndpoints(data)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(endpoints, gc.HasLen, 1)
	// Application settings should have been ignored.
	c.Assert(endpoints[0].ApplicationSettings(), gc.DeepEquals, map[string]interface{}{})
}

func (s *EndpointSerializationSuite) TestParsingV1NoAppSettings(c *gc.C) {
	noAppSettingsMap := minimalEndpointMap()
	delete(noAppSettingsMap, "application-settings")

	data := map[string]interface{}{
		"version":   1,
		"endpoints": []interface{}{noAppSettingsMap},
	}
	endpoints, err := importEndpoints(data)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(endpoints, gc.HasLen, 1)
	// No error importing, app settings empty.
	c.Assert(endpoints[0].ApplicationSettings(), gc.DeepEquals, map[string]interface{}{})
}
