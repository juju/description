// Copyright 2023 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/names/v6"
	"github.com/juju/schema"
)

// SecretRemoteConsumer represents consumer info for a secret remote consumer.
type SecretRemoteConsumer interface {
	ID() string
	Consumer() (names.Tag, error)
	CurrentRevision() int
	LatestRevision() int
}

type secretRemoteConsumer struct {
	ID_              string `yaml:"id"`
	Consumer_        string `yaml:"consumer"`
	CurrentRevision_ int    `yaml:"current-revision"`

	// Updated when added to a secret
	// but not exported.
	LatestRevision_ int `yaml:"-"`
}

// SecretRemoteConsumerArgs is an argument struct used to create a
// new internal remote secret type that supports the secret remote
// consumer interface.
type SecretRemoteConsumerArgs struct {
	ID              string
	Consumer        names.Tag
	CurrentRevision int
}

func newSecretRemoteConsumer(args SecretRemoteConsumerArgs) *secretRemoteConsumer {
	s := &secretRemoteConsumer{
		ID_:              args.ID,
		CurrentRevision_: args.CurrentRevision,
	}
	if args.Consumer != nil {
		s.Consumer_ = args.Consumer.String()
	}
	return s
}

// ID implements SecretRemoteConsumer.
func (i *secretRemoteConsumer) ID() string {
	return i.ID_
}

// Consumer implements SecretRemoteConsumer.
func (i *secretRemoteConsumer) Consumer() (names.Tag, error) {
	if i.Consumer_ == "" {
		return nil, nil
	}
	tag, err := names.ParseTag(i.Consumer_)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return tag, nil
}

// CurrentRevision implements SecretRemoteConsumer.
func (i *secretRemoteConsumer) CurrentRevision() int {
	return i.CurrentRevision_
}

// LatestRevision implements SecretRemoteConsumer.
func (i *secretRemoteConsumer) LatestRevision() int {
	return i.LatestRevision_
}

func importSecretRemoteConsumers(source map[string]interface{}, version int) ([]*secretRemoteConsumer, error) {
	importFunc, ok := secretRemoteConsumerDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList, ok := source["remote-consumers"].([]interface{})
	if !ok {
		return nil, nil
	}
	return importSecretRemoteConsumersList(sourceList, importFunc)
}

func importSecretRemoteConsumersList(sourceList []interface{}, importFunc secretRemoteConsumerDeserializationFunc) ([]*secretRemoteConsumer, error) {
	result := make([]*secretRemoteConsumer, 0, len(sourceList))
	for i, consumer := range sourceList {
		source, ok := consumer.(map[interface{}]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for consumer %d, %T", i, consumer)
		}
		consumer, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "consumer %v", i)
		}
		result = append(result, consumer)
	}
	return result, nil
}

type secretRemoteConsumerDeserializationFunc func(map[interface{}]interface{}) (*secretRemoteConsumer, error)

var secretRemoteConsumerDeserializationFuncs = map[int]secretRemoteConsumerDeserializationFunc{
	1: importSecretRemoteConsumerV1,
	2: importSecretRemoteConsumerV2,
}

func importSecretRemoteConsumerV2(source map[interface{}]interface{}) (*secretRemoteConsumer, error) {
	return importSecretRemoteConsumerV1(source)
}

func importSecretRemoteConsumerV1(source map[interface{}]interface{}) (*secretRemoteConsumer, error) {
	fields := schema.Fields{
		"id":               schema.String(),
		"consumer":         schema.String(),
		"current-revision": schema.Int(),
	}
	checker := schema.FieldMap(fields, schema.Defaults{})

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "secret remote consumers v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	consumer := &secretRemoteConsumer{
		ID_:              valid["id"].(string),
		Consumer_:        valid["consumer"].(string),
		CurrentRevision_: int(valid["current-revision"].(int64)),
	}
	return consumer, nil
}
