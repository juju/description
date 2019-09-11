// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type RemoteEntitySerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&RemoteEntitySerializationSuite{})

func (s *RemoteEntitySerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "remote entities"
	s.sliceName = "remote-entities"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importRemoteEntities(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["remote-entities"] = []interface{}{}
	}
}

func minimalRemoteEntityMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"token":    "rnd-token",
		"macaroon": "macaroon-value",
	}
}

func minimalRemoteEntity() *remoteEntity {
	c := newRemoteEntity(RemoteEntityArgs{
		Token:    "rnd-token",
		Macaroon: "macaroon-value",
	})
	return c
}

func (*RemoteEntitySerializationSuite) TestNew(c *gc.C) {
	e := minimalRemoteEntity()
	c.Check(e.Token(), gc.Equals, "rnd-token")
	c.Check(e.Macaroon(), gc.Equals, "macaroon-value")
}

func (*RemoteEntitySerializationSuite) TestBadSchema1(c *gc.C) {
	container := map[string]interface{}{
		"version":         1,
		"remote-entities": []interface{}{1234},
	}
	_, err := importRemoteEntities(container)
	c.Assert(err, gc.ErrorMatches, `remote entities version schema check failed: remote-entities\[0\]: expected map, got int\(1234\)`)
}

func (*RemoteEntitySerializationSuite) TestBadSchema2(c *gc.C) {
	m := minimalRemoteEntityMap()
	m["token"] = true
	container := map[string]interface{}{
		"version":         1,
		"remote-entities": []interface{}{m},
	}
	_, err := importRemoteEntities(container)
	c.Assert(err, gc.ErrorMatches, `remote entity 0 v1 schema check failed: token: expected string, got bool\(true\)`)
}

func (*RemoteEntitySerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalRemoteEntityMap())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalRemoteEntityMap())
}

func (s *RemoteEntitySerializationSuite) TestRoundTrip(c *gc.C) {
	rIn := minimalRemoteEntity()
	rOut := s.exportImport(c, rIn)
	c.Assert(rOut, jc.DeepEquals, rIn)
}

func (s *RemoteEntitySerializationSuite) exportImport(c *gc.C, remoteEntityIn *remoteEntity) *remoteEntity {
	remoteEntitiesIn := &remoteEntities{
		Version:        1,
		RemoteEntities: []*remoteEntity{remoteEntityIn},
	}
	bytes, err := yaml.Marshal(remoteEntitiesIn)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	remoteEntitiesOut, err := importRemoteEntities(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(remoteEntitiesOut, gc.HasLen, 1)
	return remoteEntitiesOut[0]
}
