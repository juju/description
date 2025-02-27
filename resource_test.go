// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"time"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type ResourceSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&ResourceSuite{})

func (s *ResourceSuite) SetUpTest(c *gc.C) {
	s.SerializationSuite.SetUpTest(c)
	s.importName = "resources"
	s.sliceName = "resources"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importResources(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["resources"] = []interface{}{}
	}
}

func minimalResourceMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"name": "bdist",
		"application-revision": map[interface{}]interface{}{
			"revision":     3,
			"sha384":       "aaaaaaaa",
			"origin":       "store",
			"size":         111,
			"timestamp":    "2016-10-18T02:03:04Z",
			"type":         "file",
			"retrieved-by": "user",
		},
	}
}

func minimalResourcesMapV1() map[string]interface{} {
	return map[string]interface{}{
		"version": 1,
		"resources": []map[string]interface{}{{
			"name": "bdist",
			"application-revision": map[interface{}]interface{}{
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

func minimalResource() *resource {
	r := newResource(ResourceArgs{Name: "bdist"})
	r.SetApplicationRevision(ResourceRevisionArgs{
		Revision:    3,
		Type:        "file",
		Origin:      "store",
		SHA384:      "aaaaaaaa",
		Size:        111,
		Timestamp:   time.Date(2016, 10, 18, 2, 3, 4, 0, time.UTC),
		RetrievedBy: "user",
	})
	return r
}

func (s *ResourceSuite) TestNew(c *gc.C) {
	r := minimalResource()
	c.Check(r.Name(), gc.Equals, "bdist")

	appRev := r.ApplicationRevision()
	c.Check(appRev.Revision(), gc.Equals, 3)
	c.Check(appRev.Type(), gc.Equals, "file")
	c.Check(appRev.Origin(), gc.Equals, "store")
	c.Check(appRev.SHA384(), gc.Equals, "aaaaaaaa")
	c.Check(appRev.Size(), gc.Equals, int64(111))
	c.Check(appRev.Timestamp(), gc.Equals, time.Date(2016, 10, 18, 2, 3, 4, 0, time.UTC))
	c.Check(appRev.RetrievedBy(), gc.Equals, "user")

	c.Assert(r.CharmStoreRevision(), gc.IsNil)

	r.SetCharmStoreRevision(ResourceRevisionArgs{
		Revision: 4,
		Type:     "file",
		Origin:   "store",
		SHA384:   "bbbbbbbb",
		Size:     222,
	})
	csRev := r.CharmStoreRevision()
	c.Check(csRev.Revision(), gc.Equals, 4)
	c.Check(csRev.Type(), gc.Equals, "file")
	c.Check(csRev.Origin(), gc.Equals, "store")
	c.Check(csRev.SHA384(), gc.Equals, "bbbbbbbb")
	c.Check(csRev.Size(), gc.Equals, int64(222))
	c.Check(csRev.Timestamp(), gc.Equals, time.Time{})
	c.Check(csRev.RetrievedBy(), gc.Equals, "")
}

func (s *ResourceSuite) TestNilRevisions(c *gc.C) {
	r := newResource(ResourceArgs{"z"})
	c.Check(r.ApplicationRevision(), gc.IsNil)
	c.Check(r.CharmStoreRevision(), gc.IsNil)
}

func (s *ResourceSuite) TestMinimalValid(c *gc.C) {
	r := minimalResource()
	c.Assert(r.Validate(), jc.ErrorIsNil)
}

func (s *ResourceSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalResource())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalResourceMap())
}

func (s *ResourceSuite) TestValidateEmptyResource(c *gc.C) {
	r := newResource(ResourceArgs{Name: "empty"})
	c.Check(r.Validate(), gc.ErrorMatches, "no application revision set")
}

func (s *ResourceSuite) TestValidateMissingApplicationRev(c *gc.C) {
	r := newResource(ResourceArgs{Name: "bdist"})
	r.SetCharmStoreRevision(ResourceRevisionArgs{
		Revision: 4,
		Type:     "file",
		Origin:   "store",
		SHA384:   "bbbbbbbb",
		Size:     222,
	})
	c.Check(r.Validate(), gc.ErrorMatches, "no application revision set")
}

func (s *ResourceSuite) TestMinimalResourceImportV1(c *gc.C) {
	resourcesOut, err := importResources(minimalResourcesMapV1())
	c.Assert(err, jc.ErrorIsNil)
	r := newResource(ResourceArgs{
		Name: "bdist",
	})
	r.SetApplicationRevision(ResourceRevisionArgs{
		Revision:    3,
		Type:        "file",
		SHA384:      "aaaaaaaa",
		Size:        111,
		RetrievedBy: "user",
		Origin:      "store",
		Timestamp:   time.Date(2016, 10, 18, 2, 3, 4, 0, time.UTC),
	})
	c.Assert(resourcesOut, gc.DeepEquals, []*resource{r})
}

func (s *ResourceSuite) TestRoundTrip(c *gc.C) {
	rIn := minimalResource()
	rOut := s.exportImport(c, rIn)
	c.Assert(rOut, jc.DeepEquals, rIn)
}

func (s *ResourceSuite) exportImport(c *gc.C, resourceIn *resource) *resource {
	resourcesIn := &resources{
		Version:    2,
		Resources_: []*resource{resourceIn},
	}
	bytes, err := yaml.Marshal(resourcesIn)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	resourcesOut, err := importResources(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(resourcesOut, gc.HasLen, 1)
	return resourcesOut[0]
}
