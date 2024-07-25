// Copyright 2020 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// CharmMetadataArgs is an argument struct used to create a new
// CharmMetadata.
type CharmMetadataArgs struct {
	Name           string
	Summary        string
	Description    string
	Subordinate    bool
	MinJujuVersion string
	RunAs          string
	Assumes        string
	Relations      map[string]CharmMetadataRelation
	ExtraBindings  map[string]string
	Categories     []string
	Tags           []string
	Storage        map[string]CharmMetadataStorage
	Devices        map[string]CharmMetadataDevice
	Payloads       map[string]CharmMetadataPayload
	Resources      map[string]CharmMetadataResource
	Terms          []string
	Containers     map[string]CharmMetadataContainer
}

func newCharmMetadata(args CharmMetadataArgs) *charmMetadata {
	var relations map[string]charmMetadataRelation
	if args.Relations != nil {
		relations = make(map[string]charmMetadataRelation, len(args.Relations))
		for k, v := range args.Relations {
			relations[k] = charmMetadataRelation{
				Name_:      v.Name(),
				Role_:      v.Role(),
				Interface_: v.Interface(),
				Optional_:  v.Optional(),
				Limit_:     v.Limit(),
				Scope_:     v.Scope(),
			}
		}
	}

	var storage map[string]charmMetadataStorage
	if args.Storage != nil {
		storage = make(map[string]charmMetadataStorage, len(args.Storage))
		for k, v := range args.Storage {
			storage[k] = charmMetadataStorage{
				Name_:        v.Name(),
				Description_: v.Description(),
				Type_:        v.Type(),
				Shared_:      v.Shared(),
				Readonly_:    v.Readonly(),
				CountMin_:    v.CountMin(),
				CountMax_:    v.CountMax(),
				MinimumSize_: v.MinimumSize(),
				Location_:    v.Location(),
				Properties_:  v.Properties(),
			}
		}
	}

	var devices map[string]charmMetadataDevice
	if args.Devices != nil {
		devices = make(map[string]charmMetadataDevice, len(args.Devices))
		for k, v := range args.Devices {
			devices[k] = charmMetadataDevice{
				Name_:        v.Name(),
				Description_: v.Description(),
				Type_:        v.Type(),
				CountMin_:    v.CountMin(),
				CountMax_:    v.CountMax(),
			}
		}
	}

	var payloads map[string]charmMetadataPayload
	if args.Payloads != nil {
		payloads = make(map[string]charmMetadataPayload, len(args.Payloads))
		for k, v := range args.Payloads {
			payloads[k] = charmMetadataPayload{
				Name_: v.Name(),
				Type_: v.Type(),
			}
		}
	}

	var resources map[string]charmMetadataResource
	if args.Resources != nil {
		resources = make(map[string]charmMetadataResource, len(args.Resources))
		for k, v := range args.Resources {
			resources[k] = charmMetadataResource{
				Name_:        v.Name(),
				Type_:        v.Type(),
				Path_:        v.Path(),
				Description_: v.Description(),
			}
		}
	}

	var containers map[string]charmMetadataContainer
	if args.Containers != nil {
		containers = make(map[string]charmMetadataContainer, len(args.Containers))
		for k, v := range args.Containers {
			mounts := make([]charmMetadataContainerMount, len(v.Mounts()))
			for i, m := range v.Mounts() {
				mounts[i] = charmMetadataContainerMount{
					Storage_:  m.Storage(),
					Location_: m.Location(),
				}
			}
			containers[k] = charmMetadataContainer{
				Resource_: v.Resource(),
				Mounts_:   mounts,
				Uid_:      v.Uid(),
				Gid_:      v.Gid(),
			}
		}
	}

	return &charmMetadata{
		Version_:        1,
		Name_:           args.Name,
		Summary_:        args.Summary,
		Description_:    args.Description,
		Subordinate_:    args.Subordinate,
		MinJujuVersion_: args.MinJujuVersion,
		RunAs_:          args.RunAs,
		Assumes_:        args.Assumes,
		Relations_:      relations,
		ExtraBindings_:  args.ExtraBindings,
		Categories_:     args.Categories,
		Tags_:           args.Tags,
		Storage_:        storage,
		Devices_:        devices,
		Payloads_:       payloads,
		Resources_:      resources,
		Terms_:          args.Terms,
		Containers_:     containers,
	}
}

// charmMetadata represents the metadata of a charm.
type charmMetadata struct {
	Version_        int                               `yaml:"version"`
	Name_           string                            `yaml:"name,omitempty"`
	Summary_        string                            `yaml:"summary,omitempty"`
	Description_    string                            `yaml:"description,omitempty"`
	Subordinate_    bool                              `yaml:"subordinate,omitempty"`
	MinJujuVersion_ string                            `yaml:"min-juju-version,omitempty"`
	RunAs_          string                            `yaml:"run-as,omitempty"`
	Assumes_        string                            `yaml:"assumes,omitempty"`
	Relations_      map[string]charmMetadataRelation  `yaml:"relations,omitempty"`
	ExtraBindings_  map[string]string                 `yaml:"extra-bindings,omitempty"`
	Categories_     []string                          `yaml:"categories,omitempty"`
	Tags_           []string                          `yaml:"tags,omitempty"`
	Storage_        map[string]charmMetadataStorage   `yaml:"storage,omitempty"`
	Devices_        map[string]charmMetadataDevice    `yaml:"devices,omitempty"`
	Payloads_       map[string]charmMetadataPayload   `yaml:"payloads,omitempty"`
	Resources_      map[string]charmMetadataResource  `yaml:"resources,omitempty"`
	Terms_          []string                          `yaml:"terms,omitempty"`
	Containers_     map[string]charmMetadataContainer `yaml:"containers,omitempty"`
}

// Name returns the name of the charm.
func (m *charmMetadata) Name() string {
	return m.Name_
}

// Summary returns the summary of the charm.
func (m *charmMetadata) Summary() string {
	return m.Summary_
}

// Description returns the description of the charm.
func (m *charmMetadata) Description() string {
	return m.Description_
}

// Subordinate returns whether the charm is a subordinate charm.
func (m *charmMetadata) Subordinate() bool {
	return m.Subordinate_
}

// MinJujuVersion returns the minimum Juju version required by the charm.
func (m *charmMetadata) MinJujuVersion() string {
	return m.MinJujuVersion_
}

// RunAs returns the user the charm should run as.
func (m *charmMetadata) RunAs() string {
	return m.RunAs_
}

// Assumes returns the charm the charm assumes.
func (m *charmMetadata) Assumes() string {
	return m.Assumes_
}

// Relations returns the relations of the charm.
func (m *charmMetadata) Relations() map[string]CharmMetadataRelation {
	relations := make(map[string]CharmMetadataRelation, len(m.Relations_))
	for k, v := range m.Relations_ {
		relations[k] = v
	}
	return relations
}

// ExtraBindings returns the extra bindings of the charm.
func (m *charmMetadata) ExtraBindings() map[string]string {
	return m.ExtraBindings_
}

// Categories returns the categories of the charm.
func (m *charmMetadata) Categories() []string {
	return m.Categories_
}

// Tags returns the tags of the charm.
func (m *charmMetadata) Tags() []string {
	return m.Tags_
}

// Storage returns the storage of the charm.
func (m *charmMetadata) Storage() map[string]CharmMetadataStorage {
	storage := make(map[string]CharmMetadataStorage, len(m.Storage_))
	for k, v := range m.Storage_ {
		storage[k] = v
	}
	return storage
}

// Devices returns the devices of the charm.
func (m *charmMetadata) Devices() map[string]CharmMetadataDevice {
	devices := make(map[string]CharmMetadataDevice, len(m.Devices_))
	for k, v := range m.Devices_ {
		devices[k] = v
	}
	return devices
}

// Payloads returns the payloads of the charm.
func (m *charmMetadata) Payloads() map[string]CharmMetadataPayload {
	payloads := make(map[string]CharmMetadataPayload, len(m.Payloads_))
	for k, v := range m.Payloads_ {
		payloads[k] = v
	}
	return payloads
}

// Resources returns the resources of the charm.
func (m *charmMetadata) Resources() map[string]CharmMetadataResource {
	resources := make(map[string]CharmMetadataResource, len(m.Resources_))
	for k, v := range m.Resources_ {
		resources[k] = v
	}
	return resources
}

// Terms returns the terms of the charm.
func (m *charmMetadata) Terms() []string {
	return m.Terms_
}

// Containers returns the containers of the charm.
func (m *charmMetadata) Containers() map[string]CharmMetadataContainer {
	containers := make(map[string]CharmMetadataContainer, len(m.Containers_))
	for k, v := range m.Containers_ {
		containers[k] = v
	}
	return containers
}

func importCharmMetadata(source map[string]interface{}) (*charmMetadata, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "charmMetadata version schema check failed")
	}

	importFunc, ok := charmMetadataDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	return importFunc(source)
}

type charmMetadataDeserializationFunc func(map[string]interface{}) (*charmMetadata, error)

var charmMetadataDeserializationFuncs = map[int]charmMetadataDeserializationFunc{
	1: importCharmMetadataV1,
}

func importCharmMetadataV1(source map[string]interface{}) (*charmMetadata, error) {
	return importCharmMetadataVersion(source, 1)
}

func importCharmMetadataVersion(source map[string]interface{}, importVersion int) (*charmMetadata, error) {
	fields := schema.Fields{
		"name":             schema.String(),
		"summary":          schema.String(),
		"description":      schema.String(),
		"subordinate":      schema.Bool(),
		"min-juju-version": schema.String(),
		"run-as":           schema.String(),
		"assumes":          schema.String(),
		"relations":        schema.StringMap(schema.Any()),
		"extra-bindings":   schema.StringMap(schema.String()),
		"categories":       schema.List(schema.String()),
		"tags":             schema.List(schema.String()),
		"storage":          schema.StringMap(schema.Any()),
		"devices":          schema.StringMap(schema.Any()),
		"payloads":         schema.StringMap(schema.Any()),
		"resources":        schema.StringMap(schema.Any()),
		"terms":            schema.List(schema.String()),
		"containers":       schema.StringMap(schema.Any()),
	}
	defaults := schema.Defaults{
		"summary":          schema.Omit,
		"description":      schema.Omit,
		"subordinate":      schema.Omit,
		"min-juju-version": schema.Omit,
		"run-as":           schema.Omit,
		"assumes":          schema.Omit,
		"relations":        schema.Omit,
		"extra-bindings":   schema.Omit,
		"categories":       schema.Omit,
		"tags":             schema.Omit,
		"storage":          schema.Omit,
		"devices":          schema.Omit,
		"payloads":         schema.Omit,
		"resources":        schema.Omit,
		"terms":            schema.Omit,
		"containers":       schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "charmOrigin v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	var relations map[string]charmMetadataRelation
	if valid["relations"] != nil {
		relations = make(map[string]charmMetadataRelation)
		for k, v := range valid["relations"].(map[string]interface{}) {
			var err error
			if relations[k], err = importCharmMetadataRelation(v, importVersion); err != nil {
				return nil, errors.Annotatef(err, "relation %q", k)
			}
		}
	}

	var extraBindings map[string]string
	if valid["extra-bindings"] != nil {
		extraBindings = make(map[string]string)
		for k, v := range valid["extra-bindings"].(map[string]interface{}) {
			extraBindings[k] = v.(string)
		}
	}

	var categories []string
	if valid["categories"] != nil {
		categories = make([]string, 0)
		for _, v := range valid["categories"].([]interface{}) {
			categories = append(categories, v.(string))
		}
	}

	var tags []string
	if valid["tags"] != nil {
		tags = make([]string, 0)
		for _, v := range valid["tags"].([]interface{}) {
			tags = append(tags, v.(string))
		}
	}

	var storage map[string]charmMetadataStorage
	if valid["storage"] != nil {
		storage = make(map[string]charmMetadataStorage)
		for k, v := range valid["storage"].(map[string]interface{}) {
			var err error
			if storage[k], err = importCharmMetadataStorage(v, importVersion); err != nil {
				return nil, errors.Annotatef(err, "storage %q", k)
			}
		}
	}

	var devices map[string]charmMetadataDevice
	if valid["devices"] != nil {
		devices = make(map[string]charmMetadataDevice)
		for k, v := range valid["devices"].(map[string]interface{}) {
			var err error
			if devices[k], err = importCharmMetadataDevice(v, importVersion); err != nil {
				return nil, errors.Annotatef(err, "device %q", k)
			}
		}
	}

	var payloads map[string]charmMetadataPayload
	if valid["payloads"] != nil {
		payloads = make(map[string]charmMetadataPayload)
		for k, v := range valid["payloads"].(map[string]interface{}) {
			var err error
			if payloads[k], err = importCharmMetadataPayload(v, importVersion); err != nil {
				return nil, errors.Annotatef(err, "payload %q", k)
			}
		}
	}

	var resources map[string]charmMetadataResource
	if valid["resources"] != nil {
		resources = make(map[string]charmMetadataResource)
		for k, v := range valid["resources"].(map[string]interface{}) {
			var err error
			if resources[k], err = importCharmMetadataResource(v, importVersion); err != nil {
				return nil, errors.Annotatef(err, "resource %q", k)
			}
		}
	}

	var containers map[string]charmMetadataContainer
	if valid["containers"] != nil {
		containers = make(map[string]charmMetadataContainer)
		for k, v := range valid["containers"].(map[string]interface{}) {
			var err error
			if containers[k], err = importCharmMetadataContainer(v, importVersion); err != nil {
				return nil, errors.Annotatef(err, "container %q", k)
			}
		}
	}

	var terms []string
	if valid["terms"] != nil {
		terms = make([]string, 0)
		for _, v := range valid["terms"].([]interface{}) {
			terms = append(terms, v.(string))
		}
	}

	var (
		summary        string
		description    string
		subordinate    bool
		minJujuVersion string
		runAs          string
		assumes        string
	)

	if valid["summary"] != nil {
		summary = valid["summary"].(string)
	}
	if valid["description"] != nil {
		description = valid["description"].(string)
	}
	if valid["subordinate"] != nil {
		subordinate = valid["subordinate"].(bool)
	}
	if valid["min-juju-version"] != nil {
		minJujuVersion = valid["min-juju-version"].(string)
	}
	if valid["run-as"] != nil {
		runAs = valid["run-as"].(string)
	}
	if valid["assumes"] != nil {
		assumes = valid["assumes"].(string)
	}

	return &charmMetadata{
		Version_:        1,
		Name_:           valid["name"].(string),
		Summary_:        summary,
		Description_:    description,
		Subordinate_:    subordinate,
		MinJujuVersion_: minJujuVersion,
		RunAs_:          runAs,
		Assumes_:        assumes,
		Relations_:      relations,
		ExtraBindings_:  extraBindings,
		Categories_:     categories,
		Tags_:           tags,
		Storage_:        storage,
		Devices_:        devices,
		Resources_:      resources,
		Payloads_:       payloads,
		Containers_:     containers,
		Terms_:          terms,
	}, nil
}

func importCharmMetadataRelation(source interface{}, importVersion int) (charmMetadataRelation, error) {
	fields := schema.Fields{
		"name":      schema.String(),
		"role":      schema.String(),
		"interface": schema.String(),
		"optional":  schema.Bool(),
		"limit":     schema.Int(),
		"scope":     schema.String(),
	}
	defaults := schema.Defaults{
		"optional": schema.Omit,
		"limit":    schema.Omit,
		"scope":    schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return charmMetadataRelation{}, errors.Annotate(err, "charmMetadataRelation schema check failed")
	}
	valid := coerced.(map[string]interface{})

	return charmMetadataRelation{
		Name_:      valid["name"].(string),
		Role_:      valid["role"].(string),
		Interface_: valid["interface"].(string),
		Optional_:  valid["optional"].(bool),
		Limit_:     int(valid["limit"].(int64)),
		Scope_:     valid["scope"].(string),
	}, nil
}

func importCharmMetadataStorage(source interface{}, importVersion int) (charmMetadataStorage, error) {
	fields := schema.Fields{
		"name":         schema.String(),
		"description":  schema.String(),
		"type":         schema.String(),
		"shared":       schema.Bool(),
		"readonly":     schema.Bool(),
		"count-min":    schema.Int(),
		"count-max":    schema.Int(),
		"minimum-size": schema.Int(),
		"location":     schema.String(),
		"properties":   schema.List(schema.String()),
	}
	defaults := schema.Defaults{
		"description":  schema.Omit,
		"shared":       schema.Omit,
		"readonly":     schema.Omit,
		"count-min":    schema.Omit,
		"count-max":    schema.Omit,
		"minimum-size": schema.Omit,
		"location":     schema.Omit,
		"properties":   schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return charmMetadataStorage{}, errors.Annotate(err, "charmMetadataStorage schema check failed")
	}
	valid := coerced.(map[string]interface{})

	properties := make([]string, 0)
	for _, v := range valid["properties"].([]interface{}) {
		properties = append(properties, v.(string))
	}

	return charmMetadataStorage{
		Name_:        valid["name"].(string),
		Description_: valid["description"].(string),
		Type_:        valid["type"].(string),
		Shared_:      valid["shared"].(bool),
		Readonly_:    valid["readonly"].(bool),
		CountMin_:    int(valid["count-min"].(int64)),
		CountMax_:    int(valid["count-max"].(int64)),
		MinimumSize_: int(valid["minimum-size"].(int64)),
		Location_:    valid["location"].(string),
		Properties_:  properties,
	}, nil
}

func importCharmMetadataDevice(source interface{}, importVersion int) (charmMetadataDevice, error) {
	fields := schema.Fields{
		"name":        schema.String(),
		"description": schema.String(),
		"type":        schema.String(),
		"count-min":   schema.Int(),
		"count-max":   schema.Int(),
	}
	defaults := schema.Defaults{
		"description": schema.Omit,
		"count-min":   schema.Omit,
		"count-max":   schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return charmMetadataDevice{}, errors.Annotate(err, "charmMetadataDevice schema check failed")
	}
	valid := coerced.(map[string]interface{})

	return charmMetadataDevice{
		Name_:        valid["name"].(string),
		Description_: valid["description"].(string),
		Type_:        valid["type"].(string),
		CountMin_:    int(valid["count-min"].(int64)),
		CountMax_:    int(valid["count-max"].(int64)),
	}, nil
}

func importCharmMetadataPayload(source interface{}, importVersion int) (charmMetadataPayload, error) {
	fields := schema.Fields{
		"name": schema.String(),
		"type": schema.String(),
	}
	defaults := schema.Defaults{}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return charmMetadataPayload{}, errors.Annotate(err, "charmMetadataPayload schema check failed")
	}
	valid := coerced.(map[string]interface{})

	return charmMetadataPayload{
		Name_: valid["name"].(string),
		Type_: valid["type"].(string),
	}, nil
}

func importCharmMetadataResource(source interface{}, importVersion int) (charmMetadataResource, error) {
	fields := schema.Fields{
		"name":        schema.String(),
		"type":        schema.String(),
		"path":        schema.String(),
		"description": schema.String(),
	}
	defaults := schema.Defaults{
		"description": schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return charmMetadataResource{}, errors.Annotate(err, "charmMetadataResource schema check failed")
	}
	valid := coerced.(map[string]interface{})

	return charmMetadataResource{
		Name_:        valid["name"].(string),
		Type_:        valid["type"].(string),
		Path_:        valid["path"].(string),
		Description_: valid["description"].(string),
	}, nil
}

func importCharmMetadataContainer(source interface{}, importVersion int) (charmMetadataContainer, error) {
	fields := schema.Fields{
		"resource": schema.String(),
		"mounts":   schema.List(schema.Any()),
		"uid":      schema.Int(),
		"gid":      schema.Int(),
	}
	defaults := schema.Defaults{
		"uid": schema.Omit,
		"gid": schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return charmMetadataContainer{}, errors.Annotate(err, "charmMetadataContainer schema check failed")
	}
	valid := coerced.(map[string]interface{})

	mounts := make([]charmMetadataContainerMount, 0)
	for _, v := range valid["mounts"].([]interface{}) {
		mount, err := importCharmMetadataContainerMount(v, importVersion)
		if err != nil {
			return charmMetadataContainer{}, errors.Annotate(err, "mount")
		}
		mounts = append(mounts, mount)
	}

	var uid *int
	if valid["uid"] != nil {
		uid = int64ToIntPtr(valid["uid"].(*int64))
	}
	var gid *int
	if valid["gid"] != nil {
		uid = int64ToIntPtr(valid["gid"].(*int64))
	}

	return charmMetadataContainer{
		Resource_: valid["resource"].(string),
		Mounts_:   mounts,
		Uid_:      uid,
		Gid_:      gid,
	}, nil
}

func importCharmMetadataContainerMount(source interface{}, importVersion int) (charmMetadataContainerMount, error) {
	fields := schema.Fields{
		"storage":  schema.String(),
		"location": schema.String(),
	}
	defaults := schema.Defaults{}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return charmMetadataContainerMount{}, errors.Annotate(err, "charmMetadataContainerMount schema check failed")
	}
	valid := coerced.(map[string]interface{})

	return charmMetadataContainerMount{
		Storage_:  valid["storage"].(string),
		Location_: valid["location"].(string),
	}, nil
}

type charmMetadataRelation struct {
	Name_      string `yaml:"name"`
	Role_      string `yaml:"role"`
	Interface_ string `yaml:"interface"`
	Optional_  bool   `yaml:"optional"`
	Limit_     int    `yaml:"limit"`
	Scope_     string `yaml:"scope"`
}

// Name returns the name of the relation.
func (r charmMetadataRelation) Name() string {
	return r.Name_
}

// Role returns the role of the relation.
func (r charmMetadataRelation) Role() string {
	return r.Role_
}

// Interface returns the interface of the relation.
func (r charmMetadataRelation) Interface() string {
	return r.Interface_
}

// Optional returns whether the relation is optional.
func (r charmMetadataRelation) Optional() bool {
	return r.Optional_
}

// Limit returns the limit of the relation.
func (r charmMetadataRelation) Limit() int {
	return r.Limit_
}

// Scope returns the scope of the relation.
func (r charmMetadataRelation) Scope() string {
	return r.Scope_
}

type charmMetadataStorage struct {
	Name_        string   `yaml:"name"`
	Description_ string   `yaml:"description"`
	Type_        string   `yaml:"type"`
	Shared_      bool     `yaml:"shared"`
	Readonly_    bool     `yaml:"readonly"`
	CountMin_    int      `yaml:"count-min"`
	CountMax_    int      `yaml:"count-max"`
	MinimumSize_ int      `yaml:"minimum-size"`
	Location_    string   `yaml:"location"`
	Properties_  []string `yaml:"properties"`
}

// Name returns the name of the storage.
func (s charmMetadataStorage) Name() string {
	return s.Name_
}

// Description returns the description of the storage.
func (s charmMetadataStorage) Description() string {
	return s.Description_
}

// Type returns the type of the storage.
func (s charmMetadataStorage) Type() string {
	return s.Type_
}

// Shared returns whether the storage is shared.
func (s charmMetadataStorage) Shared() bool {
	return s.Shared_
}

// Readonly returns whether the storage is readonly.
func (s charmMetadataStorage) Readonly() bool {
	return s.Readonly_
}

// CountMin returns the minimum count of the storage.
func (s charmMetadataStorage) CountMin() int {
	return s.CountMin_
}

// CountMax returns the maximum count of the storage.
func (s charmMetadataStorage) CountMax() int {
	return s.CountMax_
}

// MinimumSize returns the minimum size of the storage.
func (s charmMetadataStorage) MinimumSize() int {
	return s.MinimumSize_
}

// Location returns the location of the storage.
func (s charmMetadataStorage) Location() string {
	return s.Location_
}

// Properties returns the properties of the storage.
func (s charmMetadataStorage) Properties() []string {
	return s.Properties_
}

type charmMetadataDevice struct {
	Name_        string `yaml:"name"`
	Description_ string `yaml:"description"`
	Type_        string `yaml:"type"`
	CountMin_    int    `yaml:"count-min"`
	CountMax_    int    `yaml:"count-max"`
}

// Name returns the name of the device.
func (d charmMetadataDevice) Name() string {
	return d.Name_
}

// Description returns the description of the device.
func (d charmMetadataDevice) Description() string {
	return d.Description_
}

// Type returns the type of the device.
func (d charmMetadataDevice) Type() string {
	return d.Type_
}

// CountMin returns the minimum count of the device.
func (d charmMetadataDevice) CountMin() int {
	return d.CountMin_
}

// CountMax returns the maximum count of the device.
func (d charmMetadataDevice) CountMax() int {
	return d.CountMax_
}

type charmMetadataPayload struct {
	Name_ string `yaml:"name"`
	Type_ string `yaml:"type"`
}

// Name returns the name of the payload.
func (p charmMetadataPayload) Name() string {
	return p.Name_
}

// Type returns the type of the payload.
func (p charmMetadataPayload) Type() string {
	return p.Type_
}

type charmMetadataResource struct {
	Name_        string `yaml:"name"`
	Type_        string `yaml:"type"`
	Path_        string `yaml:"path"`
	Description_ string `yaml:"description"`
}

// Name returns the name of the resource.
func (r charmMetadataResource) Name() string {
	return r.Name_
}

// Type returns the type of the resource.
func (r charmMetadataResource) Type() string {
	return r.Type_
}

// Path returns the path of the resource.
func (r charmMetadataResource) Path() string {
	return r.Path_
}

// Description returns the description of the resource.
func (r charmMetadataResource) Description() string {
	return r.Description_
}

type charmMetadataContainer struct {
	Resource_ string                        `yaml:"resource"`
	Mounts_   []charmMetadataContainerMount `yaml:"mounts"`
	Uid_      *int                          `yaml:"uid,omitempty"`
	Gid_      *int                          `yaml:"gid,omitempty"`
}

// Resource returns the resource of the container.
func (c charmMetadataContainer) Resource() string {
	return c.Resource_
}

// Mounts returns the mounts of the container.
func (c charmMetadataContainer) Mounts() []CharmMetadataContainerMount {
	mounts := make([]CharmMetadataContainerMount, len(c.Mounts_))
	for i, m := range c.Mounts_ {
		mounts[i] = m
	}
	return mounts
}

// Uid returns the uid of the container.
func (c charmMetadataContainer) Uid() *int {
	return c.Uid_
}

// Gid returns the gid of the container.
func (c charmMetadataContainer) Gid() *int {
	return c.Gid_
}

type charmMetadataContainerMount struct {
	Storage_  string `yaml:"storage"`
	Location_ string `yaml:"location"`
}

// Storage returns the storage of the mount.
func (m charmMetadataContainerMount) Storage() string {
	return m.Storage_
}

// Location returns the location of the mount.
func (m charmMetadataContainerMount) Location() string {
	return m.Location_
}

func int64ToIntPtr(i *int64) *int {
	if i == nil {
		return nil
	}
	p := int(*i)
	return &p
}
