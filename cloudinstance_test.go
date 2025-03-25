// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type CloudInstanceSerializationSuite struct {
	SerializationSuite
}

var _ = gc.Suite(&CloudInstanceSerializationSuite{})

func (s *CloudInstanceSerializationSuite) SetUpTest(c *gc.C) {
	s.importName = "cloudInstance"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importCloudInstance(m)
	}
}

func minimalCloudInstanceMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version":             6,
		"instance-id":         "instance id",
		"status":              minimalStatusMap(),
		"status-history":      emptyStatusHistoryMap(),
		"modification-status": minimalStatusMap(),
	}
}

func minimalCloudInstanceMapV3() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version":        3,
		"instance-id":    "instance id",
		"status":         minimalStatusMap(),
		"status-history": emptyStatusHistoryMap(),
	}
}

func minimalCloudInstance() *cloudInstance {
	instance := newCloudInstance(minimalCloudInstanceArgs())
	instance.SetStatus(minimalStatusArgs())
	instance.SetModificationStatus(minimalStatusArgs())
	return instance
}

func minimalCloudInstanceArgs() CloudInstanceArgs {
	return CloudInstanceArgs{
		InstanceId: "instance id",
	}
}

func (s *CloudInstanceSerializationSuite) TestNewCloudInstance(c *gc.C) {
	args := s.testArgs()
	var instance CloudInstance = s.testCloudInstance()
	c.Check(instance.Validate(), jc.ErrorIsNil)
	c.Check(instance.InstanceId(), gc.Equals, args.InstanceId)
	c.Check(instance.Architecture(), gc.Equals, args.Architecture)
	c.Check(instance.Memory(), gc.Equals, args.Memory)
	c.Check(instance.RootDisk(), gc.Equals, args.RootDisk)
	c.Check(instance.RootDiskSource(), gc.Equals, args.RootDiskSource)
	c.Check(instance.CpuCores(), gc.Equals, args.CpuCores)
	c.Check(instance.CpuPower(), gc.Equals, args.CpuPower)
	c.Check(instance.VirtType(), gc.Equals, args.VirtType)
	c.Check(instance.AvailabilityZone(), gc.Equals, args.AvailabilityZone)

	// Before we check tags, modify args to make sure that the instance ones
	// don't change.
	args.Tags[0] = "weird"
	tags := instance.Tags()
	c.Assert(tags, jc.DeepEquals, []string{"much", "strong"})

	// Also, changing the tags returned, doesn't modify the instance
	tags[0] = "weird"
	c.Assert(instance.Tags(), jc.DeepEquals, []string{"much", "strong"})

	// Before we check charm profiles, modify args to make sure that the instance ones
	// don't change.
	args.CharmProfiles[0] = "weird"
	profiles := instance.CharmProfiles()
	c.Assert(profiles, jc.DeepEquals, []string{"much", "strong"})

	// Also, changing the tags returned, doesn't modify the instance
	profiles[0] = "weird"
	c.Assert(instance.CharmProfiles(), jc.DeepEquals, []string{"much", "strong"})

	// Check that the modification status is valid
	c.Check(instance.ModificationStatus(), gc.DeepEquals, newStatus(minimalStatusArgs()))
}

func (s *CloudInstanceSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalCloudInstance())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalCloudInstanceMap())
}

func (s *CloudInstanceSerializationSuite) TestParsingSerializedData(c *gc.C) {
	const MaxUint64 = 1<<64 - 1
	initial := newCloudInstance(CloudInstanceArgs{
		InstanceId:   "instance id",
		Architecture: "amd64",
		Memory:       16 * gig,
		CpuPower:     MaxUint64,
		Tags:         []string{"much", "strong"},
	})
	initial.SetStatus(minimalStatusArgs())
	initial.SetModificationStatus(minimalStatusArgs())
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importCloudInstance(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}

func (s *CloudInstanceSerializationSuite) TestValidateMissingID(c *gc.C) {
	initial := newCloudInstance(CloudInstanceArgs{})
	err := initial.Validate()
	c.Check(err, jc.ErrorIs, errors.NotValid)
	c.Check(err, gc.ErrorMatches, "instance missing id not valid")
}

func (s *CloudInstanceSerializationSuite) TestValidateMissingStatus(c *gc.C) {
	initial := newCloudInstance(CloudInstanceArgs{InstanceId: "magic"})
	err := initial.Validate()
	c.Check(err, jc.ErrorIs, errors.NotValid)
	c.Check(err, gc.ErrorMatches, `instance "magic" missing status not valid`)
}

func (s *CloudInstanceSerializationSuite) TestValidateInvalidModificationStatus(c *gc.C) {
	args := CloudInstanceArgs{
		InstanceId: "instance id",
	}
	instance := newCloudInstance(args)
	instance.SetStatus(minimalStatusArgs())
	instance.SetModificationStatus(StatusArgs{})

	err := instance.Validate()
	c.Check(err, gc.IsNil)
}

func (s *CloudInstanceSerializationSuite) importCloudInstance(c *gc.C, source map[string]interface{}) *cloudInstance {
	imported, err := importCloudInstance(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(imported, gc.NotNil)
	return imported
}

func (s *CloudInstanceSerializationSuite) testArgs() CloudInstanceArgs {
	// NOTE: using gig from package_test.go
	return CloudInstanceArgs{
		InstanceId:       "instance id",
		DisplayName:      "foo",
		Architecture:     "amd64",
		Memory:           16 * gig,
		RootDisk:         200 * gig,
		RootDiskSource:   "my-house",
		CpuCores:         8,
		CpuPower:         4000,
		Tags:             []string{"much", "strong"},
		AvailabilityZone: "everywhere",
		VirtType:         "container",
		CharmProfiles:    []string{"much", "strong"},
	}
}

func (s *CloudInstanceSerializationSuite) testCloudInstance() *cloudInstance {
	instance := newCloudInstance(s.testArgs())
	instance.SetStatus(minimalStatusArgs())
	instance.SetModificationStatus(minimalStatusArgs())
	return instance
}

func (s *CloudInstanceSerializationSuite) allV4Map() map[string]interface{} {
	return map[string]interface{}{
		"version":             4,
		"instance-id":         "instance id",
		"display-name":        "foo",
		"status":              minimalStatusMap(),
		"status-history":      emptyStatusHistoryMap(),
		"modification-status": minimalStatusMap(),
		"architecture":        "amd64",
		"memory":              16 * gig,
		"root-disk":           200 * gig,
		"cores":               8,
		"cpu-power":           4000,
		"tags":                []string{"much", "strong"},
		"availability-zone":   "everywhere",
		"charm-profiles":      []string{"much", "strong"},
	}
}

func (s *CloudInstanceSerializationSuite) TestParsingV4Full(c *gc.C) {
	original := s.allV4Map()
	imported := s.importCloudInstance(c, original)
	expected := s.testCloudInstance()
	expected.RootDiskSource_ = ""
	expected.VirtType_ = ""
	expected.Version = 4
	c.Assert(imported, jc.DeepEquals, expected)
}

func (s *CloudInstanceSerializationSuite) TestParsingV4Minimal(c *gc.C) {
	original := map[string]interface{}{
		"version":             4,
		"instance-id":         "instance id",
		"status":              minimalStatusMap(),
		"status-history":      emptyStatusHistoryMap(),
		"modification-status": minimalStatusMap(),
	}
	imported := s.importCloudInstance(c, original)
	expected := newCloudInstance(minimalCloudInstanceArgs())
	expected.SetStatus(minimalStatusArgs())
	expected.SetModificationStatus(minimalStatusArgs())
	expected.Version = 4
	c.Assert(imported, jc.DeepEquals, expected)
}

func (s *CloudInstanceSerializationSuite) TestParsingV4IgnoresNewField(c *gc.C) {
	original := s.allV4Map()
	original["root-disk-source"] = "somewhere"
	imported := s.importCloudInstance(c, original)
	c.Assert(imported.RootDiskSource_, gc.Equals, "")
}

func (s *CloudInstanceSerializationSuite) allV5Map() map[string]interface{} {
	return map[string]interface{}{
		"version":             5,
		"instance-id":         "instance id",
		"display-name":        "foo",
		"status":              minimalStatusMap(),
		"status-history":      emptyStatusHistoryMap(),
		"modification-status": minimalStatusMap(),
		"architecture":        "amd64",
		"memory":              16 * gig,
		"root-disk":           200 * gig,
		"root-disk-source":    "my-house",
		"cores":               8,
		"cpu-power":           4000,
		"tags":                []string{"much", "strong"},
		"availability-zone":   "everywhere",
		"charm-profiles":      []string{"much", "strong"},
	}
}

func (s *CloudInstanceSerializationSuite) TestParsingV5Full(c *gc.C) {
	original := s.allV5Map()
	imported := s.importCloudInstance(c, original)
	expected := s.testCloudInstance()
	expected.VirtType_ = ""
	expected.Version = 5
	c.Assert(imported, jc.DeepEquals, expected)
}

func (s *CloudInstanceSerializationSuite) TestParsingV5Minimal(c *gc.C) {
	original := map[string]interface{}{
		"version":             5,
		"instance-id":         "instance id",
		"status":              minimalStatusMap(),
		"status-history":      emptyStatusHistoryMap(),
		"modification-status": minimalStatusMap(),
	}
	imported := s.importCloudInstance(c, original)
	expected := newCloudInstance(minimalCloudInstanceArgs())
	expected.SetStatus(minimalStatusArgs())
	expected.SetModificationStatus(minimalStatusArgs())
	expected.Version = 5
	c.Assert(imported, jc.DeepEquals, expected)
}

func (s *CloudInstanceSerializationSuite) allV6Map() map[string]interface{} {
	return map[string]interface{}{
		"version":             6,
		"instance-id":         "instance id",
		"display-name":        "foo",
		"status":              minimalStatusMap(),
		"status-history":      emptyStatusHistoryMap(),
		"modification-status": minimalStatusMap(),
		"architecture":        "amd64",
		"memory":              16 * gig,
		"root-disk":           200 * gig,
		"root-disk-source":    "my-house",
		"cores":               8,
		"cpu-power":           4000,
		"tags":                []string{"much", "strong"},
		"availability-zone":   "everywhere",
		"virt-type":           "container",
		"charm-profiles":      []string{"much", "strong"},
	}
}

func (s *CloudInstanceSerializationSuite) TestParsingV6Full(c *gc.C) {
	original := s.allV6Map()
	imported := s.importCloudInstance(c, original)
	expected := s.testCloudInstance()
	c.Assert(imported, jc.DeepEquals, expected)
}

func (s *CloudInstanceSerializationSuite) TestParsingV6Minimal(c *gc.C) {
	original := map[string]interface{}{
		"version":             6,
		"instance-id":         "instance id",
		"status":              minimalStatusMap(),
		"status-history":      emptyStatusHistoryMap(),
		"modification-status": minimalStatusMap(),
	}
	imported := s.importCloudInstance(c, original)
	expected := newCloudInstance(minimalCloudInstanceArgs())
	expected.SetStatus(minimalStatusArgs())
	expected.SetModificationStatus(minimalStatusArgs())
	c.Assert(imported, jc.DeepEquals, expected)
}
