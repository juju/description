// Copyright 2020 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"time"

	"github.com/juju/errors"
	"github.com/juju/schema"
)

// Operation represents an operation.
type Operation interface {
	Id() string
	Summary() string
	Fail() string
	Enqueued() time.Time
	Started() time.Time
	Completed() time.Time
	Status() string
	CompleteTaskCount() int
	SpawnedTaskCount() int
}

type operations struct {
	Version     int          `yaml:"version"`
	Operations_ []*operation `yaml:"operations"`
}

type operation struct {
	Id_       string    `yaml:"id"`
	Summary_  string    `yaml:"summary"`
	Enqueued_ time.Time `yaml:"enqueued"`
	// Can't use omitempty with time.Time, it just doesn't work
	// (nothing is serialised), so use a pointer in the struct.
	Started_           *time.Time `yaml:"started,omitempty"`
	Completed_         *time.Time `yaml:"completed,omitempty"`
	Status_            string     `yaml:"status"`
	Fail_              string     `yaml:"fail,omitempty"`
	CompleteTaskCount_ int        `yaml:"complete-task-count"`
	SpawnedTaskCount_  int        `yaml:"spawned-task-count"`
}

// Id implements Operation.
func (i *operation) Id() string {
	return i.Id_
}

// Summary implements Operation.
func (i *operation) Summary() string {
	return i.Summary_
}

// Fail implements Operation.
func (i *operation) Fail() string {
	return i.Fail_
}

// Enqueued implements Operation.
func (i *operation) Enqueued() time.Time {
	return i.Enqueued_
}

// Started implements Operation.
func (i *operation) Started() time.Time {
	var zero time.Time
	if i.Started_ == nil {
		return zero
	}
	return *i.Started_
}

// Completed implements Operation.
func (i *operation) Completed() time.Time {
	var zero time.Time
	if i.Completed_ == nil {
		return zero
	}
	return *i.Completed_
}

// Status implements Operation.
func (i *operation) Status() string {
	return i.Status_
}

// CompleteTaskCount implements Operation.
func (i *operation) CompleteTaskCount() int {
	return i.CompleteTaskCount_
}

// SpawnedTaskCount implements Operation.
func (i *operation) SpawnedTaskCount() int {
	return i.SpawnedTaskCount_
}

// OperationArgs is an argument struct used to create a
// new internal operation type that supports the Operation interface.
type OperationArgs struct {
	Id                string
	Summary           string
	Enqueued          time.Time
	Started           time.Time
	Completed         time.Time
	Status            string
	Fail              string
	CompleteTaskCount int
	SpawnedTaskCount  int
}

func newOperation(args OperationArgs) *operation {
	operation := &operation{
		Id_:                args.Id,
		Summary_:           args.Summary,
		Enqueued_:          args.Enqueued,
		Status_:            args.Status,
		Fail_:              args.Fail,
		CompleteTaskCount_: args.CompleteTaskCount,
		SpawnedTaskCount_:  args.SpawnedTaskCount,
	}
	if !args.Started.IsZero() {
		value := args.Started
		operation.Started_ = &value
	}
	if !args.Completed.IsZero() {
		value := args.Completed
		operation.Completed_ = &value
	}
	return operation
}

func importOperations(source map[string]interface{}) ([]*operation, error) {
	checker := versionedChecker("operations")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "operations version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	sourceList := valid["operations"].([]interface{})
	return importOperationList(sourceList, version)
}

func importOperationList(sourceList []interface{}, version int) ([]*operation, error) {
	getFields, ok := operationFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	result := make([]*operation, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for operation %d, %T", i, value)
		}
		operation, err := importOperation(source, version, getFields)
		if err != nil {
			return nil, errors.Annotatef(err, "operation %d", i)
		}
		result = append(result, operation)
	}
	return result, nil
}

var operationFieldsFuncs = map[int]fieldsFunc{
	1: operationV1Fields,
	2: operationV2Fields,
	3: operationV3Fields,
}

func operationV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"id":                  schema.String(),
		"summary":             schema.String(),
		"enqueued":            schema.Time(),
		"started":             schema.Time(),
		"completed":           schema.Time(),
		"status":              schema.String(),
		"complete-task-count": schema.Int(),
	}
	// Some values don't have to be there.
	defaults := schema.Defaults{
		"started":   schema.Omit,
		"completed": schema.Omit,
	}
	return fields, defaults
}

func operationV2Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := operationV1Fields()
	fields["fail"] = schema.String()
	defaults["fail"] = schema.Omit
	return fields, defaults
}

func operationV3Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := operationV2Fields()
	fields["spawned-task-count"] = schema.Int()
	return fields, defaults
}

func importOperation(source map[string]interface{}, importVersion int, fieldFunc func() (schema.Fields, schema.Defaults)) (*operation, error) {
	fields, defaults := fieldFunc()
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "operation v%d schema check failed", importVersion)
	}
	valid := coerced.(map[string]interface{})
	operation := &operation{
		Id_:                valid["id"].(string),
		Summary_:           valid["summary"].(string),
		Status_:            valid["status"].(string),
		Enqueued_:          valid["enqueued"].(time.Time).UTC(),
		Started_:           fieldToTimePtr(valid, "started"),
		Completed_:         fieldToTimePtr(valid, "completed"),
		CompleteTaskCount_: int(valid["complete-task-count"].(int64)),
	}

	if importVersion >= 2 {
		operation.Fail_, _ = valid["fail"].(string)
	}

	if importVersion >= 3 {
		operation.SpawnedTaskCount_ = int(valid["spawned-task-count"].(int64))
	}

	return operation, nil
}
