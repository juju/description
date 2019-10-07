// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// Relation represents a relationship between two applications,
// or a peer relation between different instances of an application.
type Relation interface {
	HasStatus

	Id() int
	Key() string
	Suspended() bool
	SuspendedReason() string

	Endpoints() []Endpoint
	AddEndpoint(EndpointArgs) Endpoint
}

type relations struct {
	Version    int         `yaml:"version"`
	Relations_ []*relation `yaml:"relations"`
}

type relation struct {
	Id_              int        `yaml:"id"`
	Key_             string     `yaml:"key"`
	Endpoints_       *endpoints `yaml:"endpoints"`
	Suspended_       bool       `yaml:"suspended"`
	SuspendedReason_ string     `yaml:"suspended-reason"`
	Status_          *status    `yaml:"status,omitempty"`
}

// RelationArgs is an argument struct used to specify a relation.
type RelationArgs struct {
	Id              int
	Key             string
	Suspended       bool
	SuspendedReason string
}

func newRelation(args RelationArgs) *relation {
	relation := &relation{
		Id_:              args.Id,
		Key_:             args.Key,
		Suspended_:       args.Suspended,
		SuspendedReason_: args.SuspendedReason,
	}
	relation.setEndpoints(nil)
	return relation
}

// Id implements Relation.
func (r *relation) Id() int {
	return r.Id_
}

// Key implements Relation.
func (r *relation) Key() string {
	return r.Key_
}

// Suspended implements Relation.
func (r *relation) Suspended() bool {
	return r.Suspended_
}

// SuspendedReason implements Relation.
func (r *relation) SuspendedReason() string {
	return r.SuspendedReason_
}

// Status implements Relation.
func (r *relation) Status() Status {
	// To avoid typed nils check nil here.
	if r.Status_ == nil {
		return nil
	}
	return r.Status_
}

// SetStatus implements Relation.
func (r *relation) SetStatus(args StatusArgs) {
	r.Status_ = newStatus(args)
}

// Endpoints implements Relation.
func (r *relation) Endpoints() []Endpoint {
	result := make([]Endpoint, len(r.Endpoints_.Endpoints_))
	for i, ep := range r.Endpoints_.Endpoints_ {
		result[i] = ep
	}
	return result
}

// AddEndpoint implements Relation.
func (r *relation) AddEndpoint(args EndpointArgs) Endpoint {
	ep := newEndpoint(args)
	r.Endpoints_.Endpoints_ = append(r.Endpoints_.Endpoints_, ep)
	return ep
}

func (r *relation) setEndpoints(endpointList []*endpoint) {
	r.Endpoints_ = &endpoints{
		Version:    2,
		Endpoints_: endpointList,
	}
}

func importRelations(source map[string]interface{}) ([]*relation, error) {
	checker := versionedChecker("relations")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "relations version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))

	getFields, ok := relationFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	relationList := valid["relations"].([]interface{})
	return importRelationList(relationList, schema.FieldMap(getFields()), version)
}

func importRelationList(sourceList []interface{}, checker schema.Checker, version int) ([]*relation, error) {
	result := make([]*relation, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for relation %d, %T", i, value)
		}
		// Some relations don't have status.
		// Older broken exports included status even when it was nil.
		// Remove any nil value so schema validation doesn't complain.
		if source["status"] == nil {
			delete(source, "status")
		}
		coerced, err := checker.Coerce(source, nil)
		if err != nil {
			return nil, errors.Annotatef(err, "relation %d v%d schema check failed", i, version)
		}
		valid := coerced.(map[string]interface{})
		relation, err := newRelationFromValid(valid, version)
		if err != nil {
			return nil, errors.Annotatef(err, "relation %d", i)
		}
		result = append(result, relation)
	}
	return result, nil
}

var relationFieldsFuncs = map[int]fieldsFunc{
	1: relationV1Fields,
	2: relationV2Fields,
	3: relationV3Fields,
}

func relationV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"id":        schema.Int(),
		"key":       schema.String(),
		"endpoints": schema.StringMap(schema.Any()),
	}
	return fields, nil
}

func relationV2Fields() (schema.Fields, schema.Defaults) {
	// v1 has no defaults.
	fields, _ := relationV1Fields()
	fields["status"] = schema.StringMap(schema.Any())
	defaults := schema.Defaults{"status": schema.Omit}
	return fields, defaults
}

func relationV3Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := relationV2Fields()
	fields["suspended"] = schema.Bool()
	fields["suspended-reason"] = schema.String()
	defaults["suspended"] = false
	defaults["suspended-reason"] = ""
	return fields, defaults
}

func newRelationFromValid(valid map[string]interface{}, importVersion int) (*relation, error) {
	suspended := false
	suspendedReason := ""
	if importVersion >= 3 {
		suspended = valid["suspended"].(bool)
		suspendedReason = valid["suspended-reason"].(string)
	}
	// We're always making a version 4 relation, no matter what we got on
	// the way in.
	result := &relation{
		Id_:              int(valid["id"].(int64)),
		Key_:             valid["key"].(string),
		Suspended_:       suspended,
		SuspendedReason_: suspendedReason,
	}
	// Version 1 relations don't have status info in the export yaml.
	// Some relations also don't have status.
	_, ok := valid["status"]
	if importVersion >= 2 && ok {
		relStatus, err := importStatus(valid["status"].(map[string]interface{}))
		if err != nil {
			return nil, errors.Trace(err)
		}
		result.Status_ = relStatus
	}
	endpoints, err := importEndpoints(valid["endpoints"].(map[string]interface{}))
	if err != nil {
		return nil, errors.Trace(err)
	}
	result.setEndpoints(endpoints)

	return result, nil
}
