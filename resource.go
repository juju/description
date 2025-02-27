// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"time"

	"github.com/juju/errors"
	"github.com/juju/schema"
)

// Resource represents an application resource.
type Resource interface {
	// Name returns the name of the resource.
	Name() string

	// SetApplicationRevision sets the application revision of the
	// resource.
	SetApplicationRevision(ResourceRevisionArgs) ResourceRevision

	// ApplicationRevision returns the revision of the resource as set
	// on the application. May return nil if SetApplicationRevision
	// hasn't been called yet.
	ApplicationRevision() ResourceRevision

	// SetCharmStoreRevision sets the application revision of the
	// resource.
	SetCharmStoreRevision(ResourceRevisionArgs) ResourceRevision

	// CharmStoreRevision returns the revision the charmstore has, as
	// seen at the last poll. May return nil if SetCharmStoreRevision
	// hasn't been called yet.
	CharmStoreRevision() ResourceRevision

	// Validate checks the consistency of the resource and its
	// revisions.
	Validate() error
}

// ResourceRevision represents a revision of an application resource.
type ResourceRevision interface {
	// Revision returns the resource revision, or nil if unset.
	Revision() int

	// Type returns the resource type.
	Type() string

	// Origin returns the resource origin.
	Origin() string

	// SHA384 returns the SHA384 hash of the resource blob.
	SHA384() string

	// Size returns the size of the resource blob.
	Size() int64

	// Timestamp returns the time the blob associated with this resource was
	// added.
	Timestamp() time.Time

	// RetrievedBy returns the name of the entity that retrieved this
	// resource.
	RetrievedBy() string
}

// ResourceArgs is an argument struct used to create a new internal
// resource type that supports the Resource interface.
type ResourceArgs struct {
	// Name is the name of the resource.
	Name string
}

// newResource returns a new *resource (which implements the Resource
// interface).
func newResource(args ResourceArgs) *resource {
	return &resource{
		Name_: args.Name,
	}
}

type resources struct {
	Version    int         `yaml:"version"`
	Resources_ []*resource `yaml:"resources"`
}

type resource struct {
	Name_                string            `yaml:"name"`
	ApplicationRevision_ *resourceRevision `yaml:"application-revision"`
	CharmStoreRevision_  *resourceRevision `yaml:"charmstore-revision,omitempty"`
}

// ResourceRevisionArgs is an argument struct used to add a new
// internal resource revision to a Resource.
type ResourceRevisionArgs struct {
	// Revision is the resource revision, or nil if origin is upload.
	Revision int

	// Type is the resource type.
	Type string

	// Origin is the resource origin.
	Origin string

	// SHA384 is the hash of the blob associated with this resource.
	SHA384 string

	// Size is the size in bytes of the blob associated with this resource.
	Size int64

	// Timestamp is the time the blob associated with this resource was
	// added.
	Timestamp time.Time

	// RetrievedBy is the name of the entity that retrieved this resource.
	RetrievedBy string
}

// Name implements Resource.
func (r *resource) Name() string {
	return r.Name_
}

// SetApplicationRevision implements Resource.
func (r *resource) SetApplicationRevision(args ResourceRevisionArgs) ResourceRevision {
	r.ApplicationRevision_ = newResourceRevision(args)
	return r.ApplicationRevision_
}

// ApplicationRevision implements Resource.
func (r *resource) ApplicationRevision() ResourceRevision {
	if r.ApplicationRevision_ == nil {
		return nil // Return untyped nil when not set
	}
	return r.ApplicationRevision_
}

// SetCharmStoreRevision implements Resource.
func (r *resource) SetCharmStoreRevision(args ResourceRevisionArgs) ResourceRevision {
	r.CharmStoreRevision_ = newResourceRevision(args)
	return r.CharmStoreRevision_
}

// CharmStoreRevision implements Resource.
func (r *resource) CharmStoreRevision() ResourceRevision {
	if r.CharmStoreRevision_ == nil {
		return nil // Return untyped nil when not set
	}
	return r.CharmStoreRevision_
}

// Validate implements Resource.
func (r *resource) Validate() error {
	if r.ApplicationRevision_ == nil {
		return errors.New("no application revision set")
	}
	return nil
}

func newResourceRevision(args ResourceRevisionArgs) *resourceRevision {
	return &resourceRevision{
		Revision_:    args.Revision,
		Type_:        args.Type,
		Origin_:      args.Origin,
		SHA384_:      args.SHA384,
		Size_:        args.Size,
		Timestamp_:   timePtr(args.Timestamp),
		RetrievedBy_: args.RetrievedBy,
	}
}

type resourceRevision struct {
	Revision_    int        `yaml:"revision"`
	Type_        string     `yaml:"type"`
	Origin_      string     `yaml:"origin"`
	SHA384_      string     `yaml:"sha384"`
	Size_        int64      `yaml:"size"`
	Timestamp_   *time.Time `yaml:"timestamp,omitempty"`
	RetrievedBy_ string     `yaml:"retrieved-by,omitempty"`
}

// Revision implements ResourceRevision.
func (r *resourceRevision) Revision() int {
	return r.Revision_
}

// Type implements ResourceRevision.
func (r *resourceRevision) Type() string {
	return r.Type_
}

// Origin implements ResourceRevision.
func (r *resourceRevision) Origin() string {
	return r.Origin_
}

// SHA384 implements ResourceRevision.
func (r *resourceRevision) SHA384() string {
	return r.SHA384_
}

// Size implements ResourceRevision.
func (r *resourceRevision) Size() int64 {
	return r.Size_
}

// Timestamp implements ResourceRevision.
func (r *resourceRevision) Timestamp() time.Time {
	if r.Timestamp_ == nil {
		return time.Time{}
	}
	return *r.Timestamp_
}

// RetrievedBy implements ResourceRevision.
func (r *resourceRevision) RetrievedBy() string {
	return r.RetrievedBy_
}

func importResources(source map[string]interface{}) ([]*resource, error) {
	checker := versionedChecker("resources")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotate(err, "resources version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := resourceDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["resources"].([]interface{})
	return importResourceList(sourceList, importFunc)
}

func importResourceList(sourceList []interface{}, importFunc resourceDeserializationFunc) ([]*resource, error) {
	result := make([]*resource, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for resource %d, %T", i, value)
		}
		resource, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "resource %d", i)
		}
		result = append(result, resource)
	}
	return result, nil
}

type resourceDeserializationFunc func(map[string]interface{}) (*resource, error)

var resourceDeserializationFuncs = map[int]resourceDeserializationFunc{
	1: importResourceV1,
	2: importResourceV2,
}

func importResourceV1(source map[string]interface{}) (*resource, error) {
	return importResource(source, importResourceRevisionV1)
}

func importResourceV2(source map[string]interface{}) (*resource, error) {
	return importResource(source, importResourceRevisionV2)
}

func importResource(
	source map[string]interface{},
	importRevisionFunc func(source interface{}) (*resourceRevision, error),
) (*resource, error) {
	fields := schema.Fields{
		"name":                 schema.String(),
		"application-revision": schema.StringMap(schema.Any()),
		"charmstore-revision":  schema.StringMap(schema.Any()),
	}
	defaults := schema.Defaults{
		"charmstore-revision": schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "resource schema check failed")
	}
	valid := coerced.(map[string]interface{})

	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	r := newResource(ResourceArgs{
		Name: valid["name"].(string),
	})
	appRev, err := importRevisionFunc(valid["application-revision"])
	if err != nil {
		return nil, errors.Annotatef(err, "resource %s: application revision", r.Name_)
	}
	r.ApplicationRevision_ = appRev
	if source, exists := valid["charmstore-revision"]; exists {
		csRev, err := importRevisionFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "resource %s: charmstore revision", r.Name_)
		}
		r.CharmStoreRevision_ = csRev
	}
	return r, nil
}

func importResourceRevisionV2(source interface{}) (*resourceRevision, error) {
	fields := schema.Fields{
		"revision":     schema.Int(),
		"type":         schema.String(),
		"origin":       schema.String(),
		"sha384":       schema.String(),
		"size":         schema.Int(),
		"timestamp":    schema.Time(),
		"retrieved-by": schema.String(),
	}
	defaults := schema.Defaults{
		"timestamp":    schema.Omit,
		"retrieved-by": "",
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "resource v2 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	r := &resourceRevision{
		Type_:        valid["type"].(string),
		Origin_:      valid["origin"].(string),
		SHA384_:      valid["sha384"].(string),
		Size_:        valid["size"].(int64),
		Timestamp_:   fieldToTimePtr(valid, "timestamp"),
		RetrievedBy_: valid["retrieved-by"].(string),
		Revision_:    int(valid["revision"].(int64)),
	}

	return r, nil
}

func importResourceRevisionV1(source interface{}) (*resourceRevision, error) {
	fields := schema.Fields{
		"revision":    schema.Int(),
		"type":        schema.String(),
		"path":        schema.String(),
		"description": schema.String(),
		"origin":      schema.String(),
		"fingerprint": schema.String(),
		"size":        schema.Int(),
		"timestamp":   schema.Time(),
		"username":    schema.String(),
	}
	defaults := schema.Defaults{
		"timestamp": schema.Omit,
		"username":  "",
	}
	checker := schema.FieldMap(fields, defaults)
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "resource revision schema check failed")
	}
	valid := coerced.(map[string]interface{})

	rev := &resourceRevision{
		Revision_:    int(valid["revision"].(int64)),
		Type_:        valid["type"].(string),
		Origin_:      valid["origin"].(string),
		SHA384_:      valid["fingerprint"].(string),
		Size_:        valid["size"].(int64),
		Timestamp_:   fieldToTimePtr(valid, "timestamp"),
		RetrievedBy_: valid["username"].(string),
	}
	return rev, nil
}
