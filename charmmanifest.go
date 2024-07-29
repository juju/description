// Copyright 2024 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// CharmManifestArgs is an argument struct used to create a new
// CharmManifest.
type CharmManifestArgs struct {
	Bases []CharmManifestBase
}

func newCharmManifest(args CharmManifestArgs) *charmManifest {
	var bases []charmManifestBase
	if args.Bases != nil {
		bases = make([]charmManifestBase, len(args.Bases))
		for i, b := range args.Bases {
			bases[i] = charmManifestBase{
				Name_:          b.Name(),
				Channel_:       b.Channel(),
				Architectures_: b.Architectures(),
			}
		}
	}

	return &charmManifest{
		Version_: 1,
		Bases_:   bases,
	}
}

// charmManifest represents the metadata of a charm.
type charmManifest struct {
	Version_ int                 `yaml:"version"`
	Bases_   []charmManifestBase `yaml:"bases"`
}

// Bases returns the list of the base the charm supports.
func (m *charmManifest) Bases() []CharmManifestBase {
	bases := make([]CharmManifestBase, len(m.Bases_))
	for i, b := range m.Bases_ {
		bases[i] = b
	}
	return bases
}

func importCharmManifest(source map[string]interface{}) (*charmManifest, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "charmManifest version schema check failed")
	}

	importFunc, ok := charmManifestDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	return importFunc(source)
}

type charmManifestDeserializationFunc func(map[string]interface{}) (*charmManifest, error)

var charmManifestDeserializationFuncs = map[int]charmManifestDeserializationFunc{
	1: importCharmManifestV1,
}

func importCharmManifestV1(source map[string]interface{}) (*charmManifest, error) {
	return importCharmManifestVersion(source, 1)
}

func importCharmManifestVersion(source map[string]interface{}, importVersion int) (*charmManifest, error) {
	fields := schema.Fields{
		"bases": schema.List(schema.Any()),
	}
	defaults := schema.Defaults{
		"bases": schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "charmManifest v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	var bases []charmManifestBase
	if valid["bases"] != nil {
		bases = make([]charmManifestBase, len(valid["bases"].([]interface{})))
		for k, v := range valid["bases"].([]interface{}) {
			var err error
			bases[k], err = importCharmManifestBase(v, importVersion)
			if err != nil {
				return nil, errors.Annotate(err, "charmManifest bases schema check failed")
			}
		}
	}

	return &charmManifest{
		Version_: 1,
		Bases_:   bases,
	}, nil
}

func importCharmManifestBase(source interface{}, importVersion int) (charmManifestBase, error) {
	fields := schema.Fields{
		"name":          schema.String(),
		"channel":       schema.String(),
		"architectures": schema.List(schema.String()),
	}
	defaults := schema.Defaults{
		"name":          schema.Omit,
		"channel":       schema.Omit,
		"architectures": schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return charmManifestBase{}, errors.Annotate(err, "charmManifestBase schema check failed")
	}
	valid := coerced.(map[string]interface{})

	var architectures []string
	if valid["architectures"] != nil {
		architectures = make([]string, len(valid["architectures"].([]interface{})))
		for i, a := range valid["architectures"].([]interface{}) {
			architectures[i] = a.(string)
		}
	}

	return charmManifestBase{
		Name_:          valid["name"].(string),
		Channel_:       valid["channel"].(string),
		Architectures_: architectures,
	}, nil
}

type charmManifestBase struct {
	Name_          string   `yaml:"name"`
	Channel_       string   `yaml:"channel"`
	Architectures_ []string `yaml:"architectures"`
}

// Name returns the name of the base.
func (r charmManifestBase) Name() string {
	return r.Name_
}

// Channel returns the channel of the base.
func (r charmManifestBase) Channel() string {
	return r.Channel_
}

// Architectures returns the architectures of the base.
func (r charmManifestBase) Architectures() []string {
	return r.Architectures_
}
