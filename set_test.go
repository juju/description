// Copyright 2025 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

type stringSetSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&stringSetSuite{})

func (suite *stringSetSuite) TestEmpty(c *gc.C) {
	s := make(stringsSet)
	c.Assert(s.values(), gc.HasLen, 0)
}

func (suite *stringSetSuite) TestAdd(c *gc.C) {
	s := make(stringsSet)
	s.add("foo")
	s.add("foo")
	s.add("bar")
	c.Assert(s.values(), jc.DeepEquals, []string{"bar", "foo"})
}

func (suite *stringSetSuite) TestContains(c *gc.C) {
	s := make(stringsSet)
	s.add("foo")
	s.add("bar")
	c.Assert(s.contains("foo"), gc.Equals, true)
	c.Assert(s.contains("bar"), gc.Equals, true)
	c.Assert(s.contains("baz"), gc.Equals, false)
}

func (suite *stringSetSuite) TestUnion(c *gc.C) {
	s1 := make(stringsSet)
	s1.add("foo")
	s1.add("bar")
	s2 := make(stringsSet)
	s2.add("foo")
	s2.add("baz")
	s2.add("bang")
	union1 := s1.union(s2)
	union2 := s2.union(s1)

	c.Assert(union1.values(), jc.DeepEquals, []string{"bang", "bar", "baz", "foo"})
	c.Assert(union2.values(), jc.DeepEquals, []string{"bang", "bar", "baz", "foo"})
}

func (suite *stringSetSuite) TestDifference(c *gc.C) {
	s1 := make(stringsSet)
	s1.add("foo")
	s1.add("bar")
	s2 := make(stringsSet)
	s2.add("foo")
	s2.add("baz")
	s2.add("bang")
	diff1 := s1.difference(s2)
	diff2 := s2.difference(s1)

	c.Assert(diff1.values(), jc.DeepEquals, []string{"bar"})
	c.Assert(diff2.values(), jc.DeepEquals, []string{"bang", "baz"})
}
