// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type RemoteEndpointSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&RemoteEndpointSerializationSuite{})

func (s *RemoteEndpointSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "remote endpoints"
	s.sliceName = "endpoints"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importRemoteEndpoints(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["endpoints"] = []interface{}{}
	}
}

func minimalRemoteEndpointMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"name":      "lana",
		"role":      "provider",
		"interface": "mysql",
	}
}

func minimalRemoteEndpoint() *remoteEndpoint {
	return newRemoteEndpoint(RemoteEndpointArgs{
		Name:      "lana",
		Role:      "provider",
		Interface: "mysql",
	})
}

func (*RemoteEndpointSerializationSuite) TestNew(c *gc.C) {
	r := minimalRemoteEndpoint()
	c.Check(r.Name(), gc.Equals, "lana")
	c.Check(r.Role(), gc.Equals, "provider")
	c.Check(r.Interface(), gc.Equals, "mysql")
}

func (*RemoteEndpointSerializationSuite) TestBadSchema1(c *gc.C) {
	container := map[string]interface{}{
		"version":   1,
		"endpoints": []interface{}{1234},
	}
	_, err := importRemoteEndpoints(container)
	c.Assert(err, gc.ErrorMatches, `remote endpoints version schema check failed: endpoints\[0\]: expected map, got int\(1234\)`)
}

func (*RemoteEndpointSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalRemoteEndpoint())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalRemoteEndpointMap())
}

func (s *RemoteEndpointSerializationSuite) TestRoundTrip(c *gc.C) {
	rIn := minimalRemoteEndpoint()
	rOut := s.exportImport(c, rIn)
	c.Assert(rOut, jc.DeepEquals, rIn)
}

func (s *RemoteEndpointSerializationSuite) exportImport(c *gc.C, endpointIn *remoteEndpoint) *remoteEndpoint {
	endpointsIn := &remoteEndpoints{
		Version:   1,
		Endpoints: []*remoteEndpoint{endpointIn},
	}
	bytes, err := yaml.Marshal(endpointsIn)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	endpointsOut, err := importRemoteEndpoints(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(endpointsOut, gc.HasLen, 1)
	return endpointsOut[0]
}
