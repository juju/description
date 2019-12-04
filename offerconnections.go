// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// OfferConnection represents an offer connection for a an application's endpoints.
type OfferConnection interface {
	OfferUUID() string
	RelationID() int
	RelationKey() string
	UserName() string
	SourceModelUUID() string
}

var _ OfferConnection = (*offerConnection)(nil)

type offerConnections struct {
	Version          int                `yaml:"version"`
	OfferConnections []*offerConnection `yaml:"offer-connections"`
}

type offerConnection struct {
	OfferUUID_       string `yaml:"offer-uuid"`
	RelationID_      int    `yaml:"relation-id"`
	RelationKey_     string `yaml:"relation-key"`
	UserName_        string `yaml:"user-name"`
	SourceModelUUID_ string `yaml:"source-model-uuid"`
}

// OfferConnectionArgs is an argument struct used to add a offer connection to
// the model.
type OfferConnectionArgs struct {
	OfferUUID       string
	RelationID      int
	RelationKey     string
	UserName        string
	SourceModelUUID string
}

func newOfferConnection(args OfferConnectionArgs) *offerConnection {
	return &offerConnection{
		OfferUUID_:       args.OfferUUID,
		RelationID_:      args.RelationID,
		RelationKey_:     args.RelationKey,
		UserName_:        args.UserName,
		SourceModelUUID_: args.SourceModelUUID,
	}
}

// OfferUUID returns the offer uuid for the connection.
func (c *offerConnection) OfferUUID() string {
	return c.OfferUUID_
}

// RelationID returns the relation id for the connection.
func (c *offerConnection) RelationID() int {
	return c.RelationID_
}

// RelationKey returns the relation key for the connection.
func (c *offerConnection) RelationKey() string {
	return c.RelationKey_
}

// UserName returns the user name for the connection.
func (c *offerConnection) UserName() string {
	return c.UserName_
}

// SourceModelUUID returns the user name for the connection.
func (c *offerConnection) SourceModelUUID() string {
	return c.SourceModelUUID_
}

var offerConnectionDeserializationFuncs = map[int]offerConnectionDeserializationFunc{
	1: importOfferConnectionV1,
}

func importOfferConnections(source interface{}) ([]*offerConnection, error) {
	checker := versionedChecker("offer-connections")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "offer connections version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := offerConnectionDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	sourceList := valid["offer-connections"].([]interface{})
	return importOfferConnectionList(sourceList, importFunc)
}

type offerConnectionDeserializationFunc func(interface{}) (*offerConnection, error)

func importOfferConnectionList(sourceList []interface{}, importFunc offerConnectionDeserializationFunc) ([]*offerConnection, error) {
	result := make([]*offerConnection, 0, len(sourceList))

	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for offer connection %d, %T", i, value)
		}

		offerConnection, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "offer connection %d", i)
		}
		result = append(result, offerConnection)
	}
	return result, nil
}

func offerConnectionV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"offer-uuid":        schema.String(),
		"relation-id":       schema.Int(),
		"relation-key":      schema.String(),
		"user-name":         schema.String(),
		"source-model-uuid": schema.String(),
	}
	return fields, schema.Defaults{}
}

func importOfferConnection(fields schema.Fields, defaults schema.Defaults, importVersion int, source interface{}) (*offerConnection, error) {
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "offer connection v%d schema check failed", importVersion)
	}
	valid := coerced.(map[string]interface{})

	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	result := &offerConnection{
		OfferUUID_:       valid["offer-uuid"].(string),
		RelationID_:      int(valid["relation-id"].(int64)),
		RelationKey_:     valid["relation-key"].(string),
		SourceModelUUID_: valid["source-model-uuid"].(string),
		UserName_:        valid["user-name"].(string),
	}

	return result, nil
}

func importOfferConnectionV1(source interface{}) (*offerConnection, error) {
	fields, defaults := offerConnectionV1Fields()
	return importOfferConnection(fields, defaults, 1, source)
}
