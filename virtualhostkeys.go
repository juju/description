// Copyright 2025 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"encoding/base64"

	"github.com/juju/errors"
	"github.com/juju/schema"
)

type virtualHostKeys struct {
	Version         int               `yaml:"version"`
	VirtualHostKeys []*virtualHostKey `yaml:"virtual-host-keys"`
}

type virtualHostKey struct {
	ID_      string `yaml:"id"`
	HostKey_ string `yaml:"host-key"`
}

// ID implements VirtualHostKey
func (f *virtualHostKey) ID() string {
	return f.ID_
}

// HostKey implements VirtualHostKey
func (f *virtualHostKey) HostKey() []byte {
	// Here we are explicitly throwing away any decode error. We encode
	// a byte array, so we know that the stored host key can be decoded.
	hostKey, _ := base64.StdEncoding.DecodeString(f.HostKey_)
	return hostKey
}

// Validate implements VirtualHostKey.
func (f *virtualHostKey) Validate() error {
	if f.ID_ == "" {
		return errors.NotValidf("empty id")
	}
	if len(f.HostKey_) == 0 {
		return errors.NotValidf("zero length key")
	}
	return nil
}

// VirtualHostKeyArgs is an argument struct used to add a virtual host key.
type VirtualHostKeyArgs struct {
	ID      string
	HostKey []byte
}

func newVirtualHostKey(args VirtualHostKeyArgs) *virtualHostKey {
	hostKey := base64.StdEncoding.EncodeToString(args.HostKey)
	f := &virtualHostKey{
		ID_:      args.ID,
		HostKey_: hostKey,
	}
	return f
}

func importVirtualHostKeys(source map[string]interface{}) ([]*virtualHostKey, error) {
	checker := versionedChecker("virtual-host-keys")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "virtual host keys version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	sourceList := valid["virtual-host-keys"].([]interface{})
	return importVirtualHostKeyList(sourceList, version)
}

func importVirtualHostKeyList(sourceList []interface{}, version int) ([]*virtualHostKey, error) {
	getFields, ok := virtualHostKeyFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	result := make([]*virtualHostKey, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for virtual host keys %d, %T", i, value)
		}
		virtualHostKey, err := importVirtualHostKey(source, version, getFields)
		if err != nil {
			return nil, errors.Annotatef(err, "virtual host keys %d", i)
		}
		result = append(result, virtualHostKey)
	}
	return result, nil
}

var virtualHostKeyFieldsFuncs = map[int]fieldsFunc{
	1: virtualHostKeyV1Fields,
}

func virtualHostKeyV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"id":       schema.String(),
		"host-key": schema.String(),
	}
	return fields, schema.Defaults{}
}

func importVirtualHostKey(source map[string]interface{}, importVersion int, fieldFunc func() (schema.Fields, schema.Defaults)) (*virtualHostKey, error) {
	fields, defaults := fieldFunc()
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "virtual host keys v%d schema check failed", importVersion)
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	consumer := &virtualHostKey{
		ID_:      valid["id"].(string),
		HostKey_: valid["host-key"].(string),
	}
	return consumer, nil
}
