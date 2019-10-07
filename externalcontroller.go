// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
	"gopkg.in/juju/names.v3"
)

// ExternalController represents the state of a controller hosting
// other models.
type ExternalController interface {
	ID() names.ControllerTag
	Alias() string
	Addrs() []string
	CACert() string
}

type externalControllers struct {
	Version             int                   `yaml:"version"`
	ExternalControllers []*externalController `yaml:"external-controllers"`
}

type externalController struct {
	ID_     string   `yaml:"id"`
	Alias_  string   `yaml:"alias"`
	Addrs_  []string `yaml:"addrs"`
	CACert_ string   `yaml:"cacert"`
}

// ExternalControllerArgs is an argument struct used to add a external
// controller to a model.
type ExternalControllerArgs struct {
	Tag    names.ControllerTag
	Alias  string
	Addrs  []string
	CACert string
}

func newExternalController(args ExternalControllerArgs) *externalController {
	return &externalController{
		ID_:     args.Tag.Id(),
		Alias_:  args.Alias,
		Addrs_:  args.Addrs,
		CACert_: args.CACert,
	}
}

// ID implements ExternalController
func (e *externalController) ID() names.ControllerTag {
	return names.NewControllerTag(e.ID_)
}

// Alias implements ExternalController
func (e *externalController) Alias() string {
	return e.Alias_
}

// Addrs implements ExternalController
func (e *externalController) Addrs() []string {
	return e.Addrs_
}

// CACert implements ExternalController
func (e *externalController) CACert() string {
	return e.CACert_
}

func importExternalControllers(source interface{}) ([]*externalController, error) {
	checker := versionedChecker("external-controllers")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "external controllers version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	getFields, ok := externalControllerFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["external-controllers"].([]interface{})
	return importExternalControllerList(sourceList, schema.FieldMap(getFields()), version)
}

func importExternalControllerList(sourceList []interface{}, checker schema.Checker, version int) ([]*externalController, error) {
	result := make([]*externalController, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for external controller %d, %T", i, value)
		}
		coerced, err := checker.Coerce(source, nil)
		if err != nil {
			return nil, errors.Annotatef(err, "external controller %d v%d schema check failed", i, version)
		}
		valid := coerced.(map[string]interface{})
		externalCtrl, err := newExternalControllerFromValid(valid, version)
		if err != nil {
			return nil, errors.Annotatef(err, "external controller %d", i)
		}
		result[i] = externalCtrl
	}
	return result, nil
}

func newExternalControllerFromValid(valid map[string]interface{}, version int) (*externalController, error) {
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	result := &externalController{
		ID_:     valid["id"].(string),
		Alias_:  valid["alias"].(string),
		Addrs_:  convertToStringSlice(valid["addrs"]),
		CACert_: valid["cacert"].(string),
	}

	return result, nil
}

var externalControllerFieldsFuncs = map[int]fieldsFunc{
	1: externalControllerV1Fields,
}

func externalControllerV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"id":     schema.String(),
		"alias":  schema.String(),
		"addrs":  schema.List(schema.String()),
		"cacert": schema.String(),
	}
	defaults := schema.Defaults{
		"alias": schema.Omit,
	}
	return fields, defaults
}
