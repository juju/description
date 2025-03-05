// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"strings"

	"github.com/juju/errors"
	"github.com/juju/schema"
)

type storages struct {
	Version   int        `yaml:"version"`
	Storages_ []*storage `yaml:"storages"`
}

type storage struct {
	ID_        string `yaml:"id"`
	Kind_      string `yaml:"kind"`
	UnitOwner_ string `yaml:"unit-owner,omitempty"`
	Name_      string `yaml:"name"`

	Attachments_ []string                    `yaml:"attachments,omitempty"`
	Constraints_ *StorageInstanceConstraints `yaml:"constraints,omitempty"`
}

// StorageArgs is an argument struct used to add a storage to the Model.
type StorageArgs struct {
	ID          string
	Kind        string
	UnitOwner   string
	Name        string
	Attachments []string
	Constraints *StorageInstanceConstraints
}

func newStorage(args StorageArgs) *storage {
	s := &storage{
		ID_:          args.ID,
		Kind_:        args.Kind,
		Name_:        args.Name,
		Constraints_: args.Constraints,
		UnitOwner_:   args.UnitOwner,
	}
	for _, unit := range args.Attachments {
		s.Attachments_ = append(s.Attachments_, unit)
	}
	return s
}

// ID implements Storage.
func (s *storage) ID() string {
	return s.ID_
}

// Kind implements Storage.
func (s *storage) Kind() string {
	return s.Kind_
}

// UnitOwner implements Storage.
func (s *storage) UnitOwner() (string, bool) {
	return s.UnitOwner_, s.UnitOwner_ != ""
}

// Name implements Storage.
func (s *storage) Name() string {
	return s.Name_
}

// Attachments implements Storage.
func (s *storage) Attachments() []string {
	result := make([]string, len(s.Attachments_))
	for i, unit := range s.Attachments_ {
		result[i] = unit
	}
	return result
}

// Constraints implements Storage.
func (s *storage) Constraints() (StorageInstanceConstraints, bool) {
	if s.Constraints_ != nil {
		return *s.Constraints_, true
	}
	return StorageInstanceConstraints{}, false
}

// Validate implements Storage.
func (s *storage) Validate() error {
	if s.ID_ == "" {
		return errors.NotValidf("storage missing id")
	}
	if s.Constraints_ != nil {
		if s.Constraints_.Pool == "" {
			return errors.NotValidf("storage %q invalid empty pool name", s.ID_)
		}
		if s.Constraints_.Size == 0 {
			return errors.NotValidf("storage %q invalid zero size", s.ID_)
		}
	}
	return nil
}

func importStorages(source map[string]interface{}) ([]*storage, error) {
	checker := versionedChecker("storages")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "storages version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := storageDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["storages"].([]interface{})
	return importStorageList(sourceList, importFunc)
}

func importStorageList(sourceList []interface{}, importFunc storageDeserializationFunc) ([]*storage, error) {
	result := make([]*storage, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for storage %d, %T", i, value)
		}
		storage, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "storage %d", i)
		}
		result = append(result, storage)
	}
	return result, nil
}

type storageDeserializationFunc func(map[string]interface{}) (*storage, error)

var storageDeserializationFuncs = map[int]storageDeserializationFunc{
	1: importStorageV1,
	2: importStorageV2,
	3: importStorageV3,
	4: importStorageV4,
}

func importStorageV4(source map[string]interface{}) (*storage, error) {
	checker := schema.FieldMap(storageV4Fields())
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "storage v4 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	return newStorageFromValid(valid, 4)
}

func importStorageV3(source map[string]interface{}) (*storage, error) {
	checker := schema.FieldMap(storageV3Fields())
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "storage v3 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	return newStorageFromValid(valid, 3)
}

func importStorageV2(source map[string]interface{}) (*storage, error) {
	checker := schema.FieldMap(storageV2Fields())
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "storage v2 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	return newStorageFromValid(valid, 2)
}

func importStorageV1(source map[string]interface{}) (*storage, error) {
	checker := schema.FieldMap(storageV1Fields())
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "storage v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	return newStorageFromValid(valid, 1)
}

func newStorageFromValid(valid map[string]interface{}, version int) (*storage, error) {
	result := &storage{
		ID_:   valid["id"].(string),
		Kind_: valid["kind"].(string),
		Name_: valid["name"].(string),
	}
	if version <= 3 {
		if owner, ok := valid["owner"].(string); ok {
			unitOwner, _ := strings.CutPrefix(owner, "unit-")
			result.UnitOwner_ = unitOwner
		}
	}
	if owner, ok := valid["unit-owner"].(string); ok {
		result.UnitOwner_ = owner
	}
	if attachments, ok := valid["attachments"]; ok {
		result.Attachments_ = convertToStringSlice(attachments)
	}
	if cons, ok := valid["constraints"]; ok {
		consM := cons.(map[string]interface{})
		result.Constraints_ = &StorageInstanceConstraints{
			Pool: consM["pool"].(string),
			Size: consM["size"].(uint64),
		}
	}
	return result, nil
}

func storageV4Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := storageV3Fields()
	fields["unit-owner"] = schema.String()
	defaults["unit-owner"] = schema.Omit
	delete(fields, "owner")
	delete(defaults, "owner")
	return fields, defaults
}

func storageV3Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := storageV2Fields()
	fields["constraints"] = schema.FieldMap(
		schema.Fields{
			"pool": schema.String(),
			"size": schema.Uint(),
		},
		schema.Defaults{},
	)
	return fields, defaults
}

func storageV2Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := storageV1Fields()
	defaults["owner"] = schema.Omit
	defaults["attachments"] = schema.Omit
	return fields, defaults
}

func storageV1Fields() (schema.Fields, schema.Defaults) {
	// Normally a list would have defaults, but in this case storage
	// should always have at least one attachment.
	defaults := schema.Defaults{}
	return schema.Fields{
		"id":          schema.String(),
		"kind":        schema.String(),
		"owner":       schema.String(),
		"name":        schema.String(),
		"attachments": schema.List(schema.String()),
	}, defaults
}
