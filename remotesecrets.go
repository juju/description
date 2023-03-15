// Copyright 2023 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/names/v4"
	"github.com/juju/schema"
	"github.com/rs/xid"
)

// RemoteSecret represents consumer info for a remote secret.
type RemoteSecret interface {
	ID() string
	SourceUUID() string
	Consumer() (names.Tag, error)
	Label() string
	CurrentRevision() int
	LatestRevision() int

	Validate() error
}

type remoteSecrets struct {
	Version        int             `yaml:"version"`
	RemoteSecrets_ []*remoteSecret `yaml:"remote-secrets"`
}

type remoteSecret struct {
	ID_              string `yaml:"id"`
	SourceUUID_      string `yaml:"source-uuid"`
	Consumer_        string `yaml:"consumer"`
	Label_           string `yaml:"label"`
	CurrentRevision_ int    `yaml:"current-revision"`
	LatestRevision_  int    `yaml:"latest-revision"`
}

// RemoteSecretArgs is an argument struct used to create a
// new internal remote secret type that supports the remote
// secret interface.
type RemoteSecretArgs struct {
	ID              string
	SourceUUID      string
	Consumer        names.Tag
	Label           string
	CurrentRevision int
	LatestRevision  int
}

func newRemoteSecret(args RemoteSecretArgs) *remoteSecret {
	s := &remoteSecret{
		ID_:              args.ID,
		SourceUUID_:      args.SourceUUID,
		Label_:           args.Label,
		CurrentRevision_: args.CurrentRevision,
		LatestRevision_:  args.LatestRevision,
	}
	if args.Consumer != nil {
		s.Consumer_ = args.Consumer.String()
	}
	return s
}

// ID implements RemoteSecret.
func (i *remoteSecret) ID() string {
	return i.ID_
}

// SourceUUID implements RemoteSecret.
func (i *remoteSecret) SourceUUID() string {
	return i.SourceUUID_
}

// Consumer implements RemoteSecret.
func (i *remoteSecret) Consumer() (names.Tag, error) {
	if i.Consumer_ == "" {
		return nil, nil
	}
	tag, err := names.ParseTag(i.Consumer_)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return tag, nil
}

// Label implements RemoteSecret.
func (i *remoteSecret) Label() string {
	return i.Label_
}

// CurrentRevision implements RemoteSecret.
func (i *remoteSecret) CurrentRevision() int {
	return i.CurrentRevision_
}

// LatestRevision implements RemoteSecret.
func (i *remoteSecret) LatestRevision() int {
	return i.LatestRevision_
}

// Validate implements RemoteSecret.
func (i *remoteSecret) Validate() error {
	if i.ID_ == "" {
		return errors.NotValidf("remote secret missing id")
	}
	if _, err := xid.FromString(i.ID_); err != nil {
		return errors.Wrap(err, errors.NotValidf("remote secret ID %q", i.ID_))
	}
	if _, err := i.Consumer(); err != nil {
		return errors.Wrap(err, errors.NotValidf("remote secret %q invalid consumer", i.ID_))
	}
	return nil
}

func importRemoteSecrets(source map[string]interface{}) ([]*remoteSecret, error) {
	checker := versionedChecker("remote-secrets")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "remote secret version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	sourceList := valid["remote-secrets"].([]interface{})
	return importRemoteSecretList(sourceList, version)
}

func importRemoteSecretList(sourceList []interface{}, version int) ([]*remoteSecret, error) {
	getFields, ok := remoteSecretFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	result := make([]*remoteSecret, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for remote secret %d, %T", i, value)
		}
		secretConsumer, err := importRemoteSecret(source, version, getFields)
		if err != nil {
			return nil, errors.Annotatef(err, "remote secret %d", i)
		}
		result = append(result, secretConsumer)
	}
	return result, nil
}

var remoteSecretFieldsFuncs = map[int]fieldsFunc{
	1: remoteSecretV1Fields,
}

func remoteSecretV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"id":               schema.String(),
		"source-uuid":      schema.String(),
		"consumer":         schema.String(),
		"label":            schema.String(),
		"current-revision": schema.Int(),
		"latest-revision":  schema.Int(),
	}
	// Some values don't have to be there.
	defaults := schema.Defaults{
		"label": schema.Omit,
	}
	return fields, defaults
}

func importRemoteSecret(source map[string]interface{}, importVersion int, fieldFunc func() (schema.Fields, schema.Defaults)) (*remoteSecret, error) {
	fields, defaults := fieldFunc()
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "remote secrets v%d schema check failed", importVersion)
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	consumer := &remoteSecret{
		ID_:              valid["id"].(string),
		SourceUUID_:      valid["source-uuid"].(string),
		Consumer_:        valid["consumer"].(string),
		Label_:           valid["label"].(string),
		CurrentRevision_: int(valid["current-revision"].(int64)),
		LatestRevision_:  int(valid["latest-revision"].(int64)),
	}
	return consumer, nil
}
