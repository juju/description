// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"time"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type ActionSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&ActionSerializationSuite{})

func (s *ActionSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "actions"
	s.sliceName = "actions"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importActions(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["actions"] = []interface{}{}
	}
}

func minimalActionMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"id":         "foo",
		"name":       "bam",
		"receiver":   "bar",
		"enqueued":   "2019-01-01T06:06:06Z",
		"started":    "2019-01-02T06:06:06Z",
		"completed":  "2019-01-03T06:06:06Z",
		"message":    "a message",
		"parameters": map[interface{}]interface{}{"bar": "bam", "foo": 3},
		"results":    map[interface{}]interface{}{"the": 3, "thing": "bam"},
		"status":     "happy",
	}
}

func minimalActionMapWithLogs() map[interface{}]interface{} {
	result := minimalActionMap()
	result["logs"] = map[interface{}]interface{}{
		"version": 1,
		"messages": []interface{}{
			map[interface{}]interface{}{
				"timestamp": "2019-01-01T06:06:06Z",
				"message":   "hello",
			},
		},
	}
	return result
}

func minimalAction() *action {
	action := newAction(ActionArgs{
		Id:         "foo",
		Receiver:   "bar",
		Name:       "bam",
		Parameters: map[string]interface{}{"foo": 3, "bar": "bam"},
		Enqueued:   time.Date(2019, 01, 01, 6, 6, 6, 0, time.UTC),
		Started:    time.Date(2019, 01, 02, 6, 6, 6, 0, time.UTC),
		Completed:  time.Date(2019, 01, 03, 6, 6, 6, 0, time.UTC),
		Status:     "happy",
		Message:    "a message",
		Results:    map[string]interface{}{"the": 3, "thing": "bam"},
	})
	action.setLogs([]*actionMessage{
		{
			Timestamp_: time.Date(2019, 01, 01, 6, 6, 6, 0, time.UTC),
			Message_:   "hello",
		},
	})
	return action
}

func (s *ActionSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalAction())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalActionMapWithLogs())
}

func (s *ActionSerializationSuite) TestNewAction(c *gc.C) {
	args := ActionArgs{
		Id:         "foo",
		Receiver:   "bar",
		Name:       "bam",
		Parameters: map[string]interface{}{"foo": 3, "bar": "bam"},
		Enqueued:   time.Now(),
		Started:    time.Now(),
		Completed:  time.Now(),
		Status:     "happy",
		Message:    "a message",
		Results:    map[string]interface{}{"the": 3, "thing": "bam"},
		Messages: []ActionMessage{
			&actionMessage{Timestamp_: time.Now(), Message_: "hello"},
		},
	}
	action := newAction(args)
	c.Check(action.Id(), gc.Equals, args.Id)
	c.Check(action.Receiver(), gc.Equals, args.Receiver)
	c.Check(action.Name(), gc.Equals, args.Name)
	c.Check(action.Parameters(), jc.DeepEquals, args.Parameters)
	c.Check(action.Enqueued(), gc.Equals, args.Enqueued)
	c.Check(action.Started(), gc.Equals, args.Started)
	c.Check(action.Completed(), gc.Equals, args.Completed)
	c.Check(action.Status(), gc.Equals, args.Status)
	c.Check(action.Message(), gc.Equals, args.Message)
	c.Check(action.Results(), jc.DeepEquals, args.Results)
	c.Check(action.Logs(), jc.DeepEquals, args.Messages)
}

func (s *ActionSerializationSuite) exportImportVersion(c *gc.C, action_ *action, version int) *action {
	initial := actions{
		Version:  version,
		Actions_: []*action{action_},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	actions, err := importActions(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(actions, gc.HasLen, 1)
	return actions[0]
}

func (s *ActionSerializationSuite) exportImportLatest(c *gc.C, action_ *action) *action {
	return s.exportImportVersion(c, action_, 2)
}

func (s *ActionSerializationSuite) TestV1ParsingReturnsLatest(c *gc.C) {
	actionV1 := minimalAction()

	// Make an action with fields not in v1 removed.
	actionLatest := minimalAction()
	actionLatest.Logs_ = nil

	actionResult := s.exportImportVersion(c, actionV1, 1)
	c.Assert(actionResult, jc.DeepEquals, actionLatest)
}

func (s *ActionSerializationSuite) TestParsingSerializedData(c *gc.C) {
	action := minimalAction()
	actionResult := s.exportImportLatest(c, action)
	c.Assert(actionResult, jc.DeepEquals, action)
}
