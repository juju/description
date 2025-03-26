// Copyright 2025 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/testing"
	gc "gopkg.in/check.v1"
)

type versionSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&versionSuite{})

var parseTests = []struct {
	v  string
	ok bool
}{{
	v:  "0.0.0",
	ok: true,
}, {
	v:  "0.0.1",
	ok: true,
}, {
	v:  "0.1.0",
	ok: true,
}, {
	v:  "0.2.3",
	ok: true,
}, {
	v:  "1.0.0",
	ok: true,
}, {
	v:  "1.21.0",
	ok: true,
}, {
	v:  "10.234.3456",
	ok: true,
}, {
	v:  "10.234.3456.1",
	ok: true,
}, {
	v:  "1.21-alpha1",
	ok: true,
}, {
	v:  "1.21-alpha1.1",
	ok: true,
}, {
	v:  "1.21-alpha10",
	ok: true,
}, {
	v:  "1234567890.2.1",
	ok: false,
}, {
	v:  "0.2..1",
	ok: false,
}, {
	v:  "1.21.alpha1",
	ok: false,
}, {
	v:  "1.21-alpha",
	ok: false,
}, {
	v:  "1.21-alpha1beta",
	ok: false,
}, {
	v:  "1.21-alpha-dev",
	ok: false,
}, {
	v:  "1.21-alpha_dev3",
	ok: false,
}, {
	v:  "1.21-alpha123dev3",
	ok: false,
}}

func (*versionSuite) TestParse(c *gc.C) {
	for i, test := range parseTests {
		c.Logf("test %d: %q", i, test.v)
		c.Check(numberPat.MatchString(test.v), gc.Equals, test.ok)
	}
}
