// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	gc "gopkg.in/check.v1"
)

type ApplicationOfferSerializationSuite struct {
}

var _ = gc.Suite(&ApplicationOfferSerializationSuite{})

func (s *ApplicationOfferSerializationSuite) TestNewApplicationOffer(c *gc.C) {
	offer := newApplicationOffer(ApplicationOfferArgs{
		OfferName: "my-offer",
		Endpoints: []string{"endpoint-1", "endpoint-2"},
	})

	c.Check(offer.OfferName(), gc.Equals, "my-offer")
	c.Check(offer.Endpoints(), gc.DeepEquals, []string{"endpoint-1", "endpoint-2"})
}
