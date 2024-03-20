// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// StorageDirectiveArgs is an argument struct used to create a new internal
// storageDirective type that supports the StorageDirective interface.
type StorageDirectiveArgs struct {
	Pool  string
	Size  uint64
	Count uint64
}

func newStorageDirective(args StorageDirectiveArgs) *storageDirective {
	return &storageDirective{
		Version: 1,
		Pool_:   args.Pool,
		Size_:   args.Size,
		Count_:  args.Count,
	}
}

type storageDirective struct {
	Version int `yaml:"version"`

	Pool_  string `yaml:"pool"`
	Size_  uint64 `yaml:"size"`
	Count_ uint64 `yaml:"count"`
}

// Pool implements StorageDirective.
func (s *storageDirective) Pool() string {
	return s.Pool_
}

// Size implements StorageDirective.
func (s *storageDirective) Size() uint64 {
	return s.Size_
}

// Count implements StorageDirective.
func (s *storageDirective) Count() uint64 {
	return s.Count_
}

func importStorageDirectives(sourceMap map[string]interface{}) (map[string]*storageDirective, error) {
	result := make(map[string]*storageDirective)
	for key, value := range sourceMap {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for storage directive %q, %T", key, value)
		}
		directive, err := importStorageDirective(source)
		if err != nil {
			return nil, errors.Trace(err)
		}
		result[key] = directive
	}
	return result, nil
}

// importStorageDirective constructs a new StorageDirective from a map representing a serialised
// StorageDirective instance.
func importStorageDirective(source map[string]interface{}) (*storageDirective, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "storage directive version schema check failed")
	}

	importFunc, ok := storageconstraintDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	return importFunc(source)
}

type storageDirectiveDeserializationFunc func(map[string]interface{}) (*storageDirective, error)

var storageconstraintDeserializationFuncs = map[int]storageDirectiveDeserializationFunc{
	1: importStorageDirectiveV1,
}

func importStorageDirectiveV1(source map[string]interface{}) (*storageDirective, error) {
	fields := schema.Fields{
		"pool":  schema.String(),
		"size":  schema.Uint(),
		"count": schema.Uint(),
	}
	checker := schema.FieldMap(fields, nil)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "storage directive v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	return &storageDirective{
		Version: 1,
		Pool_:   valid["pool"].(string),
		Size_:   valid["size"].(uint64),
		Count_:  valid["count"].(uint64),
	}, nil
}
