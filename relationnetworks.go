// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// RelationNetwork instances describe the ingress or egress
// networks required for a cross model relation.
type RelationNetwork interface {
	ID() string
	RelationKey() string
	CIDRS() []string
}

type relationNetworks struct {
	Version          int                `yaml:"version"`
	RelationNetworks []*relationNetwork `yaml:"relation-networks"`
}

type relationNetwork struct {
	ID_          string   `yaml:"id"`
	RelationKey_ string   `yaml:"relation-key"`
	CIDRS_       []string `yaml:"cidrs"`
}

// RelationNetworkArgs is an argument struct used to add a relation network
// to a model.
type RelationNetworkArgs struct {
	ID          string
	RelationKey string
	CIDRS       []string
}

func newRelationNetwork(args RelationNetworkArgs) *relationNetwork {
	r := &relationNetwork{
		ID_:          args.ID,
		RelationKey_: args.RelationKey,
		CIDRS_:       args.CIDRS,
	}
	return r
}

// ID implements RelationNetwork
func (r *relationNetwork) ID() string {
	return r.ID_
}

// RelationKey implements RelationNetwork
func (r *relationNetwork) RelationKey() string {
	return r.RelationKey_
}

// CIDRS implements RelationNetwork
func (r *relationNetwork) CIDRS() []string {
	return r.CIDRS_
}

func importRelationNetworks(source interface{}) ([]*relationNetwork, error) {
	checker := versionedChecker("relation-networks")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "relation networks version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	getFields, ok := relationNetworksFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["relation-networks"].([]interface{})
	return importRelationNetworkList(sourceList, schema.FieldMap(getFields()), version)
}

func importRelationNetworkList(sourceList []interface{}, checker schema.Checker, version int) ([]*relationNetwork, error) {
	result := make([]*relationNetwork, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for relation network %d, %T", i, value)
		}
		coerced, err := checker.Coerce(source, nil)
		if err != nil {
			return nil, errors.Annotatef(err, "relation network %d v%d schema check failed", i, version)
		}
		valid := coerced.(map[string]interface{})
		relationNetw, err := newRelationNetworkFromValid(valid, version)
		if err != nil {
			return nil, errors.Annotatef(err, "relation network %d", i)
		}
		result[i] = relationNetw
	}
	return result, nil
}

func newRelationNetworkFromValid(valid map[string]interface{}, version int) (*relationNetwork, error) {
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	result := &relationNetwork{
		ID_:          valid["id"].(string),
		RelationKey_: valid["relation-key"].(string),
		CIDRS_:       convertToStringSlice(valid["cidrs"]),
	}
	return result, nil
}

var relationNetworksFieldsFuncs = map[int]fieldsFunc{
	1: relationNetworksV1Fields,
}

func relationNetworksV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"id":           schema.String(),
		"relation-key": schema.String(),
		"cidrs":        schema.List(schema.String()),
	}
	defaults := schema.Defaults{}
	return fields, defaults
}
