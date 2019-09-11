// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// RemoteEntity represents the internal state of a remote entity.
// Remote entities may be exported local entities, or imported
// remote entities
type RemoteEntity interface {
	Token() string
	Macaroon() string
}

type remoteEntities struct {
	Version        int             `yaml:"version"`
	RemoteEntities []*remoteEntity `yaml:"remote-entities"`
}

type remoteEntity struct {
	Token_    string `yaml:"token"`
	Macaroon_ string `yaml:"macaroon"`
}

// RemoteEntityArgs is an argument struct used to add a remote entity.
type RemoteEntityArgs struct {
	Token    string
	Macaroon string
}

func newRemoteEntity(args RemoteEntityArgs) *remoteEntity {
	f := &remoteEntity{
		Token_:    args.Token,
		Macaroon_: args.Macaroon,
	}
	return f
}

// Token implements RemoteEntity
func (f *remoteEntity) Token() string {
	return f.Token_
}

// Macaroon implements RemoteEntity
func (f *remoteEntity) Macaroon() string {
	return f.Macaroon_
}

func importRemoteEntities(source interface{}) ([]*remoteEntity, error) {
	checker := versionedChecker("remote-entities")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "remote entities version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	getFields, ok := remoteEntityFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["remote-entities"].([]interface{})
	return importRemoteEntityList(sourceList, schema.FieldMap(getFields()), version)
}

func importRemoteEntityList(sourceList []interface{}, checker schema.Checker, version int) ([]*remoteEntity, error) {
	result := make([]*remoteEntity, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for remote entity %d, %T", i, value)
		}
		coerced, err := checker.Coerce(source, nil)
		if err != nil {
			return nil, errors.Annotatef(err, "remote entity %d v%d schema check failed", i, version)
		}
		valid := coerced.(map[string]interface{})
		remoteEnt, err := newRemoteEntityFromValid(valid, version)
		if err != nil {
			return nil, errors.Annotatef(err, "remote entity %d", i)
		}
		result[i] = remoteEnt
	}
	return result, nil
}

func newRemoteEntityFromValid(valid map[string]interface{}, version int) (*remoteEntity, error) {
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	result := &remoteEntity{
		Token_:    valid["token"].(string),
		Macaroon_: valid["macaroon"].(string),
	}
	return result, nil
}

var remoteEntityFieldsFuncs = map[int]fieldsFunc{
	1: remoteEntityV1Fields,
}

func remoteEntityV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"token":    schema.String(),
		"macaroon": schema.String(),
	}
	defaults := schema.Defaults{
		"macaroon": schema.Omit,
	}
	return fields, defaults
}
