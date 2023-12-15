// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/names/v5"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type ExternalControllerSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&ExternalControllerSerializationSuite{})

func (s *ExternalControllerSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "external controllers"
	s.sliceName = "external-controllers"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importExternalControllers(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["external-controllers"] = []interface{}{}
	}
}

func minimalExternalControllerMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"id":    "ext-ctrl",
		"alias": "ext-ctrl-alias",
		"addrs": []interface{}{
			"1.2.3.4/24",
			"0.0.0.1",
		},
		"ca-cert": "magic-cert",
		"models": []interface{}{
			"aaaa-bbbb",
		},
	}
}

func minimalExternalController() *externalController {
	c := newExternalController(ExternalControllerArgs{
		Tag:   names.NewControllerTag("ext-ctrl"),
		Alias: "ext-ctrl-alias",
		Addrs: []string{
			"1.2.3.4/24",
			"0.0.0.1",
		},
		CACert: "magic-cert",
		Models: []string{
			"aaaa-bbbb",
		},
	})
	return c
}

func (*ExternalControllerSerializationSuite) TestNew(c *gc.C) {
	e := minimalExternalController()
	c.Check(e.ID(), gc.Equals, names.NewControllerTag("ext-ctrl"))
	c.Check(e.Addrs(), gc.DeepEquals, []string{
		"1.2.3.4/24",
		"0.0.0.1",
	})
	c.Check(e.Alias(), gc.Equals, "ext-ctrl-alias")
	c.Check(e.CACert(), gc.Equals, "magic-cert")
	c.Check(e.Models(), gc.DeepEquals, []string{"aaaa-bbbb"})
}

func (*ExternalControllerSerializationSuite) TestBadSchema1(c *gc.C) {
	container := map[string]interface{}{
		"version":              1,
		"external-controllers": []interface{}{1234},
	}
	_, err := importExternalControllers(container)
	c.Assert(err, gc.ErrorMatches, `external controllers version schema check failed: external-controllers\[0\]: expected map, got int\(1234\)`)
}

func (*ExternalControllerSerializationSuite) TestBadSchema2(c *gc.C) {
	m := minimalExternalControllerMap()
	m["id"] = true
	container := map[string]interface{}{
		"version":              1,
		"external-controllers": []interface{}{m},
	}
	_, err := importExternalControllers(container)
	c.Assert(err, gc.ErrorMatches, `.*external controller v1 schema check failed: id: expected string, got bool\(true\)`)
}

func (*ExternalControllerSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalExternalControllerMap())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalExternalControllerMap())
}

func (*ExternalControllerSerializationSuite) TestMinimalMatchesWithoutAlias(c *gc.C) {
	m := minimalExternalControllerMap()
	delete(m, "alias")

	bytes, err := yaml.Marshal(m)
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, m)
}

func (s *ExternalControllerSerializationSuite) TestRoundTrip(c *gc.C) {
	rIn := minimalExternalController()
	rOut := s.exportImport(c, rIn)
	c.Assert(rOut, jc.DeepEquals, rIn)
}

func (s *ExternalControllerSerializationSuite) exportImport(c *gc.C, controllerIn *externalController) *externalController {
	externalControllersIn := &externalControllers{
		Version:             1,
		ExternalControllers: []*externalController{controllerIn},
	}
	bytes, err := yaml.Marshal(externalControllersIn)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	externalControllersOut, err := importExternalControllers(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(externalControllersOut, gc.HasLen, 1)
	return externalControllersOut[0]
}
