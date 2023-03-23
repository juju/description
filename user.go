// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"time"

	"github.com/juju/errors"
	"github.com/juju/names/v4"
	"github.com/juju/schema"
)

// User represents a user of the model. Users are able to connect to, and
// depending on the read only flag, modify the model.
type User interface {
	Name() names.UserTag
	DisplayName() string
	CreatedBy() names.UserTag
	DateCreated() time.Time
	LastConnection() time.Time
	Access() string
	UserRemovedLogEntry() []UserRemovedLogEntry
}

type users struct {
	Version int     `yaml:"version"`
	Users_  []*user `yaml:"users"`
}

type UserArgs struct {
	Name           names.UserTag
	DisplayName    string
	CreatedBy      names.UserTag
	DateCreated    time.Time
	LastConnection time.Time
	Access         string
	// RemovalLog keeps a track of removals for this user
	RemovalLog []UserRemovedLogEntryArgs
}

// userRemovedLogEntryArgs is an argument struct used to create a
// new internal UserRemovedLogEntry
type UserRemovedLogEntryArgs struct {
	RemovedBy   string
	DateCreated time.Time
	DateRemoved time.Time
}

type UserRemovedLogEntry struct {
	RemovedBy   string
	DateCreated time.Time
	DateRemoved time.Time
}

func newUser(args UserArgs) *user {
	u := &user{
		Name_:        args.Name.Id(),
		DisplayName_: args.DisplayName,
		CreatedBy_:   args.CreatedBy.Id(),
		DateCreated_: args.DateCreated,
		Access_:      args.Access,
	}
	if !args.LastConnection.IsZero() {
		value := args.LastConnection
		u.LastConnection_ = &value
	}
	if args.RemovalLog != nil {
		u.setUserRemovedLogEntry(args.RemovalLog)
	}
	return u
}

type user struct {
	Name_        string    `yaml:"name"`
	DisplayName_ string    `yaml:"display-name,omitempty"`
	CreatedBy_   string    `yaml:"created-by"`
	DateCreated_ time.Time `yaml:"date-created"`
	Access_      string    `yaml:"access"`
	// Can't use omitempty with time.Time, it just doesn't work,
	// so use a pointer in the struct.
	LastConnection_ *time.Time            `yaml:"last-connection,omitempty"`
	RemovalLog_     []userRemovedLogEntry `yaml:"user-removed-log,omitempty"`
}

type userRemovedLogEntry struct {
	RemovedBy_   string    `yaml:"removed-by,omitempty"`
	DateCreated_ time.Time `yaml:"date-created,omitempty"`
	DateRemoved_ time.Time `yaml:"date-removed,omitempty"`
}

func newUserRemovedLogEntry(args UserRemovedLogEntryArgs) userRemovedLogEntry {
	return userRemovedLogEntry{
		RemovedBy_:   args.RemovedBy,
		DateCreated_: args.DateCreated,
		DateRemoved_: args.DateRemoved,
	}
}

func (u *userRemovedLogEntry) RemovedBy() string {
	return u.RemovedBy_
}

func (u *userRemovedLogEntry) DateCreated() time.Time {
	return u.DateCreated_
}

func (u *userRemovedLogEntry) DateRemoved() time.Time {
	return u.DateRemoved_
}

// Name implements User.
func (u *user) Name() names.UserTag {
	return names.NewUserTag(u.Name_)
}

// DisplayName implements User.
func (u *user) DisplayName() string {
	return u.DisplayName_
}

// CreatedBy implements User.
func (u *user) CreatedBy() names.UserTag {
	return names.NewUserTag(u.CreatedBy_)
}

// DateCreated implements User.
func (u *user) DateCreated() time.Time {
	return u.DateCreated_
}

// LastConnection implements User.
func (u *user) LastConnection() time.Time {
	var zero time.Time
	if u.LastConnection_ == nil {
		return zero
	}
	return *u.LastConnection_
}

// Access implements User.
func (u *user) Access() string {
	return u.Access_
}

func (u *user) setUserRemovedLogEntry(args []UserRemovedLogEntryArgs) {
	u.RemovalLog_ = make([]userRemovedLogEntry, len(args))
	for i, entry := range args {
		u.RemovalLog_[i] = newUserRemovedLogEntry(entry)
	}
}

// UserRemovedLogEntry implements User.
func (u *user) UserRemovedLogEntry() []UserRemovedLogEntry {
	var result []UserRemovedLogEntry
	for _, entry := range u.RemovalLog_ {
		aux := UserRemovedLogEntry{
			RemovedBy:   entry.RemovedBy_,
			DateCreated: entry.DateCreated_,
			DateRemoved: entry.DateRemoved_,
		}
		result = append(result, aux)
	}
	return result
}

func importUsers(source map[string]interface{}) ([]*user, error) {
	checker := versionedChecker("users")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "users version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := userDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["users"].([]interface{})
	return importUserList(sourceList, importFunc)
}

func importUserList(sourceList []interface{}, importFunc userDeserializationFunc) ([]*user, error) {
	result := make([]*user, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for user %d, %T", i, value)
		}
		user, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "user %d", i)
		}
		result = append(result, user)
	}
	return result, nil
}

type userDeserializationFunc func(map[string]interface{}) (*user, error)

var userDeserializationFuncs = map[int]userDeserializationFunc{
	1: importUserV1,
	2: importUserV2,
}

func userV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"name":            schema.String(),
		"display-name":    schema.String(),
		"created-by":      schema.String(),
		"read-only":       schema.Bool(),
		"date-created":    schema.Time(),
		"last-connection": schema.Time(),
		"access":          schema.String(),
	}

	// Some values don't have to be there.
	defaults := schema.Defaults{
		"display-name":    "",
		"last-connection": schema.Omit,
		"read-only":       false,
	}
	return fields, defaults
}

func importUserV1(source map[string]interface{}) (*user, error) {
	fields, defaults := userV1Fields()
	checker := schema.FieldMap(fields, defaults)
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "user v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	result := &user{
		Name_:           valid["name"].(string),
		DisplayName_:    valid["display-name"].(string),
		CreatedBy_:      valid["created-by"].(string),
		DateCreated_:    valid["date-created"].(time.Time),
		Access_:         valid["access"].(string),
		LastConnection_: fieldToTimePtr(valid, "last-connection"),
	}
	return result, nil

}

func userV2Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := userV1Fields()
	// add the user-removed-log field
	fields["user-removed-log"] = schema.List(schema.Any())
	defaults["user-removed-log"] = schema.Omit
	return fields, defaults
}

func importUserV2(source map[string]interface{}) (*user, error) {
	fields, defaults := userV2Fields()
	checker := schema.FieldMap(fields, defaults)
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "user v2 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	result := &user{
		Name_:           valid["name"].(string),
		DisplayName_:    valid["display-name"].(string),
		CreatedBy_:      valid["created-by"].(string),
		DateCreated_:    valid["date-created"].(time.Time),
		Access_:         valid["access"].(string),
		LastConnection_: fieldToTimePtr(valid, "last-connection"),
	}

	// the removed-log can be empty
	removedEntry, ok := valid["user-removed-log"]
	if !ok {
		result.RemovalLog_ = []userRemovedLogEntry{}
	} else {
		removed, err := importUsersRemovedLog(removedEntry.([]any))
		if err != nil {
			return nil, errors.Trace(err)
		}
		result.RemovalLog_ = removed
	}

	return result, nil
}

func importUsersRemovedLogEntry(source map[any]any) (*userRemovedLogEntry, error) {
	fields := schema.Fields{
		"removed-by":   schema.String(),
		"date-created": schema.Time(),
		"date-removed": schema.Time(),
	}
	defaults := schema.Defaults{}
	checker := schema.FieldMap(fields, defaults)
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "usersRemovedLogEntry v2 failed")
	}
	valid := coerced.(map[string]any)
	result := &userRemovedLogEntry{
		RemovedBy_:   valid["removed-by"].(string),
		DateCreated_: valid["date-created"].(time.Time),
		DateRemoved_: valid["date-removed"].(time.Time),
	}
	return result, nil
}

func importUsersRemovedLog(sourceInput []any) ([]userRemovedLogEntry, error) {
	result := make([]userRemovedLogEntry, len(sourceInput))
	for i, value := range sourceInput {
		source, ok := value.(map[any]any)
		if !ok {
			return nil, errors.Errorf("unexpected value for revision %d, %T", i, value)
		}
		entry, err := importUsersRemovedLogEntry(source)
		if err != nil {
			return nil, errors.Annotatef(err, "UserRemovedLogEntry %d", i)
		}
		result[i] = *entry
	}
	return result, nil
}
