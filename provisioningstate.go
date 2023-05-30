// Copyright 2023 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

type ProvisioningState interface {
	Scaling() bool
	ScaleTarget() int
}

type provisioningState struct {
	Version_     int  `yaml:"version"`
	Scaling_     bool `yaml:"scaling"`
	ScaleTarget_ int  `yaml:"scale-target"`
}

func (i *provisioningState) Scaling() bool {
	return i.Scaling_
}

func (i *provisioningState) ScaleTarget() int {
	return i.ScaleTarget_
}

// ProvisioningStateArgs is an argument struct used to create a
// new internal provisioningState type that supports the ProvisioningState interface.
type ProvisioningStateArgs struct {
	Scaling     bool
	ScaleTarget int
}

func newProvisioningState(args *ProvisioningStateArgs) *provisioningState {
	if args == nil {
		return nil
	}
	return &provisioningState{
		Version_:     1,
		Scaling_:     args.Scaling,
		ScaleTarget_: args.ScaleTarget,
	}
}

func importProvisioningState(source map[string]interface{}) (*provisioningState, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "provisioning-state version schema check failed")
	}
	importFunc, ok := provisioningStateDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	return importFunc(source)
}

type provisioningStateDeserializationFunc func(map[string]interface{}) (*provisioningState, error)

var provisioningStateDeserializationFuncs = map[int]provisioningStateDeserializationFunc{
	1: importProvisioningStateV1,
}

func importProvisioningStateV1(source map[string]interface{}) (*provisioningState, error) {
	fields, defaults := provisioningStateV1Schema()
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "provisioning-state v1 schema check failed")
	}

	return provisioningStateV1(coerced.(map[string]interface{})), nil
}

func provisioningStateV1Schema() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"scaling":      schema.Bool(),
		"scale-target": schema.Int(),
	}
	defaults := schema.Defaults{
		"scaling":      false,
		"scale-target": 0,
	}
	return fields, defaults
}

func provisioningStateV1(valid map[string]interface{}) *provisioningState {
	return &provisioningState{
		Version_:     1,
		Scaling_:     valid["scaling"].(bool),
		ScaleTarget_: int(valid["scale-target"].(int64)),
	}
}
