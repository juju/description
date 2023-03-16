// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// HasConstraints defines the common methods for setting and
// getting constraints for the various entities.
type HasConstraints interface {
	Constraints() Constraints
	SetConstraints(ConstraintsArgs)
}

// Constraints holds information about particular deployment
// constraints for entities.
type Constraints interface {
	AllocatePublicIP() bool
	Architecture() string
	Container() string
	CpuCores() uint64
	CpuPower() uint64
	ImageID() string
	InstanceType() string
	Memory() uint64
	RootDisk() uint64
	RootDiskSource() string

	Spaces() []string
	Tags() []string
	Zones() []string

	VirtType() string
}

// ConstraintsArgs is an argument struct to construct Constraints.
type ConstraintsArgs struct {
	AllocatePublicIP bool
	Architecture     string
	Container        string
	CpuCores         uint64
	CpuPower         uint64
	ImageID          string
	InstanceType     string
	Memory           uint64
	RootDisk         uint64
	RootDiskSource   string

	Spaces []string
	Tags   []string
	Zones  []string

	VirtType string
}

func newConstraints(args ConstraintsArgs) *constraints {
	// If the ConstraintsArgs are all empty, then we return
	// nil to indicate that there are no constraints.
	if args.empty() {
		return nil
	}

	tags := make([]string, len(args.Tags))
	copy(tags, args.Tags)
	spaces := make([]string, len(args.Spaces))
	copy(spaces, args.Spaces)
	zones := make([]string, len(args.Zones))
	copy(zones, args.Zones)

	return &constraints{
		Version:           5,
		AllocatePublicIP_: args.AllocatePublicIP,
		Architecture_:     args.Architecture,
		Container_:        args.Container,
		CpuCores_:         args.CpuCores,
		CpuPower_:         args.CpuPower,
		ImageID_:          args.ImageID,
		InstanceType_:     args.InstanceType,
		Memory_:           args.Memory,
		RootDisk_:         args.RootDisk,
		RootDiskSource_:   args.RootDiskSource,
		Spaces_:           spaces,
		Tags_:             tags,
		Zones_:            zones,
		VirtType_:         args.VirtType,
	}
}

type constraints struct {
	Version int `yaml:"version"`

	AllocatePublicIP_ bool   `yaml:"allocate-public-ip,omitempty"`
	Architecture_     string `yaml:"architecture,omitempty"`
	Container_        string `yaml:"container,omitempty"`
	CpuCores_         uint64 `yaml:"cores,omitempty"`
	CpuPower_         uint64 `yaml:"cpu-power,omitempty"`
	ImageID_          string `yaml:"image-id,omitempty"`
	InstanceType_     string `yaml:"instance-type,omitempty"`
	Memory_           uint64 `yaml:"memory,omitempty"`
	RootDisk_         uint64 `yaml:"root-disk,omitempty"`
	RootDiskSource_   string `yaml:"root-disk-source,omitempty"`

	Spaces_ []string `yaml:"spaces,omitempty"`
	Tags_   []string `yaml:"tags,omitempty"`
	Zones_  []string `yaml:"zones,omitempty"`

	VirtType_ string `yaml:"virt-type,omitempty"`
}

// AllocatePublicIP implements Constraints.
func (c *constraints) AllocatePublicIP() bool {
	return c.AllocatePublicIP_
}

// Architecture implements Constraints.
func (c *constraints) Architecture() string {
	return c.Architecture_
}

// Container implements Constraints.
func (c *constraints) Container() string {
	return c.Container_
}

// CpuCores implements Constraints.
func (c *constraints) CpuCores() uint64 {
	return c.CpuCores_
}

// CpuPower implements Constraints.
func (c *constraints) CpuPower() uint64 {
	return c.CpuPower_
}

// ImageID implements Constraints.
func (c *constraints) ImageID() string {
	return c.ImageID_
}

// InstanceType implements Constraints.
func (c *constraints) InstanceType() string {
	return c.InstanceType_
}

// Memory implements Constraints.
func (c *constraints) Memory() uint64 {
	return c.Memory_
}

// RootDisk implements Constraints.
func (c *constraints) RootDisk() uint64 {
	return c.RootDisk_
}

// RootDiskSource implements Constraints.
func (c *constraints) RootDiskSource() string {
	return c.RootDiskSource_
}

// Spaces implements Constraints.
func (c *constraints) Spaces() []string {
	var spaces []string
	if count := len(c.Spaces_); count > 0 {
		spaces = make([]string, count)
		copy(spaces, c.Spaces_)
	}
	return spaces
}

// Tags implements Constraints.
func (c *constraints) Tags() []string {
	var tags []string
	if count := len(c.Tags_); count > 0 {
		tags = make([]string, count)
		copy(tags, c.Tags_)
	}
	return tags
}

// Zones implements Constraints.
func (c *constraints) Zones() []string {
	var zones []string
	if count := len(c.Zones_); count > 0 {
		zones = make([]string, count)
		copy(zones, c.Zones_)
	}
	return zones
}

// VirtType implements Constraints.
func (c *constraints) VirtType() string {
	return c.VirtType_
}

func importConstraints(source map[string]interface{}) (*constraints, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "constraints version schema check failed")
	}
	getFields, ok := constraintsFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	checker := schema.FieldMap(getFields())

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "constraints v%d schema check failed", version)
	}

	valid := coerced.(map[string]interface{})
	cores, err := constraintsValidCPUCores(valid)
	if err != nil {
		return nil, err
	}

	return validatedConstraints(version, valid, cores), nil
}

var constraintsFieldsFuncs = map[int]fieldsFunc{
	1: constraintsV1Fields,
	2: constraintsV2Fields,
	3: constraintsV3Fields,
	4: constraintsV4Fields,
	5: constraintsV5Fields,
}

func constraintsV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"architecture":  schema.String(),
		"container":     schema.String(),
		"cpu-cores":     schema.ForceUint(),
		"cores":         schema.ForceUint(),
		"cpu-power":     schema.ForceUint(),
		"instance-type": schema.String(),
		"memory":        schema.ForceUint(),
		"root-disk":     schema.ForceUint(),

		"spaces": schema.List(schema.String()),
		"tags":   schema.List(schema.String()),

		"virt-type": schema.String(),
	}
	defaults := schema.Defaults{
		"architecture":  "",
		"container":     "",
		"cpu-cores":     schema.Omit,
		"cores":         schema.Omit,
		"cpu-power":     uint64(0),
		"instance-type": "",
		"memory":        uint64(0),
		"root-disk":     uint64(0),

		"spaces": schema.Omit,
		"tags":   schema.Omit,

		"virt-type": "",
	}
	return fields, defaults
}

func constraintsV2Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := constraintsV1Fields()
	fields["zones"] = schema.List(schema.String())
	defaults["zones"] = schema.Omit
	return fields, defaults
}

func constraintsV3Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := constraintsV2Fields()
	fields["root-disk-source"] = schema.String()
	defaults["root-disk-source"] = ""
	return fields, defaults
}

func constraintsV4Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := constraintsV3Fields()
	fields["allocate-public-ip"] = schema.Bool()
	defaults["allocate-public-ip"] = schema.Omit
	return fields, defaults
}

func constraintsV5Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := constraintsV4Fields()
	fields["image-id"] = schema.String()
	defaults["image-id"] = ""
	return fields, defaults
}

// constraintsValidCPUCores returns an error if both aliases for CPU core count
// are present in the list of fields.
// If correctly specified, the cores value is returned.
func constraintsValidCPUCores(valid map[string]interface{}) (uint64, error) {
	var cores uint64

	_, hasCPU := valid["cpu-cores"]
	_, hasCores := valid["cores"]
	if hasCPU && hasCores {
		return cores, errors.Errorf("can not specify both cores and cores constraints")
	}

	if hasCPU {
		cores = valid["cpu-cores"].(uint64)
	}
	if hasCores {
		cores = valid["cores"].(uint64)
	}
	return cores, nil
}

// validatedConstraints returns a constraints reference from the supplied
// *valid* fields.
func validatedConstraints(version int, valid map[string]interface{}, cores uint64) *constraints {
	cons := &constraints{
		Version:       version,
		Architecture_: valid["architecture"].(string),
		Container_:    valid["container"].(string),
		CpuCores_:     cores,
		CpuPower_:     valid["cpu-power"].(uint64),
		InstanceType_: valid["instance-type"].(string),
		Memory_:       valid["memory"].(uint64),
		RootDisk_:     valid["root-disk"].(uint64),

		Spaces_: convertToStringSlice(valid["spaces"]),
		Tags_:   convertToStringSlice(valid["tags"]),

		VirtType_: valid["virt-type"].(string),
	}

	if version > 1 {
		cons.Zones_ = convertToStringSlice(valid["zones"])
	}
	if version > 2 {
		cons.RootDiskSource_ = valid["root-disk-source"].(string)
	}
	if version > 3 {
		if value, ok := valid["allocate-public-ip"]; ok {
			cons.AllocatePublicIP_ = value.(bool)
		}
	}
	if version > 4 {
		cons.ImageID_ = valid["image-id"].(string)
	}

	return cons
}

func addConstraintsSchema(fields schema.Fields, defaults schema.Defaults) {
	fields["constraints"] = schema.StringMap(schema.Any())
	defaults["constraints"] = schema.Omit
}

func (c ConstraintsArgs) empty() bool {
	return c.Architecture == "" &&
		c.Container == "" &&
		c.CpuCores == 0 &&
		c.CpuPower == 0 &&
		c.ImageID == "" &&
		c.InstanceType == "" &&
		c.Memory == 0 &&
		c.RootDisk == 0 &&
		c.RootDiskSource == "" &&
		c.Spaces == nil &&
		c.Tags == nil &&
		c.Zones == nil &&
		c.VirtType == "" &&
		c.AllocatePublicIP == false
}
