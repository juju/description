// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"time"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type UnitResourceSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&UnitResourceSuite{})

func (s *UnitResourceSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "unit resources"
	s.sliceName = "resources"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importUnitResources(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["resources"] = []interface{}{}
	}
}

func minimalUnitResource() *unitResource {
	return newUnitResource(UnitResourceArgs{
		Name: "blah",
		RevisionArgs: ResourceRevisionArgs{
			Revision:    3,
			Type:        "file",
			Origin:      "store",
			SHA384:      "aaaaaaaa",
			Size:        111,
			Timestamp:   time.Date(2016, 10, 18, 2, 3, 4, 0, time.UTC),
			RetrievedBy: "user",
		},
	})
}

func minimalUnitResourcesMapV1() map[string]interface{} {
	return map[string]interface{}{
		"version": 1,
		"resources": []map[string]interface{}{{
			"name": "blah",
			"revision": map[interface{}]interface{}{
				"revision":    3,
				"description": "description",
				"fingerprint": "aaaaaaaa",
				"origin":      "store",
				"path":        "file.tar.gz",
				"size":        111,
				"timestamp":   "2016-10-18T02:03:04Z",
				"type":        "file",
				"username":    "user",
			},
		}},
	}
}

func (s *UnitResourceSuite) TestNew(c *gc.C) {
	ur := minimalUnitResource()
	c.Check(ur.Name(), gc.Equals, "blah")
	rev := ur.Revision()
	c.Check(rev.Revision(), gc.Equals, 3)
	c.Check(rev.Type(), gc.Equals, "file")
	c.Check(rev.Origin(), gc.Equals, "store")
	c.Check(rev.SHA384(), gc.Equals, "aaaaaaaa")
	c.Check(rev.Size(), gc.Equals, int64(111))
	c.Check(rev.Timestamp(), gc.Equals, time.Date(2016, 10, 18, 2, 3, 4, 0, time.UTC))
	c.Check(rev.RetrievedBy(), gc.Equals, "user")
}

func (s *UnitResourceSuite) TestRoundTrip(c *gc.C) {
	urIn := minimalUnitResource()
	urOut := s.exportImport(c, urIn)
	c.Assert(urOut, jc.DeepEquals, urIn)
}

func (s *UnitResourceSuite) TestImportV1(c *gc.C) {
	urs, err := importUnitResources(minimalUnitResourcesMapV1())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(urs, jc.DeepEquals, []*unitResource{minimalUnitResource()})
}

func (s *UnitResourceSuite) TestImportEmpty(c *gc.C) {
	r, err := importUnitResources(map[string]interface{}{
		"version":   2,
		"resources": []interface{}{},
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(r, gc.HasLen, 0)
}

func (s *UnitResourceSuite) TestUnsupportedVersion(c *gc.C) {
	_, err := importUnitResources(map[string]interface{}{
		"version":   999,
		"resources": []interface{}{},
	})
	c.Assert(err, gc.ErrorMatches, "version 999 not valid")
}

func (s *UnitResourceSuite) exportImport(c *gc.C, ur *unitResource) *unitResource {
	initial := unitResources{
		Version:    2,
		Resources_: []*unitResource{ur},
	}
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	urs, err := importUnitResources(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(urs, gc.HasLen, 1)
	return urs[0]
}
