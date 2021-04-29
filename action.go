// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"time"

	"github.com/juju/errors"
	"github.com/juju/schema"
)

// Action represents an action.
type Action interface {
	Id() string
	Receiver() string
	Name() string
	Operation() string
	Parameters() map[string]interface{}
	Enqueued() time.Time
	Started() time.Time
	Completed() time.Time
	Results() map[string]interface{}
	Status() string
	Message() string
	Logs() []ActionMessage
}

// ActionMessage represents an action log message.
type ActionMessage interface {
	Timestamp() time.Time
	Message() string
}

type actions struct {
	Version  int       `yaml:"version"`
	Actions_ []*action `yaml:"actions"`
}

type actionMessages struct {
	Version   int              `yaml:"version"`
	Messages_ []*actionMessage `yaml:"messages"`
}

type actionMessage struct {
	Timestamp_ time.Time `yaml:"timestamp"`
	Message_   string    `yaml:"message"`
}

// Timestamp implements ActionMessage.
func (m actionMessage) Timestamp() time.Time {
	return m.Timestamp_
}

// Message implements ActionMessage.
func (m actionMessage) Message() string {
	return m.Message_
}

type action struct {
	Id_         string                 `yaml:"id"`
	Receiver_   string                 `yaml:"receiver"`
	Name_       string                 `yaml:"name"`
	Operation_  string                 `yaml:"operation"`
	Parameters_ map[string]interface{} `yaml:"parameters"`
	Enqueued_   time.Time              `yaml:"enqueued"`
	// Can't use omitempty with time.Time, it just doesn't work
	// (nothing is serialised), so use a pointer in the struct.
	Started_   *time.Time             `yaml:"started,omitempty"`
	Completed_ *time.Time             `yaml:"completed,omitempty"`
	Status_    string                 `yaml:"status"`
	Message_   string                 `yaml:"message"`
	Results_   map[string]interface{} `yaml:"results"`
	Logs_      *actionMessages        `yaml:"logs,omitempty"`
}

// Id implements Action.
func (i *action) Id() string {
	return i.Id_
}

// Receiver implements Action.
func (i *action) Receiver() string {
	return i.Receiver_
}

// Name implements Action.
func (i *action) Name() string {
	return i.Name_
}

// Operation implements Action.
func (i *action) Operation() string {
	return i.Operation_
}

// Parameters implements Action.
func (i *action) Parameters() map[string]interface{} {
	return i.Parameters_
}

// Enqueued implements Action.
func (i *action) Enqueued() time.Time {
	return i.Enqueued_
}

// Started implements Action.
func (i *action) Started() time.Time {
	var zero time.Time
	if i.Started_ == nil {
		return zero
	}
	return *i.Started_
}

// Completed implements Action.
func (i *action) Completed() time.Time {
	var zero time.Time
	if i.Completed_ == nil {
		return zero
	}
	return *i.Completed_
}

// Status implements Action.
func (i *action) Status() string {
	return i.Status_
}

// Message implements Action.
func (i *action) Message() string {
	return i.Message_
}

// Results implements Action.
func (i *action) Results() map[string]interface{} {
	return i.Results_
}

// Logs implements Action.
func (i *action) Logs() []ActionMessage {
	var result []ActionMessage
	if i.Logs_ == nil {
		return result
	}
	if count := len(i.Logs_.Messages_); count > 0 {
		result = make([]ActionMessage, count)
		for i, value := range i.Logs_.Messages_ {
			result[i] = value
		}
	}
	return result
}

// ActionArgs is an argument struct used to create a
// new internal action type that supports the Action interface.
type ActionArgs struct {
	Id         string
	Receiver   string
	Name       string
	Operation  string
	Parameters map[string]interface{}
	Enqueued   time.Time
	Started    time.Time
	Completed  time.Time
	Status     string
	Message    string
	Results    map[string]interface{}
	Messages   []ActionMessage
}

func newAction(args ActionArgs) *action {
	action := &action{
		Receiver_:   args.Receiver,
		Name_:       args.Name,
		Operation_:  args.Operation,
		Parameters_: args.Parameters,
		Enqueued_:   args.Enqueued,
		Status_:     args.Status,
		Message_:    args.Message,
		Id_:         args.Id,
		Results_:    args.Results,
	}
	if len(args.Messages) > 0 {
		logs := make([]*actionMessage, len(args.Messages))
		for i, m := range args.Messages {
			logs[i] = &actionMessage{
				Timestamp_: m.Timestamp(),
				Message_:   m.Message(),
			}
		}
		action.setLogs(logs)
	}
	if !args.Started.IsZero() {
		value := args.Started
		action.Started_ = &value
	}
	if !args.Completed.IsZero() {
		value := args.Completed
		action.Completed_ = &value
	}
	return action
}

func (a *action) setLogs(messages []*actionMessage) {
	a.Logs_ = &actionMessages{
		Version:   1,
		Messages_: messages,
	}
}

func importActions(source map[string]interface{}) ([]*action, error) {
	checker := versionedChecker("actions")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "actions version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	sourceList := valid["actions"].([]interface{})
	return importActionList(sourceList, version)
}

func importActionList(sourceList []interface{}, version int) ([]*action, error) {
	getFields, ok := actionFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	result := make([]*action, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for action %d, %T", i, value)
		}
		action, err := importAction(source, version, getFields)
		if err != nil {
			return nil, errors.Annotatef(err, "action %d", i)
		}
		result = append(result, action)
	}
	return result, nil
}

var actionFieldsFuncs = map[int]fieldsFunc{
	1: actionV1Fields,
	2: actionV2Fields,
	3: actionV3Fields,
}

func actionV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"receiver":   schema.String(),
		"name":       schema.String(),
		"parameters": schema.StringMap(schema.Any()),
		"enqueued":   schema.Time(),
		"started":    schema.Time(),
		"completed":  schema.Time(),
		"status":     schema.String(),
		"message":    schema.String(),
		"results":    schema.StringMap(schema.Any()),
		"id":         schema.String(),
	}
	// Some values don't have to be there.
	defaults := schema.Defaults{
		"started":   schema.Omit,
		"completed": schema.Omit,
	}
	return fields, defaults
}

func actionV2Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := actionV1Fields()
	fields["logs"] = schema.StringMap(schema.Any())
	defaults["logs"] = schema.Omit
	return fields, defaults
}

func actionV3Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := actionV2Fields()
	fields["operation"] = schema.String()
	return fields, defaults
}

func importAction(source map[string]interface{}, importVersion int, fieldFunc func() (schema.Fields, schema.Defaults)) (*action, error) {
	fields, defaults := fieldFunc()
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "action v%d schema check failed", importVersion)
	}
	valid := coerced.(map[string]interface{})
	action := &action{
		Id_:         valid["id"].(string),
		Receiver_:   valid["receiver"].(string),
		Name_:       valid["name"].(string),
		Status_:     valid["status"].(string),
		Message_:    valid["message"].(string),
		Parameters_: valid["parameters"].(map[string]interface{}),
		Enqueued_:   valid["enqueued"].(time.Time).UTC(),
		Results_:    valid["results"].(map[string]interface{}),
		Started_:    fieldToTimePtr(valid, "started"),
		Completed_:  fieldToTimePtr(valid, "completed"),
	}

	if importVersion >= 2 {
		if logsMap, ok := valid["logs"]; ok {
			logs, err := importActionLogs(logsMap.(map[string]interface{}))
			if err != nil {
				return nil, errors.Trace(err)
			}
			action.setLogs(logs)
		}
	}

	if importVersion >= 3 {
		action.Operation_ = valid["operation"].(string)
	}

	return action, nil
}

func importActionLogs(source map[string]interface{}) ([]*actionMessage, error) {
	checker := versionedChecker("messages")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "action logs version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := actionLogsDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	sourceList := valid["messages"].([]interface{})
	return importActionLogList(sourceList, importFunc)
}

func importActionLogList(sourceList []interface{}, importFunc actionLogsDeserializationFunc) ([]*actionMessage, error) {
	result := make([]*actionMessage, 0, len(sourceList))

	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for action message %d, %T", i, value)
		}

		offer, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "action message %d", i)
		}
		result = append(result, offer)
	}
	return result, nil
}

type actionLogsDeserializationFunc func(interface{}) (*actionMessage, error)

var actionLogsDeserializationFuncs = map[int]actionLogsDeserializationFunc{
	1: importActionMessageV1,
}

func importActionMessageV1(source interface{}) (*actionMessage, error) {
	fields := schema.Fields{
		"timestamp": schema.Time(),
		"message":   schema.String(),
	}
	checker := schema.FieldMap(fields, nil)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "action message v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	return &actionMessage{
		Timestamp_: valid["timestamp"].(time.Time).UTC(),
		Message_:   valid["message"].(string),
	}, nil
}
