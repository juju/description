// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/collections/set"
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// Endpoint represents one end of a relation. A named endpoint provided
// by the charm that is deployed for the application.
type Endpoint interface {
	ApplicationName() string
	Name() string
	// Role, Interface, Optional, Limit, and Scope should all be available
	// through the Charm associated with the Application. There is no real need
	// for this information to be denormalised like this. However, for now,
	// since the import may well take place before the charms have been loaded
	// into the model, we'll send this information over.
	Role() string
	Interface() string
	Optional() bool
	Limit() int
	Scope() string

	// UnitCount returns the number of units the endpoint has settings for.
	UnitCount() int

	AllSettings() map[string]map[string]interface{}
	Settings(unitName string) map[string]interface{}
	SetUnitSettings(unitName string, settings map[string]interface{})
	ApplicationSettings() map[string]interface{}
	SetApplicationSettings(settings map[string]interface{})
}

type endpoints struct {
	Version    int         `yaml:"version"`
	Endpoints_ []*endpoint `yaml:"endpoints"`
}

type endpoint struct {
	ApplicationName_ string `yaml:"application-name"`
	Name_            string `yaml:"name"`
	Role_            string `yaml:"role"`
	Interface_       string `yaml:"interface"`
	Optional_        bool   `yaml:"optional"`
	Limit_           int    `yaml:"limit"`
	Scope_           string `yaml:"scope"`

	UnitSettings_        map[string]map[string]interface{} `yaml:"unit-settings"`
	ApplicationSettings_ map[string]interface{}            `yaml:"application-settings"`
}

// EndpointArgs is an argument struct used to specify a relation.
type EndpointArgs struct {
	ApplicationName string
	Name            string
	Role            string
	Interface       string
	Optional        bool
	Limit           int
	Scope           string
}

func newEndpoint(args EndpointArgs) *endpoint {
	return &endpoint{
		ApplicationName_:     args.ApplicationName,
		Name_:                args.Name,
		Role_:                args.Role,
		Interface_:           args.Interface,
		Optional_:            args.Optional,
		Limit_:               args.Limit,
		Scope_:               args.Scope,
		UnitSettings_:        make(map[string]map[string]interface{}),
		ApplicationSettings_: make(map[string]interface{}),
	}
}

func (e *endpoint) unitNames() set.Strings {
	result := set.NewStrings()
	for key := range e.UnitSettings_ {
		result.Add(key)
	}
	return result
}

// ApplicationName implements Endpoint.
func (e *endpoint) ApplicationName() string {
	return e.ApplicationName_
}

// Name implements Endpoint.
func (e *endpoint) Name() string {
	return e.Name_
}

// Role implements Endpoint.
func (e *endpoint) Role() string {
	return e.Role_
}

// Interface implements Endpoint.
func (e *endpoint) Interface() string {
	return e.Interface_
}

// Optional implements Endpoint.
func (e *endpoint) Optional() bool {
	return e.Optional_
}

// Limit implements Endpoint.
func (e *endpoint) Limit() int {
	return e.Limit_
}

// Scope implements Endpoint.
func (e *endpoint) Scope() string {
	return e.Scope_
}

// UnitCount implements Endpoint.
func (e *endpoint) UnitCount() int {
	return len(e.UnitSettings_)
}

// AllSettings implements Endpoint.
func (e *endpoint) AllSettings() map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	for name, settings := range e.UnitSettings_ {
		result[name] = settings
	}
	return result
}

// Settings implements Endpoint.
func (e *endpoint) Settings(unitName string) map[string]interface{} {
	return e.UnitSettings_[unitName]
}

// SetUnitSettings implements Endpoint.
func (e *endpoint) SetUnitSettings(unitName string, settings map[string]interface{}) {
	e.UnitSettings_[unitName] = settings
}

// ApplicationSettings implements Endpoint.
func (e *endpoint) ApplicationSettings() map[string]interface{} {
	return e.ApplicationSettings_
}

// SetApplicationSettings implements Endpoint.
func (e *endpoint) SetApplicationSettings(settings map[string]interface{}) {
	e.ApplicationSettings_ = settings
}

func importEndpoints(source map[string]interface{}) ([]*endpoint, error) {
	checker := versionedChecker("endpoints")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "endpoints version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	getFields, ok := endpointFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	endpointList := valid["endpoints"].([]interface{})
	return importEndpointList(endpointList, schema.FieldMap(getFields()), version)
}

func importEndpointList(sourceList []interface{}, checker schema.Checker, version int) ([]*endpoint, error) {
	result := make([]*endpoint, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for endpoint %d, %T", i, value)
		}
		coerced, err := checker.Coerce(source, nil)
		if err != nil {
			return nil, errors.Annotatef(err, "endpoint %d v%d schema check failed", i, version)
		}
		valid := coerced.(map[string]interface{})
		endpoint, err := newEndpointFromValid(valid, version)
		if err != nil {
			return nil, errors.Annotatef(err, "endpoint %d", i)
		}
		result = append(result, endpoint)
	}
	return result, nil
}

var endpointFieldsFuncs = map[int]fieldsFunc{
	1: endpointV1Fields,
	2: endpointV2Fields,
}

func endpointV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"application-name": schema.String(),
		"name":             schema.String(),
		"role":             schema.String(),
		"interface":        schema.String(),
		"optional":         schema.Bool(),
		"limit":            schema.Int(),
		"scope":            schema.String(),
		"unit-settings":    schema.StringMap(schema.StringMap(schema.Any())),
	}
	return fields, nil
}

func endpointV2Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := endpointV1Fields()
	fields["application-settings"] = schema.StringMap(schema.Any())
	return fields, defaults
}

func newEndpointFromValid(valid map[string]interface{}, version int) (*endpoint, error) {
	result := &endpoint{
		ApplicationName_:     valid["application-name"].(string),
		Name_:                valid["name"].(string),
		Role_:                valid["role"].(string),
		Interface_:           valid["interface"].(string),
		Optional_:            valid["optional"].(bool),
		Limit_:               int(valid["limit"].(int64)),
		Scope_:               valid["scope"].(string),
		UnitSettings_:        make(map[string]map[string]interface{}),
		ApplicationSettings_: make(map[string]interface{}),
	}

	for unitname, settings := range valid["unit-settings"].(map[string]interface{}) {
		result.UnitSettings_[unitname] = settings.(map[string]interface{})
	}

	if version >= 2 {
		result.ApplicationSettings_ = valid["application-settings"].(map[string]interface{})
	}

	return result, nil
}
