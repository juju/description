// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/names.v3"
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
		"id": "ext-ctrl",
		"controller-info": map[interface{}]interface{}{
			"version": 1,
			"controller-info": map[interface{}]interface{}{
				"controller-tag": "ctrl-tag",
				"alias":          "ext-ctrl-alias",
				"addrs": []interface{}{
					"1.2.3.4/24",
					"0.0.0.1",
				},
				"cacert": "magic-cert",
			},
		},
	}
}

func minimalExternalController() *externalController {
	c := newExternalController(ExternalControllerArgs{
		Tag: names.NewControllerTag("ext-ctrl"),
	})
	c.AddControllerInfo(ExternalControllerInfoArgs{
		ControllerTag: names.NewControllerTag("ctrl-tag"),
		Alias:         "ext-ctrl-alias",
		Addrs: []string{
			"1.2.3.4/24",
			"0.0.0.1",
		},
		CACert: "magic-cert",
	})
	return c
}

func (*ExternalControllerSerializationSuite) TestNew(c *gc.C) {
	e := minimalExternalController()
	c.Check(e.ID(), gc.Equals, names.NewControllerTag("ext-ctrl"))

	info := e.ControllerInfo()
	c.Check(info.ControllerTag(), gc.Equals, names.NewControllerTag("ctrl-tag"))
	c.Check(info.Addrs(), gc.DeepEquals, []string{
		"1.2.3.4/24",
		"0.0.0.1",
	})
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
	c.Assert(err, gc.ErrorMatches, `external controller 0 v1 schema check failed: id: expected string, got bool\(true\)`)
}

func (s *ExternalControllerSerializationSuite) TestBadExternalControllerInfo(c *gc.C) {
	m := minimalExternalControllerMap()
	m["controllers"] = map[interface{}]interface{}{
		"version": 1,
		"bishop":  "otter-trouserpress",
	}
	container := map[string]interface{}{
		"version":              1,
		"external-controllers": []interface{}{m},
	}
	_, err := importExternalControllers(container)
	c.Assert(err, gc.ErrorMatches, `external controller 0: external controller info version schema check failed: external-controller-info: expected map, got nothing`)
}

func (*ExternalControllerSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalExternalControllerMap())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalExternalControllerMap())
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
