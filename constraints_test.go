// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type ConstraintsSerializationSuite struct {
	SerializationSuite
}

var _ = gc.Suite(&ConstraintsSerializationSuite{})

func (s *ConstraintsSerializationSuite) SetUpTest(c *gc.C) {
	s.importName = "constraints"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importConstraints(m)
	}
}

func (s *ConstraintsSerializationSuite) allArgs() ConstraintsArgs {
	// NOTE: using gig from package_test.go
	return ConstraintsArgs{
		AllocatePublicIP: true,
		Architecture:     "amd64",
		Container:        "lxd",
		CpuCores:         8,
		CpuPower:         4000,
		InstanceType:     "magic",
		Memory:           16 * gig,
		RootDisk:         200 * gig,
		RootDiskSource:   "somewhere-good",
		Spaces:           []string{"my", "own"},
		Tags:             []string{"much", "strong"},
		Zones:            []string{"az1", "az2"},
		VirtType:         "something",
	}
}

func (s *ConstraintsSerializationSuite) TestNewConstraints(c *gc.C) {
	args := s.allArgs()
	var instance Constraints = newConstraints(args)

	c.Assert(instance.Architecture(), gc.Equals, args.Architecture)
	c.Assert(instance.Container(), gc.Equals, args.Container)
	c.Assert(instance.CpuCores(), gc.Equals, args.CpuCores)
	c.Assert(instance.CpuPower(), gc.Equals, args.CpuPower)
	c.Assert(instance.InstanceType(), gc.Equals, args.InstanceType)
	c.Assert(instance.Memory(), gc.Equals, args.Memory)
	c.Assert(instance.RootDisk(), gc.Equals, args.RootDisk)
	c.Assert(instance.RootDiskSource(), gc.Equals, args.RootDiskSource)
	c.Assert(instance.AllocatePublicIP(), gc.Equals, args.AllocatePublicIP)

	// Before we check tags, spaces and zones, modify args to make sure that
	// the instance ones do not change.
	args.Spaces[0] = "weird"
	args.Tags[0] = "weird"
	args.Zones[0] = "weird"
	spaces := instance.Spaces()
	c.Assert(spaces, jc.DeepEquals, []string{"my", "own"})
	tags := instance.Tags()
	c.Assert(tags, jc.DeepEquals, []string{"much", "strong"})
	zones := instance.Zones()
	c.Assert(zones, jc.DeepEquals, []string{"az1", "az2"})

	// Also, changing the spaces, tags or zones returned
	// does not modify the instance.
	spaces[0] = "weird"
	tags[0] = "weird"
	zones[0] = "weird"
	c.Assert(instance.Spaces(), jc.DeepEquals, []string{"my", "own"})
	c.Assert(instance.Tags(), jc.DeepEquals, []string{"much", "strong"})
	c.Assert(instance.Zones(), jc.DeepEquals, []string{"az1", "az2"})
}

func (s *ConstraintsSerializationSuite) TestNewConstraintsWithVirt(c *gc.C) {
	args := s.allArgs()
	args.VirtType = "kvm"
	instance := newConstraints(args)
	c.Assert(instance.VirtType(), gc.Equals, args.VirtType)
}

func (s *ConstraintsSerializationSuite) TestNewConstraintsEmpty(c *gc.C) {
	instance := newConstraints(ConstraintsArgs{})
	c.Assert(instance, gc.IsNil)
}

func (s *ConstraintsSerializationSuite) TestEmptyTagsSpacesZones(c *gc.C) {
	instance := newConstraints(ConstraintsArgs{Architecture: "amd64"})
	// We actually want them to be nil, not empty slices.
	c.Assert(instance.Tags(), gc.IsNil)
	c.Assert(instance.Spaces(), gc.IsNil)
	c.Assert(instance.Zones(), gc.IsNil)
}

func (s *ConstraintsSerializationSuite) TestEmptyVirt(c *gc.C) {
	instance := newConstraints(ConstraintsArgs{Architecture: "amd64"})
	c.Assert(instance.VirtType(), gc.Equals, "")
}

func (s *ConstraintsSerializationSuite) TestParsingSerializedData(c *gc.C) {
	s.assertParsingSerializedConstraints(c, newConstraints(s.allArgs()))
}

func (s *ConstraintsSerializationSuite) TestParsingSerializedVirt(c *gc.C) {
	args := s.allArgs()
	args.VirtType = "kvm"
	s.assertParsingSerializedConstraints(c, newConstraints(args))
}

func (s *ConstraintsSerializationSuite) assertParsingSerializedConstraints(c *gc.C, initial Constraints) {
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importConstraints(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}

func (s *ConstraintsSerializationSuite) testConstraints() *constraints {
	return newConstraints(s.allArgs())
}

func (s *ConstraintsSerializationSuite) importConstraints(c *gc.C, original map[string]interface{}) *constraints {
	imported, err := importConstraints(original)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(imported, gc.NotNil)
	return imported
}

func (s *ConstraintsSerializationSuite) allV1Map() map[string]interface{} {
	return map[string]interface{}{
		"version":       1,
		"architecture":  "amd64",
		"container":     "lxd",
		"cores":         8,
		"cpu-power":     4000,
		"instance-type": "magic",
		"memory":        16 * gig,
		"root-disk":     200 * gig,
		"spaces":        []interface{}{"my", "own"},
		"tags":          []interface{}{"much", "strong"},
		"virt-type":     "something",
	}
}

func (s *ConstraintsSerializationSuite) TestParsingV1Full(c *gc.C) {
	original := s.allV1Map()
	imported := s.importConstraints(c, original)
	expected := s.testConstraints()
	expected.Zones_ = nil
	expected.RootDiskSource_ = ""
	expected.AllocatePublicIP_ = false
	expected.Version = 1
	c.Assert(imported, gc.DeepEquals, expected)
}

func (s *ConstraintsSerializationSuite) TestParsingV1Minimal(c *gc.C) {
	original := map[string]interface{}{
		"version": 1,
	}
	imported := s.importConstraints(c, original)
	expected := &constraints{Version: 1}
	c.Assert(imported, gc.DeepEquals, expected)
}

func (s *ConstraintsSerializationSuite) TestParsingV1IgnoresNewFields(c *gc.C) {
	original := s.allV1Map()
	original["zones"] = []string{"whatever"}
	imported := s.importConstraints(c, original)
	c.Assert(imported.Zones_, gc.IsNil)
}

func (s *ConstraintsSerializationSuite) allV2Map() map[string]interface{} {
	return map[string]interface{}{
		"version":       2,
		"architecture":  "amd64",
		"container":     "lxd",
		"cores":         8,
		"cpu-power":     4000,
		"instance-type": "magic",
		"memory":        16 * gig,
		"root-disk":     200 * gig,
		"spaces":        []interface{}{"my", "own"},
		"tags":          []interface{}{"much", "strong"},
		"zones":         []interface{}{"az1", "az2"},
		"virt-type":     "something",
	}
}

func (s *ConstraintsSerializationSuite) TestParsingV2Full(c *gc.C) {
	original := s.allV2Map()
	imported := s.importConstraints(c, original)
	expected := s.testConstraints()
	expected.RootDiskSource_ = ""
	expected.AllocatePublicIP_ = false
	expected.Version = 2
	c.Assert(imported, gc.DeepEquals, expected)
}

func (s *ConstraintsSerializationSuite) TestParsingV2Minimal(c *gc.C) {
	original := map[string]interface{}{
		"version": 2,
	}
	imported := s.importConstraints(c, original)
	expected := &constraints{Version: 2}
	c.Assert(imported, gc.DeepEquals, expected)
}

func (s *ConstraintsSerializationSuite) TestParsingV2IgnoresNewFields(c *gc.C) {
	original := s.allV2Map()
	original["root-disk-source"] = "secret-sauce"
	imported := s.importConstraints(c, original)
	c.Assert(imported.RootDiskSource_, gc.Equals, "")
}

func (s *ConstraintsSerializationSuite) allV3Map() map[string]interface{} {
	return map[string]interface{}{
		"version":          3,
		"architecture":     "amd64",
		"container":        "lxd",
		"cores":            8,
		"cpu-power":        4000,
		"instance-type":    "magic",
		"memory":           16 * gig,
		"root-disk":        200 * gig,
		"root-disk-source": "somewhere-good",
		"spaces":           []interface{}{"my", "own"},
		"tags":             []interface{}{"much", "strong"},
		"zones":            []interface{}{"az1", "az2"},
		"virt-type":        "something",
	}
}

func (s *ConstraintsSerializationSuite) TestParsingV3Full(c *gc.C) {
	original := s.allV3Map()
	imported := s.importConstraints(c, original)
	expected := s.testConstraints()
	expected.AllocatePublicIP_ = false
	expected.Version = 3
	c.Assert(imported, gc.DeepEquals, expected)
}

func (s *ConstraintsSerializationSuite) TestParsingV3Minimal(c *gc.C) {
	original := map[string]interface{}{
		"version": 3,
	}
	imported := s.importConstraints(c, original)
	expected := &constraints{Version: 3}
	c.Assert(imported, gc.DeepEquals, expected)
}

func (s *ConstraintsSerializationSuite) allV4Map() map[string]interface{} {
	return map[string]interface{}{
		"version":            4,
		"allocate-public-ip": true,
		"architecture":       "amd64",
		"container":          "lxd",
		"cores":              8,
		"cpu-power":          4000,
		"instance-type":      "magic",
		"memory":             16 * gig,
		"root-disk":          200 * gig,
		"root-disk-source":   "somewhere-good",
		"spaces":             []interface{}{"my", "own"},
		"tags":               []interface{}{"much", "strong"},
		"zones":              []interface{}{"az1", "az2"},
		"virt-type":          "something",
	}
}

func (s *ConstraintsSerializationSuite) TestParsingV4Full(c *gc.C) {
	original := s.allV4Map()
	imported := s.importConstraints(c, original)
	expected := s.testConstraints()
	c.Assert(imported, gc.DeepEquals, expected)
}

func (s *ConstraintsSerializationSuite) TestParsingV4Minimal(c *gc.C) {
	original := map[string]interface{}{
		"version": 4,
	}
	imported := s.importConstraints(c, original)
	expected := &constraints{Version: 4}
	c.Assert(imported, gc.DeepEquals, expected)
}
