// Copyright 2024 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

type CharmConfigsArgs struct {
	Configs map[string]CharmConfig
}

func newCharmConfigs(args CharmConfigsArgs) *charmConfigs {
	config := make(map[string]charmConfig)
	if args.Configs != nil {
		for name, c := range args.Configs {
			config[name] = charmConfig{
				Type_:        c.Type(),
				Default_:     c.Default(),
				Description_: c.Description(),
			}
		}
	}

	return &charmConfigs{
		Version_: 1,
		Configs_: config,
	}
}

type charmConfigs struct {
	Version_ int                    `yaml:"version"`
	Configs_ map[string]charmConfig `yaml:"configs"`
}

// Configs returns the configs of the charm.
func (c *charmConfigs) Configs() map[string]CharmConfig {
	configs := make(map[string]CharmConfig)
	for i, b := range c.Configs_ {
		configs[i] = b
	}
	return configs
}

func importCharmConfigs(source map[string]interface{}) (*charmConfigs, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "charmConfigs version schema check failed")
	}

	importFunc, ok := charmConfigsDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	return importFunc(source)
}

var charmConfigsDeserializationFuncs = map[int]func(map[string]interface{}) (*charmConfigs, error){
	1: importCharmConfigsV1,
}

func importCharmConfigsV1(source map[string]interface{}) (*charmConfigs, error) {
	return importCharmConfigsVersion(source, 1)
}

func importCharmConfigsVersion(source map[string]interface{}, version int) (*charmConfigs, error) {
	fields := schema.Fields{
		"configs": schema.StringMap(schema.Any()),
	}
	defaults := schema.Defaults{
		"configs": schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "charmConfigs v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	var configs map[string]charmConfig
	if valid["configs"] != nil {
		configs = make(map[string]charmConfig)
		for name, raw := range valid["configs"].(map[string]interface{}) {
			var err error
			configs[name], err = importCharmConfig(raw)
			if err != nil {
				return nil, errors.Annotatef(err, "charmConfig %q schema check failed", name)
			}
		}
	}

	return &charmConfigs{
		Version_: 1,
		Configs_: configs,
	}, nil
}

func importCharmConfig(source interface{}) (charmConfig, error) {
	fields := schema.Fields{
		"type":        schema.String(),
		"default":     schema.Any(),
		"description": schema.String(),
	}
	defaults := schema.Defaults{
		"type":        schema.Omit,
		"default":     schema.Omit,
		"description": schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return charmConfig{}, errors.Annotate(err, "charmConfig schema check failed")
	}
	valid := coerced.(map[string]interface{})

	config := charmConfig{
		Type_:        valid["type"].(string),
		Default_:     valid["default"],
		Description_: valid["description"].(string),
	}
	return config, nil
}

type charmConfig struct {
	Type_        string      `yaml:"type"`
	Default_     interface{} `yaml:"default"`
	Description_ string      `yaml:"description"`
}

// Type returns the type of the config.
func (c charmConfig) Type() string {
	return c.Type_
}

// Default returns the default value of the config.
func (c charmConfig) Default() interface{} {
	return c.Default_
}

// Description returns the description of the config.
func (c charmConfig) Description() string {
	return c.Description_
}
