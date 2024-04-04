// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

type spaces struct {
	Version int      `yaml:"version"`
	Spaces_ []*space `yaml:"spaces"`
}

type space struct {
	Id_         string `yaml:"id"`
	UUID_       string `yaml:"uuid"`
	Name_       string `yaml:"name"`
	Public_     bool   `yaml:"public"`
	ProviderID_ string `yaml:"provider-id,omitempty"`
}

// SpaceArgs is an argument struct used to create a new internal space
// type that supports the Space interface.
type SpaceArgs struct {
	Id         string
	UUID       string
	Name       string
	Public     bool
	ProviderID string
}

func newSpace(args SpaceArgs) *space {
	return &space{
		Id_:         args.Id,
		UUID_:       args.UUID,
		Name_:       args.Name,
		Public_:     args.Public,
		ProviderID_: args.ProviderID,
	}
}

// Id implements Space.
func (s *space) Id() string {
	return s.Id_
}

// UUID implements Space.
func (s *space) UUID() string {
	return s.UUID_
}

// Name implements Space.
func (s *space) Name() string {
	return s.Name_
}

// Public implements Space.
func (s *space) Public() bool {
	return s.Public_
}

// ProviderID implements Space.
func (s *space) ProviderID() string {
	return s.ProviderID_
}

func importSpaces(source map[string]interface{}) ([]*space, error) {
	checker := versionedChecker("spaces")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "spaces version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	getFields, ok := spaceFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["spaces"].([]interface{})
	return importSpaceList(sourceList, schema.FieldMap(getFields()), version)
}

func importSpaceList(sourceList []interface{}, checker schema.Checker, version int) ([]*space, error) {
	result := make([]*space, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for space %d, %T", i, value)
		}
		coerced, err := checker.Coerce(source, nil)

		if err != nil {
			return nil, errors.Annotatef(err, "space %d v%d schema check failed", i, version)
		}
		valid := coerced.(map[string]interface{})
		space, err := newSpaceFromValid(valid, version)
		if err != nil {
			return nil, errors.Annotatef(err, "space %d", i)
		}
		result = append(result, space)
	}
	return result, nil
}

func newSpaceFromValid(valid map[string]interface{}, version int) (*space, error) {
	result := space{
		Name_:       valid["name"].(string),
		Public_:     valid["public"].(bool),
		ProviderID_: valid["provider-id"].(string),
	}
	// id was added in V2 and removed in V3.
	if version == 2 {
		result.Id_ = valid["id"].(string)
	}
	if version >= 3 {
		result.UUID_ = valid["uuid"].(string)
	}
	return &result, nil
}

var spaceFieldsFuncs = map[int]fieldsFunc{
	1: spaceV1Fields,
	2: spaceV2Fields,
	3: spaceV3Fields,
}

func spaceV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"name":        schema.String(),
		"public":      schema.Bool(),
		"provider-id": schema.String(),
	}
	// Some values don't have to be there.
	defaults := schema.Defaults{
		"provider-id": "",
	}
	return fields, defaults
}

func spaceV2Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := spaceV1Fields()
	fields["id"] = schema.String()

	return fields, defaults
}

func spaceV3Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := spaceV2Fields()
	fields["uuid"] = schema.String()
	delete(fields, "id")

	return fields, defaults
}
