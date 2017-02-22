// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"testing"

	jtesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

// Constraints and CloudInstance store megabytes
const gig uint64 = 1024

// None of the tests in this package require mongo.
func TestPackage(t *testing.T) {
	gc.TestingT(t)
}

type ImportTest struct{}

var _ = gc.Suite(&ImportTest{})

func (*ImportTest) TestImports(c *gc.C) {
	imps, err := jtesting.FindImports(
		"github.com/juju/description",
		"github.com/juju/juju/")
	c.Assert(err, jc.ErrorIsNil)
	// This package brings in nothing else from juju/juju
	c.Assert(imps, gc.HasLen, 0)
}
