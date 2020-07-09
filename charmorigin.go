// Copyright 2020 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// CharmOrigin represents a charm origin.
type CharmOrigin interface {
	Source() string
}

var _ CharmOrigin = (*charmOrigin)(nil)

type charmOrigins struct {
	Version      int          `yaml:"version"`
	CharmOrigin_ *charmOrigin `yaml:"charm-origin,omitempty"`
}

type charmOrigin struct {
	Source_ string `yaml:"source,omitempty"`
}

// Source returns the underlying charm origin source.
func (o *charmOrigin) Source() string {
	return o.Source_
}

// CharmOriginArgs is an argument struct used to instantiate a new
// charmOrigin instance that implements charmOrigin.
type CharmOriginArgs struct {
	Source string
}

func newCharmOrigin(args CharmOriginArgs) *charmOrigin {
	return &charmOrigin{
		Source_: args.Source,
	}
}

func importCharmOrigins(source map[string]interface{}) (*charmOrigin, error) {
	checker := versionedChecker("charm-origin")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "charm-origin version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := charmOriginDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	charmOrigin := valid["charm-origin"].(map[string]interface{})
	return importFunc(charmOrigin)
}

type charmOriginDeserializationFunc func(interface{}) (*charmOrigin, error)

var charmOriginDeserializationFuncs = map[int]charmOriginDeserializationFunc{
	1: importCharmOriginV1,
}

func importCharmOriginV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"source": schema.String(),
	}
	return fields, schema.Defaults{}
}

func importCharmOrigin(fields schema.Fields, defaults schema.Defaults, importVersion int, source interface{}) (*charmOrigin, error) {
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "application offer v%d schema check failed", importVersion)
	}
	valid := coerced.(map[string]interface{})

	origin := &charmOrigin{
		Source_: valid["source"].(string),
	}

	return origin, nil
}

func importCharmOriginV1(source interface{}) (*charmOrigin, error) {
	fields, defaults := importCharmOriginV1Fields()
	return importCharmOrigin(fields, defaults, 1, source)
}
