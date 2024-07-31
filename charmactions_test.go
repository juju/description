// Copyright 2024 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type CharmActionsSerializationSuite struct {
	SerializationSuite
}

var _ = gc.Suite(&CharmActionsSerializationSuite{})

func (s *CharmActionsSerializationSuite) SetUpTest(c *gc.C) {
	s.importName = "charmActions"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importCharmActions(m)
	}
}

func (s *CharmActionsSerializationSuite) TestNewCharmActions(c *gc.C) {
	args := CharmActionsArgs{
		Actions: map[string]CharmAction{
			"echo": charmAction{
				Description_:    "echo description",
				Parallel_:       true,
				ExecutionGroup_: "group1",
				Parameters_: map[string]interface{}{
					"message": "string",
				},
			},
		},
	}
	metadata := newCharmActions(args)

	c.Assert(metadata.Actions(), gc.DeepEquals, map[string]CharmAction{
		"echo": charmAction{
			Description_:    "echo description",
			Parallel_:       true,
			ExecutionGroup_: "group1",
			Parameters_: map[string]interface{}{
				"message": "string",
			},
		},
	})
}

func minimalCharmActionsMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version": 1,
		"actions": map[interface{}]interface{}{},
	}
}

func minimalCharmActionsArgs() CharmActionsArgs {
	return CharmActionsArgs{}
}

func minimalCharmActions() *charmActions {
	return newCharmActions(minimalCharmActionsArgs())
}

func maximalCharmActionsMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version": 1,
		"actions": map[interface{}]interface{}{
			"echo": map[interface{}]interface{}{
				"description":     "echo description",
				"parallel":        true,
				"execution-group": "group1",
				"parameters": map[interface{}]interface{}{
					"message": "string",
				},
			},
		},
	}
}

func maximalCharmActionsArgs() CharmActionsArgs {
	return CharmActionsArgs{
		Actions: map[string]CharmAction{
			"echo": charmAction{
				Description_:    "echo description",
				Parallel_:       true,
				ExecutionGroup_: "group1",
				Parameters_: map[string]interface{}{
					"message": "string",
				},
			},
		},
	}
}

func maximalCharmActions() *charmActions {
	return newCharmActions(maximalCharmActionsArgs())
}

func (s *CharmActionsSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalCharmActions())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalCharmActionsMap())
}

func (s *CharmActionsSerializationSuite) TestMaximalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(maximalCharmActions())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, maximalCharmActionsMap())
}

func (s *CharmActionsSerializationSuite) TestMinimalParsingSerializedData(c *gc.C) {
	initial := minimalCharmActions()
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importCharmActions(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}

func (s *CharmActionsSerializationSuite) TestMaximalParsingSerializedData(c *gc.C) {
	initial := maximalCharmActions()
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importCharmActions(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}

func (s *CharmActionsSerializationSuite) exportImportVersion(c *gc.C, origin_ *charmActions, version int) *charmActions {
	origin_.Version_ = version
	bytes, err := yaml.Marshal(origin_)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	origin, err := importCharmActions(source)
	c.Assert(err, jc.ErrorIsNil)
	return origin
}

func (s *CharmActionsSerializationSuite) TestV1ParsingReturnsLatest(c *gc.C) {
	args := maximalCharmActionsArgs()
	originV1 := newCharmActions(args)

	originLatest := *originV1
	originResult := s.exportImportVersion(c, originV1, 1)
	c.Assert(*originResult, jc.DeepEquals, originLatest)
}
