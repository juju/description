// Copyright 2020 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"time"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type OperationSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&OperationSerializationSuite{})

func (s *OperationSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "operations"
	s.sliceName = "operations"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importOperations(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["operations"] = []interface{}{}
	}
}

func minimalOperationMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"id":                  "foo",
		"summary":             "bam",
		"enqueued":            "2019-01-01T06:06:06Z",
		"started":             "2019-01-02T06:06:06Z",
		"completed":           "2019-01-03T06:06:06Z",
		"complete-task-count": 666,
		"status":              "happy",
	}
}

func minimalOperation() *operation {
	return newOperation(OperationArgs{
		Id:                "foo",
		Summary:           "bam",
		Enqueued:          time.Date(2019, 01, 01, 6, 6, 6, 0, time.UTC),
		Started:           time.Date(2019, 01, 02, 6, 6, 6, 0, time.UTC),
		Completed:         time.Date(2019, 01, 03, 6, 6, 6, 0, time.UTC),
		Status:            "happy",
		CompleteTaskCount: 666,
	})
}

func (s *OperationSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalOperation())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalOperationMap())
}

func (s *OperationSerializationSuite) TestNewOperation(c *gc.C) {
	args := OperationArgs{
		Id:                "foo",
		Summary:           "bam",
		Enqueued:          time.Now(),
		Started:           time.Now(),
		Completed:         time.Now(),
		Status:            "happy",
		CompleteTaskCount: 666,
	}
	operation := newOperation(args)
	c.Check(operation.Id(), gc.Equals, args.Id)
	c.Check(operation.Summary(), gc.Equals, args.Summary)
	c.Check(operation.Enqueued(), gc.Equals, args.Enqueued)
	c.Check(operation.Started(), gc.Equals, args.Started)
	c.Check(operation.Completed(), gc.Equals, args.Completed)
	c.Check(operation.Status(), gc.Equals, args.Status)
	c.Check(operation.CompleteTaskCount(), gc.Equals, args.CompleteTaskCount)
}

func (s *OperationSerializationSuite) exportImportVersion(c *gc.C, operation_ *operation, version int) *operation {
	initial := operations{
		Version:     version,
		Operations_: []*operation{operation_},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	operations, err := importOperations(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(operations, gc.HasLen, 1)
	return operations[0]
}

func (s *OperationSerializationSuite) exportImportLatest(c *gc.C, operation_ *operation) *operation {
	return s.exportImportVersion(c, operation_, 1)
}

func (s *OperationSerializationSuite) TestParsingSerializedData(c *gc.C) {
	operation := minimalOperation()
	operationResult := s.exportImportLatest(c, operation)
	c.Assert(operationResult, jc.DeepEquals, operation)
}
