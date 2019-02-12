// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"time"

	"github.com/juju/errors"
	"github.com/juju/schema"
)

// HasStatus defines the common methods for setting and getting status
// entries for the various entities.
type HasStatus interface {
	Status() Status
	SetStatus(StatusArgs)
}

type HasOperatorStatus interface {
	SetOperatorStatus(StatusArgs)
	OperatorStatus() Status
}

// HasModificationStatus defines the comment methods for setting and getting
// status entries for the various entities that are modified by actions.
// The modification changes, are changes that can alter the machine instance
// and setting the status can then be surfaced to the operator using the status.
// This is different from agent-status or machine-status, where the statuses
// tend to imply how the machine health is during a provisioning cycle or hook
// integration.
// Statuses that are expected: Applied, Error.
type HasModificationStatus interface {
	ModificationStatus() Status
	// SetModificationStatus allows the changing of the modification status, of
	// a type, which is meant to highlight the changing to a machine instance
	// after it's been provisioned.
	SetModificationStatus(StatusArgs)
}

// HasStatusHistory defines the common methods for setting and
// getting historical status entries for the various entities.
type HasStatusHistory interface {
	StatusHistory() []Status
	SetStatusHistory([]StatusArgs)
}

// Status represents an agent, application, or workload status.
type Status interface {
	Value() string
	Message() string
	Data() map[string]interface{}
	Updated() time.Time
	NeverSet() bool
}

// StatusArgs is an argument struct used to set the agent, application, or
// workload status.
type StatusArgs struct {
	Value    string
	Message  string
	Data     map[string]interface{}
	Updated  time.Time
	NeverSet bool
}

func newStatus(args StatusArgs) *status {
	return &status{
		Version: 2,
		StatusPoint_: StatusPoint_{
			Value_:    args.Value,
			Message_:  args.Message,
			Data_:     args.Data,
			Updated_:  args.Updated.UTC(),
			NeverSet_: args.NeverSet,
		},
	}
}

func newStatusHistory() StatusHistory_ {
	return StatusHistory_{
		Version: 2,
	}
}

// StatusPoint_ implements Status, and represents the status
// of an entity at a point in time. Used in the serialization of
// both status and StatusHistory_.
type StatusPoint_ struct {
	Value_    string                 `yaml:"value"`
	Message_  string                 `yaml:"message,omitempty"`
	Data_     map[string]interface{} `yaml:"data,omitempty"`
	Updated_  time.Time              `yaml:"updated"`
	NeverSet_ bool                   `yaml:"neverset"`
}

type status struct {
	Version      int `yaml:"version"`
	StatusPoint_ `yaml:"status"`
}

type StatusHistory_ struct {
	Version int             `yaml:"version"`
	History []*StatusPoint_ `yaml:"history"`
}

// Value implements Status.
func (a *StatusPoint_) Value() string {
	return a.Value_
}

// Message implements Status.
func (a *StatusPoint_) Message() string {
	return a.Message_
}

// Data implements Status.
func (a *StatusPoint_) Data() map[string]interface{} {
	return a.Data_
}

// Updated implements Status.
func (a *StatusPoint_) Updated() time.Time {
	return a.Updated_
}

// NeverSet implements Status.
func (a *StatusPoint_) NeverSet() bool {
	return a.NeverSet_
}

func importStatus(source map[string]interface{}) (*status, error) {
	checker := versionedEmbeddedChecker("status")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotate(err, "status version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))

	getFields, ok := statusFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	source = valid["status"].(map[string]interface{})
	point, err := importStatusAllVersions(schema.FieldMap(getFields()), version, source)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &status{
		Version:      2,
		StatusPoint_: point,
	}, nil
}

func importStatusHistory(history *StatusHistory_, source map[string]interface{}) error {
	checker := versionedChecker("history")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return errors.Annotate(err, "status version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	getFields, ok := statusFieldsFuncs[version]
	if !ok {
		return errors.NotValidf("version %d", version)
	}

	sourceList := valid["history"].([]interface{})
	points, err := importStatusList(sourceList, getFields, version)
	if err != nil {
		return errors.Trace(err)
	}
	history.History = points
	return nil
}

func importStatusList(sourceList []interface{}, getFields statusFieldsFunc, version int) ([]*StatusPoint_, error) {
	result := make([]*StatusPoint_, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for status %d, %T", i, value)
		}
		point, err := importStatusAllVersions(schema.FieldMap(getFields()), version, source)
		if err != nil {
			return nil, errors.Annotatef(err, "status history %d", i)
		}
		result = append(result, &point)
	}
	return result, nil
}

func importModificationStatus(source interface{}) (*status, error) {
	if sourceMap, ok := source.(map[string]interface{}); ok {
		return importStatus(sourceMap)
	}
	return nil, nil
}

type statusFieldsFunc func() (schema.Fields, schema.Defaults)

var statusFieldsFuncs = map[int]statusFieldsFunc{
	1: statusV1Fields,
	2: statusV2Fields,
}

func statusV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"value":   schema.String(),
		"message": schema.String(),
		"data":    schema.StringMap(schema.Any()),
		"updated": schema.Time(),
	}
	// Some values don't have to be there.
	defaults := schema.Defaults{
		"message": "",
		"data":    schema.Omit,
	}

	return fields, defaults
}

func statusV2Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := statusV1Fields()
	fields["neverset"] = schema.Bool()
	defaults["neverset"] = false
	return fields, defaults
}

func importStatusAllVersions(checker schema.Checker, importVersion int, source map[string]interface{}) (StatusPoint_, error) {
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return StatusPoint_{}, errors.Annotatef(err, "status v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	var data map[string]interface{}
	if sourceData, set := valid["data"]; set {
		data = sourceData.(map[string]interface{})
	}
	statusPoint := StatusPoint_{
		Value_:   valid["value"].(string),
		Message_: valid["message"].(string),
		Data_:    data,
		Updated_: valid["updated"].(time.Time),
	}

	if importVersion >= 2 {
		statusPoint.NeverSet_ = valid["neverset"].(bool)
	}
	return statusPoint, nil
}

// StatusHistory implements HasStatusHistory.
func (s *StatusHistory_) StatusHistory() []Status {
	var result []Status
	if count := len(s.History); count > 0 {
		result = make([]Status, count)
		for i, value := range s.History {
			result[i] = value
		}
	}
	return result
}

// SetStatusHistory implements HasStatusHistory.
func (s *StatusHistory_) SetStatusHistory(args []StatusArgs) {
	points := make([]*StatusPoint_, len(args))
	for i, arg := range args {
		points[i] = &StatusPoint_{
			Value_:    arg.Value,
			Message_:  arg.Message,
			Data_:     arg.Data,
			Updated_:  arg.Updated.UTC(),
			NeverSet_: arg.NeverSet,
		}
	}
	s.History = points
}

func addStatusHistorySchema(fields schema.Fields) {
	fields["status-history"] = schema.StringMap(schema.Any())
}

func (s *StatusHistory_) importStatusHistory(valid map[string]interface{}) error {
	return importStatusHistory(s, valid["status-history"].(map[string]interface{}))
}
