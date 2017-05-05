// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// RemoteSubnet represents a subnet in a remote space.
type RemoteSubnet interface {
	CIDR() string
	ProviderId() string
	VLANTag() int
	AvailabilityZones() []string
	ProviderSpaceId() string
	ProviderNetworkId() string
}

type remoteSubnets struct {
	Version int             `yaml:"version"`
	Subnets []*remoteSubnet `yaml:"subnets"`
}

type remoteSubnet struct {
	CIDR_              string   `yaml:"cidr"`
	ProviderId_        string   `yaml:"provider-id"`
	VLANTag_           int      `yaml:"vlan-tag,omitempty"`
	AvailabilityZones_ []string `yaml:"availability-zones,omitempty"`
	ProviderSpaceId_   string   `yaml:"provider-space-id"`
	ProviderNetworkId_ string   `yaml:"provider-network-id"`
}

// RemoteSubnetArgs is an argument struct used to add a remote subnet
// to a remote space.
type RemoteSubnetArgs struct {
	CIDR              string
	ProviderId        string
	VLANTag           int
	AvailabilityZones []string
	ProviderSpaceId   string
	ProviderNetworkId string
}

func newRemoteSubnet(args RemoteSubnetArgs) *remoteSubnet {
	return &remoteSubnet{
		CIDR_:              args.CIDR,
		ProviderId_:        args.ProviderId,
		VLANTag_:           args.VLANTag,
		AvailabilityZones_: args.AvailabilityZones,
		ProviderSpaceId_:   args.ProviderSpaceId,
		ProviderNetworkId_: args.ProviderNetworkId,
	}
}

// CIDR implements RemoteSubnet.
func (s *remoteSubnet) CIDR() string {
	return s.CIDR_
}

// ProviderId implements RemoteSubnet.
func (s *remoteSubnet) ProviderId() string {
	return s.ProviderId_
}

// VLANTag implements RemoteSubnet.
func (s *remoteSubnet) VLANTag() int {
	return s.VLANTag_
}

// AvailabilityZones implements RemoteSubnet.
func (s *remoteSubnet) AvailabilityZones() []string {
	return s.AvailabilityZones_
}

// ProviderSpaceId implements RemoteSubnet.
func (s *remoteSubnet) ProviderSpaceId() string {
	return s.ProviderSpaceId_
}

// ProviderNetworkId implements RemoteSubnet.
func (s *remoteSubnet) ProviderNetworkId() string {
	return s.ProviderNetworkId_
}

func importRemoteSubnets(source map[string]interface{}) ([]*remoteSubnet, error) {
	checker := versionedChecker("subnets")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "remote subnets version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	getFields, ok := remoteSubnetFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["subnets"].([]interface{})
	return importRemoteSubnetList(sourceList, schema.FieldMap(getFields()), version)
}

func importRemoteSubnetList(sourceList []interface{}, checker schema.Checker, version int) ([]*remoteSubnet, error) {
	var result []*remoteSubnet
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for remote subnet %d, %T", i, value)
		}
		coerced, err := checker.Coerce(source, nil)
		if err != nil {
			return nil, errors.Annotatef(err, "remote subnet %d v%d schema check failed", i, version)
		}
		valid := coerced.(map[string]interface{})
		subnet, err := newRemoteSubnetFromValid(valid, version)
		if err != nil {
			return nil, errors.Annotatef(err, "remote subnet %d", i)
		}
		result = append(result, subnet)
	}
	return result, nil
}

func newRemoteSubnetFromValid(valid map[string]interface{}, version int) (*remoteSubnet, error) {
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	result := remoteSubnet{
		CIDR_:              valid["cidr"].(string),
		ProviderId_:        valid["provider-id"].(string),
		VLANTag_:           int(valid["vlan-tag"].(int64)),
		ProviderSpaceId_:   valid["provider-space-id"].(string),
		ProviderNetworkId_: valid["provider-network-id"].(string),
	}
	if availabilityZones, ok := valid["availability-zones"]; ok {
		result.AvailabilityZones_ = convertToStringSlice(availabilityZones)
	}
	return &result, nil
}

var remoteSubnetFieldsFuncs = map[int]fieldsFunc{
	1: remoteSubnetV1Fields,
}

func remoteSubnetV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"cidr":                schema.String(),
		"provider-id":         schema.String(),
		"vlan-tag":            schema.Int(),
		"availability-zones":  schema.List(schema.String()),
		"provider-space-id":   schema.String(),
		"provider-network-id": schema.String(),
	}
	defaults := schema.Defaults{
		"vlan-tag":           0,
		"availability-zones": schema.Omit,
	}
	return fields, defaults
}
