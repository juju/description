// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/names.v3"
	"gopkg.in/yaml.v2"
)

type ExternalControllerInfoSerializationSuite struct {
	SerializationSuite
	externalControllerInfoFields map[string]interface{}
}

var _ = gc.Suite(&ExternalControllerInfoSerializationSuite{})

func (s *ExternalControllerInfoSerializationSuite) SetUpTest(c *gc.C) {
	s.SerializationSuite.SetUpTest(c)
	s.importName = "external controller info"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importExternalControllerInfo(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["external-controller-info"] = map[string]interface{}{}
	}
}

func minimalExternalControllerInfoMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"controller-tag": "ctrl-tag",
		"alias":          "ext-ctrl-alias",
		"addrs": []interface{}{
			"1.2.3.4/24",
			"0.0.0.1",
		},
		"cacert": "magic-cert",
	}
}

func minimalExternalControllerInfo() *externalControllerInfo {
	return newExternalControllerInfo(ExternalControllerInfoArgs{
		ControllerTag: names.NewControllerTag("ctrl-tag"),
		Alias:         "ext-ctrl-alias",
		Addrs: []string{
			"1.2.3.4/24",
			"0.0.0.1",
		},
		CACert: "magic-cert",
	})
}

func (*ExternalControllerInfoSerializationSuite) TestNew(c *gc.C) {
	info := minimalExternalControllerInfo()
	c.Check(info.ControllerTag(), gc.Equals, names.NewControllerTag("ctrl-tag"))
	c.Check(info.Addrs(), gc.DeepEquals, []string{
		"1.2.3.4/24",
		"0.0.0.1",
	})
}

func (s *ExternalControllerInfoSerializationSuite) TestBadSchema1(c *gc.C) {
	m := minimalExternalControllerInfoMap()
	m["controller-tag"] = true
	container := map[string]interface{}{
		"version":                  1,
		"external-controller-info": m,
	}
	_, err := importExternalControllerInfo(container)
	c.Check(err, gc.ErrorMatches, `external controller info version schema check failed: controller-tag: expected string, got bool\(true\)`)
}

func (*ExternalControllerInfoSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalExternalControllerInfoMap())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalExternalControllerInfoMap())
}

func (s *ExternalControllerInfoSerializationSuite) TestRoundTrip(c *gc.C) {
	rIn := minimalExternalControllerInfo()
	rOut := s.exportImport(c, rIn)
	c.Assert(rOut, jc.DeepEquals, rIn)
}

func (s *ExternalControllerInfoSerializationSuite) exportImport(c *gc.C, controllerInfoIn *externalControllerInfo) *externalControllerInfo {
	externalControllerInfoIn := &externalControllerInfo{
		Version:                 1,
		ExternalControllerInfo_: controllerInfoIn.ExternalControllerInfo_,
	}
	bytes, err := yaml.Marshal(externalControllerInfoIn)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	externalControllerInfoOut, err := importExternalControllerInfo(source)
	c.Assert(err, jc.ErrorIsNil)
	return externalControllerInfoOut
}
