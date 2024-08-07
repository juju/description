// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// CloudInstance holds information particular to a machine
// instance in a cloud.
type CloudInstance interface {
	HasStatus
	HasStatusHistory
	HasModificationStatus

	InstanceId() string
	DisplayName() string
	Architecture() string
	Memory() uint64
	RootDisk() uint64
	RootDiskSource() string
	CpuCores() uint64
	CpuPower() uint64
	Tags() []string
	AvailabilityZone() string
	VirtType() string
	CharmProfiles() []string

	Validate() error
}

// CloudInstanceArgs is an argument struct used to add information about the
// cloud instance to a Machine.
type CloudInstanceArgs struct {
	InstanceId       string
	DisplayName      string
	Architecture     string
	Memory           uint64
	RootDisk         uint64
	RootDiskSource   string
	CpuCores         uint64
	CpuPower         uint64
	Tags             []string
	AvailabilityZone string
	VirtType         string
	CharmProfiles    []string
}

func newCloudInstance(args CloudInstanceArgs) *cloudInstance {
	tags := make([]string, len(args.Tags))
	copy(tags, args.Tags)
	profiles := make([]string, len(args.CharmProfiles))
	copy(profiles, args.CharmProfiles)
	return &cloudInstance{
		Version:           6,
		InstanceId_:       args.InstanceId,
		DisplayName_:      args.DisplayName,
		Architecture_:     args.Architecture,
		Memory_:           args.Memory,
		RootDisk_:         args.RootDisk,
		RootDiskSource_:   args.RootDiskSource,
		CpuCores_:         args.CpuCores,
		CpuPower_:         args.CpuPower,
		Tags_:             tags,
		AvailabilityZone_: args.AvailabilityZone,
		VirtType_:         args.VirtType,
		CharmProfiles_:    profiles,
		StatusHistory_:    newStatusHistory(),
	}
}

type cloudInstance struct {
	Version int `yaml:"version"`

	InstanceId_  string `yaml:"instance-id"`
	DisplayName_ string `yaml:"display-name,omitempty"`

	Status_        *status `yaml:"status"`
	StatusHistory_ `yaml:"status-history"`

	// ModificationStatus_ defines a status that can be used to highlight status
	// changes to a machine instance after it's been provisioned. This is
	// different from agent-status or machine-status, where the statuses tend to
	// imply how the machine health is during a provisioning cycle or hook
	// integration.
	ModificationStatus_ *status `yaml:"modification-status,omitempty"`

	// For all the optional values, empty values make no sense, and
	// it would be better to have them not set rather than set with
	// a nonsense value.
	Architecture_     string   `yaml:"architecture,omitempty"`
	Memory_           uint64   `yaml:"memory,omitempty"`
	RootDisk_         uint64   `yaml:"root-disk,omitempty"`
	RootDiskSource_   string   `yaml:"root-disk-source,omitempty"`
	CpuCores_         uint64   `yaml:"cores,omitempty"`
	CpuPower_         uint64   `yaml:"cpu-power,omitempty"`
	Tags_             []string `yaml:"tags,omitempty"`
	AvailabilityZone_ string   `yaml:"availability-zone,omitempty"`
	VirtType_         string   `yaml:"virt-type,omitempty"`
	CharmProfiles_    []string `yaml:"charm-profiles,omitempty"`
}

// InstanceId implements CloudInstance.
func (c *cloudInstance) InstanceId() string {
	return c.InstanceId_
}

// DisplayName implements CloudInstance.
func (c *cloudInstance) DisplayName() string {
	return c.DisplayName_
}

// Status implements CloudInstance.
func (c *cloudInstance) Status() Status {
	// To avoid typed nils check nil here.
	if c.Status_ == nil {
		return nil
	}
	return c.Status_
}

// SetStatus implements CloudInstance.
func (c *cloudInstance) SetStatus(args StatusArgs) {
	c.Status_ = newStatus(args)
}

// ModificationStatus implements CloudInstance.
func (c *cloudInstance) ModificationStatus() Status {
	// To avoid typed nils check nil here.
	if c.ModificationStatus_ == nil {
		return nil
	}
	return c.ModificationStatus_
}

// SetModificationStatus implements CloudInstance.
func (c *cloudInstance) SetModificationStatus(args StatusArgs) {
	c.ModificationStatus_ = newStatus(args)
}

// Architecture implements CloudInstance.
func (c *cloudInstance) Architecture() string {
	return c.Architecture_
}

// Memory implements CloudInstance.
func (c *cloudInstance) Memory() uint64 {
	return c.Memory_
}

// RootDisk implements CloudInstance.
func (c *cloudInstance) RootDisk() uint64 {
	return c.RootDisk_
}

// RootDiskSource implements CloudInstance.
func (c *cloudInstance) RootDiskSource() string {
	return c.RootDiskSource_
}

// CpuCores implements CloudInstance.
func (c *cloudInstance) CpuCores() uint64 {
	return c.CpuCores_
}

// CpuPower implements CloudInstance.
func (c *cloudInstance) CpuPower() uint64 {
	return c.CpuPower_
}

// Tags implements CloudInstance.
func (c *cloudInstance) Tags() []string {
	tags := make([]string, len(c.Tags_))
	copy(tags, c.Tags_)
	return tags
}

// AvailabilityZone implements CloudInstance.
func (c *cloudInstance) AvailabilityZone() string {
	return c.AvailabilityZone_
}

// VirtType implements CloudInstance.
func (c *cloudInstance) VirtType() string {
	return c.VirtType_
}

// CharmProfiles implements CloudInstance.
func (c *cloudInstance) CharmProfiles() []string {
	profiles := make([]string, len(c.CharmProfiles_))
	copy(profiles, c.CharmProfiles_)
	return profiles
}

// Validate implements CloudInstance.
func (c *cloudInstance) Validate() error {
	if c.InstanceId_ == "" {
		return errors.NotValidf("instance missing id")
	}
	if c.Status_ == nil {
		return errors.NotValidf("instance %q missing status", c.InstanceId_)
	}
	return nil
}

func importCloudInstance(source map[string]interface{}) (*cloudInstance, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "cloudInstance version schema check failed")
	}

	getFields, ok := cloudInstanceFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	return importCloudInstanceVx(source, version, getFields)
}

var cloudInstanceFieldsFuncs = map[int]fieldsFunc{
	1: cloudInstanceV1Fields,
	2: cloudInstanceV2Fields,
	3: cloudInstanceV3Fields,
	4: cloudInstanceV4Fields,
	5: cloudInstanceV5Fields,
	6: cloudInstanceV6Fields,
}

func cloudInstanceV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"instance-id":       schema.String(),
		"display-name":      schema.String(),
		"status":            schema.String(),
		"architecture":      schema.String(),
		"memory":            schema.ForceUint(),
		"root-disk":         schema.ForceUint(),
		"cores":             schema.ForceUint(),
		"cpu-power":         schema.ForceUint(),
		"tags":              schema.List(schema.String()),
		"availability-zone": schema.String(),
	}
	// Some values don't have to be there.
	defaults := schema.Defaults{
		"architecture":      "",
		"display-name":      schema.Omit,
		"memory":            uint64(0),
		"root-disk":         uint64(0),
		"cores":             uint64(0),
		"cpu-power":         uint64(0),
		"tags":              schema.Omit,
		"availability-zone": "",
	}
	return fields, defaults
}

func cloudInstanceV2Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := cloudInstanceV1Fields()
	fields["status"] = schema.StringMap(schema.Any())
	addStatusHistorySchema(fields)
	return fields, defaults
}

func cloudInstanceV3Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := cloudInstanceV2Fields()
	fields["charm-profiles"] = schema.List(schema.String())
	defaults["charm-profiles"] = schema.Omit
	return fields, defaults
}

func cloudInstanceV4Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := cloudInstanceV3Fields()
	fields["modification-status"] = schema.StringMap(schema.Any())
	defaults["modification-status"] = schema.Omit
	return fields, defaults
}

func cloudInstanceV5Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := cloudInstanceV4Fields()
	fields["root-disk-source"] = schema.String()
	defaults["root-disk-source"] = ""
	return fields, defaults
}

func cloudInstanceV6Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := cloudInstanceV5Fields()
	fields["virt-type"] = schema.String()
	defaults["virt-type"] = ""
	return fields, defaults
}

func importCloudInstanceVx(source map[string]interface{}, version int, fieldFunc func() (schema.Fields, schema.Defaults)) (*cloudInstance, error) {
	fields, defaults := fieldFunc()
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "cloudInstance v%d schema check failed", version)
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	return newCloudInstanceFromValid(valid, version)
}

func newCloudInstanceFromValid(valid map[string]interface{}, importVersion int) (*cloudInstance, error) {
	instance := &cloudInstance{
		Version:           importVersion,
		InstanceId_:       valid["instance-id"].(string),
		Architecture_:     valid["architecture"].(string),
		Memory_:           valid["memory"].(uint64),
		RootDisk_:         valid["root-disk"].(uint64),
		CpuCores_:         valid["cores"].(uint64),
		CpuPower_:         valid["cpu-power"].(uint64),
		Tags_:             convertToStringSlice(valid["tags"]),
		AvailabilityZone_: valid["availability-zone"].(string),
		CharmProfiles_:    convertToStringSlice(valid["charm-profiles"]),
		StatusHistory_:    newStatusHistory(),
	}

	if displayName, ok := valid["display-name"].(string); ok {
		instance.DisplayName_ = displayName
	}

	switch {
	case importVersion == 1:
		// Status was exported incorrectly, so we fake one here.
		instance.SetStatus(StatusArgs{
			Value: "unknown",
		})

	case importVersion >= 2:
		status, err := importStatus(valid["status"].(map[string]interface{}))
		if err != nil {
			return nil, errors.Trace(err)
		}
		instance.Status_ = status
		if err := instance.importStatusHistory(valid); err != nil {
			return nil, errors.Trace(err)
		}

		if importVersion > 3 {
			modificationStatus, err := importModificationStatus(valid["modification-status"])
			if err != nil {
				return nil, errors.Trace(err)
			}
			instance.ModificationStatus_ = modificationStatus
		}

		if importVersion > 4 {
			instance.RootDiskSource_ = valid["root-disk-source"].(string)
		}

		if importVersion > 5 {
			instance.VirtType_ = valid["virt-type"].(string)
		}
	default:
		return nil, errors.NotValidf("unexpected version: %d", importVersion)
	}

	return instance, nil
}
