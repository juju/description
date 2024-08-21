// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// authorizedKeys is the internal package representation of the authorized key
// that are on a model.
type authorizedKeys struct {
	Version int `yaml:"version"`

	// UserAuthorizedKeys is a list of users and their authorized keys for a
	// model.
	UserAuthorizedKeys_ []*userAuthorizedKeys `yaml:"users-authorized-keys"`
}

// userAuthorizedKeys represents the authorized keys that a user has on a model.
type userAuthorizedKeys struct {
	Username_       string   `yaml:"user-name"`
	AuthorizedKeys_ []string `yaml:"authorized-keys"`
}

// UserAuthorizedKeysArgs is the arguments struct used for adding new authorized
// keys to the model.
type UserAuthorizedKeysArgs struct {
	// Username is the username of the user on the model that owns the
	// authorized keys.
	Username string

	// AuthorizedKeys is the set of authorized keys for the user on the model.
	AuthorizedKeys []string
}

// userAuthorizedKeysDeserializationFunc describes a type of function that is
// capable of deserializing raw users-authorized-keys attributes in
// [authorizedKeys]
type userAuthorizedKeysDeserializationFunc func(map[string]any) (*userAuthorizedKeys, error)

// userAuthorizedKeysDeserializationFuncs provides a mapping of [authorizedKeys]
// versions to the respective deserialization functions.
var userAuthorizedKeysDeserializationFuncs = map[int]userAuthorizedKeysDeserializationFunc{
	1: importAuthorizedKeysV1,
}

// newUserAuthorizedKeys creates a new [userAuthorizedKeys] from the supplied
// arguments.
func newUserAuthorizedKeys(args UserAuthorizedKeysArgs) *userAuthorizedKeys {
	return &userAuthorizedKeys{
		Username_:       args.Username,
		AuthorizedKeys_: args.AuthorizedKeys,
	}
}

// Username returns the username the owns the authorized keys in this struct.
// Implements [UserAuthorizedKeys] interface.
func (i *userAuthorizedKeys) Username() string {
	return i.Username_
}

// AuthorizedKeys returns the authorized keys for this user. Implements
// [UserAuthorizedKeys] interface.
func (i *userAuthorizedKeys) AuthorizedKeys() []string {
	return i.AuthorizedKeys_
}

// importAuthorizedKeys provides the deserialization method for importing a
// model's authorized keys. It will handle version changes of this type.
func importAuthorizedKeys(source map[string]any) ([]*userAuthorizedKeys, error) {
	checker := versionedChecker("users-authorized-keys")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "users-authorized-keys version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := userAuthorizedKeysDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["users-authorized-keys"].([]interface{})
	return importUserAuthorizedKeysList(sourceList, importFunc)
}

// importUserAuthorizedKeysList provides a deserialization method for handling
// the list of user authorized keys for a model.
func importUserAuthorizedKeysList(
	sourceList []any,
	importFunc userAuthorizedKeysDeserializationFunc,
) ([]*userAuthorizedKeys, error) {
	result := make([]*userAuthorizedKeys, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]any)
		if !ok {
			return nil, errors.Errorf("unexpected value for users-authorized-keys %d, %T", i, value)
		}
		userAuthorizedKeys, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "user-authorized-keys %d", i)
		}
		result = append(result, userAuthorizedKeys)
	}
	return result, nil
}

// importAuthorizedKeysV1 implements a [userAuthorizedKeysDeserializationFunc]
// for deserializing version 1 user authorized keys.
func importAuthorizedKeysV1(source map[string]any) (*userAuthorizedKeys, error) {
	fields := schema.Fields{
		"user-name":       schema.String(),
		"authorized-keys": schema.List(schema.String()),
	}
	defaults := schema.Defaults{}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "userAuthorizedKeys v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	keysInterface := valid["authorized-keys"].([]interface{})
	keys := make([]string, len(keysInterface))
	for i, d := range keysInterface {
		keys[i] = d.(string)
	}
	return &userAuthorizedKeys{
		Username_:       valid["user-name"].(string),
		AuthorizedKeys_: keys,
	}, nil
}
