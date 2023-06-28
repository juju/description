// Copyright 2023 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"time"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type LeaseSerializationSuite struct {
	SerializationSuite
}

var _ = gc.Suite(&LeaseSerializationSuite{})

func (s *LeaseSerializationSuite) SetUpTest(c *gc.C) {
	s.importName = "lease"
	s.importFunc = func(m map[string]any) (any, error) {
		return importLease(m)
	}
}

func (s *LeaseSerializationSuite) TestNewLease(c *gc.C) {
	now := time.Now().UTC().Round(time.Second)
	args := LeaseArgs{
		Name:   "name",
		Holder: "holder",
		Start:  now,
		Expiry: now.Add(10 * time.Minute),
		Pinned: true,
	}
	instance := newLease(&args)
	c.Assert(instance.Name(), gc.Equals, "name")
	c.Assert(instance.Holder(), gc.Equals, "holder")
	c.Assert(instance.Start(), gc.DeepEquals, now)
	c.Assert(instance.Expiry(), gc.DeepEquals, now.Add(10*time.Minute))
	c.Assert(instance.Pinned(), jc.IsTrue)
}

func minimalLeaseMap(now time.Time) map[any]any {
	return map[any]any{
		"version": 1,
		"name":    "name",
		"holder":  "holder",
		"start":   now.Format(time.RFC3339Nano),
		"expiry":  now.Add(10 * time.Minute).Format(time.RFC3339Nano),
		"pinned":  true,
	}
}

func minimalLeaseArgs(now time.Time) *LeaseArgs {
	return &LeaseArgs{
		Name:   "name",
		Holder: "holder",
		Start:  now,
		Expiry: now.Add(10 * time.Minute),
		Pinned: true,
	}
}

func minimalLease(now time.Time) *lease {
	return newLease(minimalLeaseArgs(now))
}

func maximalLeaseMap(now time.Time) map[any]any {
	return map[any]any{
		"version": 1,
		"name":    "name",
		"holder":  "holder",
		"start":   now.Format(time.RFC3339Nano),
		"expiry":  now.Add(10 * time.Minute).Format(time.RFC3339Nano),
		"pinned":  true,
	}
}

func maximalLeaseArgs(now time.Time) *LeaseArgs {
	return &LeaseArgs{
		Name:   "name",
		Holder: "holder",
		Start:  now,
		Expiry: now.Add(10 * time.Minute),
		Pinned: true,
	}
}

func maximalLease(now time.Time) *lease {
	return newLease(maximalLeaseArgs(now))
}

func (s *LeaseSerializationSuite) TestMinimalMatches(c *gc.C) {
	now := time.Now().UTC().Round(time.Second)
	bytes, err := yaml.Marshal(minimalLease(now))
	c.Assert(err, jc.ErrorIsNil)

	var source map[any]any
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalLeaseMap(now))
}

func (s *LeaseSerializationSuite) TestMaximalMatches(c *gc.C) {
	now := time.Now().UTC().Round(time.Second)
	bytes, err := yaml.Marshal(maximalLease(now))
	c.Assert(err, jc.ErrorIsNil)

	var source map[any]any
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, maximalLeaseMap(now))
}

func (s *LeaseSerializationSuite) TestParsingSerializedData(c *gc.C) {
	now := time.Now().UTC().Round(time.Second)
	initial := maximalLease(now)
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]any
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importLease(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}
