// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
	"gopkg.in/juju/names.v3"
)

// ExternalControllerInfo holds the details required to connect to a controller.
type ExternalControllerInfo interface {
	ControllerTag() names.ControllerTag
	Alias() string
	Addrs() []string
	CACert() string
}

type externalControllerInfo struct {
	Version                 int `yaml:"version"`
	ExternalControllerInfo_ `yaml:"external-controller-info"`
}

type ExternalControllerInfo_ struct {
	ControllerTag_ string   `yaml:"controller-tag"`
	Alias_         string   `yaml:"alias"`
	Addrs_         []string `yaml:"addrs"`
	CACert_        string   `yaml:"cacert"`
}

// ExternalControllerInfoArgs is an argument struct used to add a external
// controller info to a external controller.
type ExternalControllerInfoArgs struct {
	ControllerTag names.ControllerTag
	Alias         string
	Addrs         []string
	CACert        string
}

func newExternalControllerInfo(args ExternalControllerInfoArgs) *externalControllerInfo {
	return &externalControllerInfo{
		Version: 1,
		ExternalControllerInfo_: ExternalControllerInfo_{
			ControllerTag_: args.ControllerTag.Id(),
			Alias_:         args.Alias,
			Addrs_:         args.Addrs,
			CACert_:        args.CACert,
		},
	}
}

// ControllerTag implements ExternalControllerInfo
func (e *ExternalControllerInfo_) ControllerTag() names.ControllerTag {
	return names.NewControllerTag(e.ControllerTag_)
}

// Alias implements ExternalControllerInfo
func (e *ExternalControllerInfo_) Alias() string {
	return e.Alias_
}

// Addrs implements ExternalControllerInfo
func (e *ExternalControllerInfo_) Addrs() []string {
	return e.Addrs_
}

// CACert implements ExternalControllerInfo
func (e *ExternalControllerInfo_) CACert() string {
	return e.CACert_
}

func importExternalControllerInfo(source map[string]interface{}) (*externalControllerInfo, error) {
	checker := versionedEmbeddedChecker("external-controller-info")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "external controller info version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	getFields, ok := externalControllerInfoFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	sourceInfo := valid["external-controller-info"].(map[string]interface{})
	point, err := newExternalControllerInfoFromValid(schema.FieldMap(getFields()), version, sourceInfo)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &externalControllerInfo{
		Version:                 1,
		ExternalControllerInfo_: point,
	}, nil
}

func newExternalControllerInfoFromValid(checker schema.Checker, version int, source map[string]interface{}) (ExternalControllerInfo_, error) {
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return ExternalControllerInfo_{}, errors.Annotate(err, "external controller info version schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	result := ExternalControllerInfo_{
		ControllerTag_: valid["controller-tag"].(string),
		Alias_:         valid["alias"].(string),
		Addrs_:         convertToStringSlice(valid["addrs"]),
		CACert_:        valid["cacert"].(string),
	}
	return result, nil
}

var externalControllerInfoFieldsFuncs = map[int]fieldsFunc{
	1: externalControllerInfoV1Fields,
}

func externalControllerInfoV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"controller-tag": schema.String(),
		"alias":          schema.String(),
		"addrs":          schema.List(schema.String()),
		"cacert":         schema.String(),
	}
	defaults := schema.Defaults{
		"alias": schema.Omit,
	}
	return fields, defaults
}
