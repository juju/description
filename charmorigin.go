// Copyright 2020 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// CharmOriginArgs is an argument struct used to add information about the
// tools the agent is using to a Machine.
type CharmOriginArgs struct {
	Source string
}

func newCharmOrigin(args CharmOriginArgs) *charmOrigin {
	return &charmOrigin{
		Version_: 1,
		Source_:  args.Source,
	}
}

// Keeping the charmOrigin with the machine code, because we hope
// that one day we will succeed in merging the unit agents with the
// machine agents.
type charmOrigin struct {
	Version_ int    `yaml:"version"`
	Source_  string `yaml:"source"`
}

// Source implements CharmOrigin.
func (a *charmOrigin) Source() string {
	return a.Source_
}

func importCharmOrigin(source map[string]interface{}) (*charmOrigin, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "charmOrigin version schema check failed")
	}

	importFunc, ok := charmOriginDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	return importFunc(source)
}

type charmOriginDeserializationFunc func(map[string]interface{}) (*charmOrigin, error)

var charmOriginDeserializationFuncs = map[int]charmOriginDeserializationFunc{
	1: importCharmOriginV1,
}

func importCharmOriginV1(source map[string]interface{}) (*charmOrigin, error) {
	fields := schema.Fields{
		"source": schema.String(),
	}
	checker := schema.FieldMap(fields, nil) // no defaults

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "charmOrigin v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	return &charmOrigin{
		Version_: 1,
		Source_:  valid["source"].(string),
	}, nil
}
