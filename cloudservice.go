// Copyright 2018 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

type cloudService struct {
	Version int `yaml:"version"`

	ProviderId_ string   `yaml:"provider-id,omitempty"`
	Addresses_  []string `yaml:"addresses,omitempty"`
}

// ProviderId implements cloudService.
func (c *cloudService) ProviderId() string {
	return c.ProviderId_
}

// Addresses implements cloudService.
func (c *cloudService) Addresses() []string {
	return c.Addresses_
}

// CloudServiceArgs is an argument struct used to create a
// new internal cloudService type that supports the cloudService interface.
type CloudServiceArgs struct {
	ProviderId string
	Addresses  []string
}

func minimalCloudServiceArgs() CloudServiceArgs {
	return CloudServiceArgs{
		ProviderId: "provider-id",
		Addresses:  []string{"80", "443"},
	}
}

func newCloudService(args CloudServiceArgs) *cloudService {
	cloudService := &cloudService{
		Version:     1,
		ProviderId_: args.ProviderId,
		Addresses_:  args.Addresses,
	}
	return cloudService
}

func importCloudService(source map[string]interface{}) (*cloudService, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "cloudService version schema check failed")
	}

	importFunc, ok := cloudServiceDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	return importFunc(source)
}

type cloudServiceDeserializationFunc func(map[string]interface{}) (*cloudService, error)

var cloudServiceDeserializationFuncs = map[int]cloudServiceDeserializationFunc{
	1: importCloudServiceV1,
}

func importCloudServiceV1(source map[string]interface{}) (*cloudService, error) {
	fields := schema.Fields{
		"provider-id": schema.String(),
		"addresses":   schema.List(schema.String()),
	}
	// Some values don't have to be there.
	defaults := schema.Defaults{
		"provider-id": schema.Omit,
		"addresses":   schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "cloudService v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	cloudService := &cloudService{
		Version:     1,
		ProviderId_: valid["provider-id"].(string),
		Addresses_:  convertToStringSlice(valid["addresses"]),
	}

	return cloudService, nil
}
