// Copyright 2020 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// ExposedEndpointArgs is an argument struct used to create a new internal
// exposedEndpoint type that supports the ExposedEndpoint interface.
type ExposedEndpointArgs struct {
	ExposeToSpaceIDs []string
	ExposeToCIDRs    []string
}

type exposedEndpoint struct {
	Version int `yaml:"version"`

	ExposeToSpaceIDs_ []string `yaml:"expose-to-spaces,omitempty"`
	ExposeToCIDRs_    []string `yaml:"expose-to-cidrs,omitempty"`
}

func newExposedEndpoint(args ExposedEndpointArgs) *exposedEndpoint {
	return &exposedEndpoint{
		Version:           1,
		ExposeToSpaceIDs_: args.ExposeToSpaceIDs,
		ExposeToCIDRs_:    args.ExposeToCIDRs,
	}
}

// ExposeToSpaceIDs implements ExposedEndpoint.
func (exp *exposedEndpoint) ExposeToSpaceIDs() []string {
	return exp.ExposeToSpaceIDs_
}

// ExposeToCIDRs implements ExposedEndpoint.
func (exp *exposedEndpoint) ExposeToCIDRs() []string {
	return exp.ExposeToCIDRs_
}

func importExposedEndpointsMap(sourceMap map[string]interface{}) (map[string]*exposedEndpoint, error) {
	result := make(map[string]*exposedEndpoint)
	for key, value := range sourceMap {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for exposed endpoint %q, %T", key, value)
		}
		exp, err := importExposedEndpoint(source)
		if err != nil {
			return nil, errors.Trace(err)
		}
		result[key] = exp
	}
	return result, nil
}

func importExposedEndpoint(source map[string]interface{}) (*exposedEndpoint, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "expose Endpoint version schema check failed")
	}

	importFunc, ok := exposedEndpointDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	return importFunc(source)
}

type exposedEndpointDeserializationFunc func(map[string]interface{}) (*exposedEndpoint, error)

var exposedEndpointDeserializationFuncs = map[int]exposedEndpointDeserializationFunc{
	1: importExposedEndpointV1,
}

func importExposedEndpointV1(source map[string]interface{}) (*exposedEndpoint, error) {
	fields := schema.Fields{
		"expose-to-spaces": schema.List(schema.String()),
		"expose-to-cidrs":  schema.List(schema.String()),
	}
	defaults := schema.Defaults{
		"expose-to-spaces": schema.Omit,
		"expose-to-cidrs":  schema.Omit,
	}

	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "exposedEndpoint v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	res := &exposedEndpoint{
		Version: 1,
	}

	if exposeToSpaceIDs, ok := valid["expose-to-spaces"]; ok {
		res.ExposeToSpaceIDs_ = convertToStringSlice(exposeToSpaceIDs)
	}
	if exposeToCIDRs, ok := valid["expose-to-cidrs"]; ok {
		res.ExposeToCIDRs_ = convertToStringSlice(exposeToCIDRs)
	}

	return res, nil
}
