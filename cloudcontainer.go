// Copyright 2018 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

type cloudContainer struct {
	Version int `yaml:"version"`

	ProviderId_ string   `yaml:"provider-id,omitempty"`
	Address_    string   `yaml:"address,omitempty"`
	Ports_      []string `yaml:"ports,omitempty"`
}

// ProviderId implements CloudContainer.
func (c *cloudContainer) ProviderId() string {
	return c.ProviderId_
}

// Address implements CloudContainer.
func (c *cloudContainer) Address() string {
	return c.Address_
}

// Ports implements CloudContainer.
func (c *cloudContainer) Ports() []string {
	return c.Ports_
}

// CloudContainerArgs is an argument struct used to create a
// new internal cloudContainer type that supports the CloudContainer interface.
type CloudContainerArgs struct {
	ProviderId string
	Address    string
	Ports      []string
}

func minimalCloudContainerArgs() CloudContainerArgs {
	return CloudContainerArgs{
		ProviderId: "some-provider",
		Address:    "10.0.0.1",
		Ports:      []string{"80", "443"},
	}
}

func newCloudContainer(args CloudContainerArgs) *cloudContainer {
	cloudcontainer := &cloudContainer{
		Version:     1,
		ProviderId_: args.ProviderId,
		Address_:    args.Address,
		Ports_:      args.Ports,
	}
	return cloudcontainer
}

func importCloudContainer(source map[string]interface{}) (*cloudContainer, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "cloudContainer version schema check failed")
	}

	importFunc, ok := cloudContainerDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	return importFunc(source)
}

type cloudContainerDeserializationFunc func(map[string]interface{}) (*cloudContainer, error)

var cloudContainerDeserializationFuncs = map[int]cloudContainerDeserializationFunc{
	1: importCloudContainerV1,
}

func importCloudContainerV1(source map[string]interface{}) (*cloudContainer, error) {
	fields := schema.Fields{
		"provider-id": schema.String(),
		"address":     schema.String(),
		"ports":       schema.List(schema.String()),
	}
	// Some values don't have to be there.
	defaults := schema.Defaults{
		"provider-id": schema.Omit,
		"address":     schema.Omit,
		"ports":       schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "cloudContainer v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	cloudContainer := &cloudContainer{
		Version:     1,
		ProviderId_: valid["provider-id"].(string),
		Address_:    valid["address"].(string),
		Ports_:      convertToStringSlice(valid["ports"]),
	}

	return cloudContainer, nil
}
