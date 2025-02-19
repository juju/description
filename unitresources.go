// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// UnitResource represents the revision of a resource used by a unit.
type UnitResource interface {
	// Name returns the name of the resource.
	Name() string

	// Revision returns the revision of the resource as used by a
	// particular unit.
	Revision() ResourceRevision
}

// UnitResourceArgs is an argument struct used to specify the revision
// of a resource used by a unit.
type UnitResourceArgs struct {
	Name         string
	RevisionArgs ResourceRevisionArgs
}

func newUnitResource(args UnitResourceArgs) *unitResource {
	return &unitResource{
		Name_:     args.Name,
		Revision_: newResourceRevision(args.RevisionArgs),
	}
}

type unitResources struct {
	Version    int             `yaml:"version"`
	Resources_ []*unitResource `yaml:"resources"`
}

type unitResource struct {
	Name_     string            `yaml:"name"`
	Revision_ *resourceRevision `yaml:"revision"`
}

// Name implements UnitResource.
func (ur *unitResource) Name() string {
	return ur.Name_
}

// Revision implements UnitResource.
func (ur *unitResource) Revision() ResourceRevision {
	return ur.Revision_
}

type unitResourceDeserializationFunc func(map[string]interface{}) (*unitResource, error)

var unitResourceDeserializationFuncs = map[int]unitResourceDeserializationFunc{
	1: importUnitResourceV1,
	2: importUnitResourceV2,
}

func importUnitResources(source map[string]interface{}) ([]*unitResource, error) {
	checker := versionedChecker("resources")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "unit resources version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := unitResourceDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	sourceList := valid["resources"].([]interface{})
	return importUnitResourceList(sourceList, importFunc)
}

func importUnitResourceList(sourceList []interface{}, importFunc unitResourceDeserializationFunc) ([]*unitResource, error) {
	result := make([]*unitResource, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for resource %d, %T", i, value)
		}
		r, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "unit resource %d", i)
		}
		result = append(result, r)
	}
	return result, nil
}

func importUnitResourceV1(source map[string]interface{}) (*unitResource, error) {
	fields := schema.Fields{
		"name":     schema.String(),
		"revision": schema.StringMap(schema.Any()),
	}
	checker := schema.FieldMap(fields, nil)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "unit resource v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	r := &unitResource{
		Name_: valid["name"].(string),
	}
	rev, err := importResourceRevisionV1(valid["revision"])
	if err != nil {
		return nil, errors.Annotatef(err, "unit resource %s", r.Name_)
	}
	r.Revision_ = rev

	return r, nil
}

func importUnitResourceV2(source map[string]interface{}) (*unitResource, error) {
	fields := schema.Fields{
		"name":     schema.String(),
		"revision": schema.StringMap(schema.Any()),
	}
	checker := schema.FieldMap(fields, nil)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "unit resource v2 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	r := &unitResource{
		Name_: valid["name"].(string),
	}
	rev, err := importResourceRevisionV2(valid["revision"])
	if err != nil {
		return nil, errors.Annotatef(err, "unit resource %s", r.Name_)
	}
	r.Revision_ = rev

	return r, nil
}
