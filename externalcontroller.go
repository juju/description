// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// ExternalController represents the state of a controller hosting
// other models.
type ExternalController interface {
	ID() string
	Alias() string
	Addrs() []string
	CACert() string
	Models() []string
}

type externalControllers struct {
	Version             int                   `yaml:"version"`
	ExternalControllers []*externalController `yaml:"external-controllers"`
}

type externalController struct {
	ID_     string   `yaml:"id"`
	Alias_  string   `yaml:"alias,omitempty"`
	Addrs_  []string `yaml:"addrs"`
	CACert_ string   `yaml:"ca-cert"`
	Models_ []string `yaml:"models"`
}

// ExternalControllerArgs is an argument struct used to add a external
// controller to a model.
type ExternalControllerArgs struct {
	ID     string
	Alias  string
	Addrs  []string
	CACert string
	Models []string
}

func newExternalController(args ExternalControllerArgs) *externalController {
	return &externalController{
		ID_:     args.ID,
		Alias_:  args.Alias,
		Addrs_:  args.Addrs,
		CACert_: args.CACert,
		Models_: args.Models,
	}
}

// ID returns the controller tag for the external controller.
func (e *externalController) ID() string {
	return e.ID_
}

// Alias returns the controller name for the external controller.
func (e *externalController) Alias() string {
	return e.Alias_
}

// Addrs returns the addresses for the external controller.
func (e *externalController) Addrs() []string {
	return e.Addrs_
}

// CACert returns the ca cert for the external controller.
func (e *externalController) CACert() string {
	return e.CACert_
}

// Models returns the list of models for the external controller.
func (e *externalController) Models() []string {
	return e.Models_
}

func importExternalControllers(source interface{}) ([]*externalController, error) {
	checker := versionedChecker("external-controllers")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "external controllers version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := externalControllerDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["external-controllers"].([]interface{})
	return importExternalControllerList(sourceList, importFunc)
}

func importExternalControllerList(sourceList []interface{}, importFunc externalControllerDeserializationFunc) ([]*externalController, error) {
	result := make([]*externalController, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for external controller %d, %T", i, value)
		}

		externalController, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "external controller %d", i)
		}
		result[i] = externalController
	}
	return result, nil
}

var externalControllerDeserializationFuncs = map[int]externalControllerDeserializationFunc{
	1: importExternalControllerV1,
}

type externalControllerDeserializationFunc func(interface{}) (*externalController, error)

func externalControllerV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"id":      schema.String(),
		"alias":   schema.String(),
		"addrs":   schema.List(schema.String()),
		"ca-cert": schema.String(),
		"models":  schema.List(schema.String()),
	}
	defaults := schema.Defaults{
		"alias": schema.Omit,
	}
	return fields, defaults
}

func importExternalController(fields schema.Fields, defaults schema.Defaults, importVersion int, source interface{}) (*externalController, error) {
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "external controller v%d schema check failed", importVersion)
	}
	valid := coerced.(map[string]interface{})

	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	result := &externalController{
		ID_:     valid["id"].(string),
		Addrs_:  convertToStringSlice(valid["addrs"]),
		CACert_: valid["ca-cert"].(string),
		Models_: convertToStringSlice(valid["models"]),
	}

	// Alias is optional through out juju and because of that, it isn't a
	// requirement of external controller migrations.
	if alias, ok := valid["alias"]; ok {
		result.Alias_ = alias.(string)
	}

	return result, nil
}

func importExternalControllerV1(source interface{}) (*externalController, error) {
	fields, defaults := externalControllerV1Fields()
	return importExternalController(fields, defaults, 1, source)
}
