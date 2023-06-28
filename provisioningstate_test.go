// Copyright 2023 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type ProvisioningStateSerializationSuite struct {
	SerializationSuite
}

var _ = gc.Suite(&ProvisioningStateSerializationSuite{})

func (s *ProvisioningStateSerializationSuite) SetUpTest(c *gc.C) {
	s.importName = "provisioning-state"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importProvisioningState(m)
	}
}

func (s *ProvisioningStateSerializationSuite) TestNewProvisioningState(c *gc.C) {
	args := ProvisioningStateArgs{
		Scaling:     true,
		ScaleTarget: 10,
	}
	instance := newProvisioningState(&args)
	c.Assert(instance.Scaling(), jc.IsTrue)
	c.Assert(instance.ScaleTarget(), gc.Equals, 10)
}

func minimalProvisioningStateMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version":      1,
		"scaling":      true,
		"scale-target": 10,
	}
}

func minimalProvisioningStateArgs() *ProvisioningStateArgs {
	return &ProvisioningStateArgs{
		Scaling:     true,
		ScaleTarget: 10,
	}
}

func minimalProvisioningState() *provisioningState {
	return newProvisioningState(minimalProvisioningStateArgs())
}

func maximalProvisioningStateMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version":      1,
		"scaling":      true,
		"scale-target": 10,
	}
}

func maximalProvisioningStateArgs() *ProvisioningStateArgs {
	return &ProvisioningStateArgs{
		Scaling:     true,
		ScaleTarget: 10,
	}
}

func maximalProvisioningState() *provisioningState {
	return newProvisioningState(maximalProvisioningStateArgs())
}

func (s *ProvisioningStateSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalProvisioningState())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalProvisioningStateMap())
}

func (s *ProvisioningStateSerializationSuite) TestMaximalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(maximalProvisioningState())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, maximalProvisioningStateMap())
}

func (s *ProvisioningStateSerializationSuite) TestParsingSerializedData(c *gc.C) {
	initial := maximalProvisioningState()
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importProvisioningState(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}
