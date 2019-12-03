// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	yaml "gopkg.in/yaml.v2"
)

type ApplicationOfferSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&ApplicationOfferSerializationSuite{})

func (s *ApplicationOfferSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "offers"
	s.sliceName = "offers"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importApplicationOffers(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["offers"] = []interface{}{}
	}
}

func (s *ApplicationOfferSerializationSuite) TestNewApplicationOfferV1(c *gc.C) {
	offer := newApplicationOffer(ApplicationOfferArgs{
		OfferName: "my-offer",
		Endpoints: []string{"endpoint-1", "endpoint-2"},
		ACL: map[string]string{
			"admin": "admin",
			"foo":   "read",
			"bar":   "consume",
		},
	})

	c.Check(offer.OfferName(), gc.Equals, "my-offer")
	c.Check(offer.Endpoints(), gc.DeepEquals, []string{"endpoint-1", "endpoint-2"})
	c.Check(offer.ACL(), gc.DeepEquals, map[string]string{
		"admin": "admin",
		"foo":   "read",
		"bar":   "consume",
	})
}

func (s *ApplicationOfferSerializationSuite) TestNewApplicationOfferV2(c *gc.C) {
	offer := newApplicationOffer(ApplicationOfferArgs{
		OfferUUID: "offer-uuid",
		OfferName: "my-offer",
		Endpoints: []string{"endpoint-1", "endpoint-2"},
		ACL: map[string]string{
			"admin": "admin",
			"foo":   "read",
			"bar":   "consume",
		},
		ApplicationName:        "foo",
		ApplicationDescription: "foo description",
	})

	c.Check(offer.OfferUUID(), gc.Equals, "offer-uuid")
	c.Check(offer.OfferName(), gc.Equals, "my-offer")
	c.Check(offer.Endpoints(), gc.DeepEquals, []string{"endpoint-1", "endpoint-2"})
	c.Check(offer.ACL(), gc.DeepEquals, map[string]string{
		"admin": "admin",
		"foo":   "read",
		"bar":   "consume",
	})
	c.Check(offer.ApplicationName(), gc.Equals, "foo")
	c.Check(offer.ApplicationDescription(), gc.Equals, "foo description")
}

func (s *ApplicationOfferSerializationSuite) TestNewApplicationOfferV2WithOptionalFields(c *gc.C) {
	offer := newApplicationOffer(ApplicationOfferArgs{
		OfferUUID: "offer-uuid",
		OfferName: "my-offer",
		Endpoints: []string{"endpoint-1", "endpoint-2"},
		ACL: map[string]string{
			"admin": "admin",
			"foo":   "read",
			"bar":   "consume",
		},
		ApplicationName: "foo",
	})

	c.Check(offer.OfferUUID(), gc.Equals, "offer-uuid")
	c.Check(offer.OfferName(), gc.Equals, "my-offer")
	c.Check(offer.Endpoints(), gc.DeepEquals, []string{"endpoint-1", "endpoint-2"})
	c.Check(offer.ACL(), gc.DeepEquals, map[string]string{
		"admin": "admin",
		"foo":   "read",
		"bar":   "consume",
	})
	c.Check(offer.ApplicationName(), gc.Equals, "foo")
	c.Check(offer.ApplicationDescription(), gc.Equals, "")
}

func (s *ApplicationOfferSerializationSuite) TestParsingSerializedDataV1(c *gc.C) {
	initial := newApplicationOffer(ApplicationOfferArgs{
		OfferName: "my-offer",
		Endpoints: []string{"endpoint-1", "endpoint-2"},
		ACL: map[string]string{
			"admin": "admin",
			"foo":   "read",
			"bar":   "consume",
		},
	})
	offer := s.exportImportLatestV1(c, initial)
	c.Assert(offer, jc.DeepEquals, initial)
}

func (s *ApplicationOfferSerializationSuite) TestParsingSerializedDataV2(c *gc.C) {
	initial := newApplicationOffer(ApplicationOfferArgs{
		OfferUUID: "offer-uuid",
		OfferName: "my-offer",
		Endpoints: []string{"endpoint-1", "endpoint-2"},
		ACL: map[string]string{
			"admin": "admin",
			"foo":   "read",
			"bar":   "consume",
		},
		ApplicationName:        "foo",
		ApplicationDescription: "foo description",
	})
	offer := s.exportImportLatestV2(c, initial)
	c.Assert(offer, jc.DeepEquals, initial)
}

func (s *ApplicationOfferSerializationSuite) exportImportLatestV1(c *gc.C, offer *applicationOffer) *applicationOffer {
	return s.exportImportVersion(c, offer, 1)
}

func (s *ApplicationOfferSerializationSuite) exportImportLatestV2(c *gc.C, offer *applicationOffer) *applicationOffer {
	return s.exportImportVersion(c, offer, 2)
}

func (s *ApplicationOfferSerializationSuite) exportImportVersion(c *gc.C, offer_ *applicationOffer, version int) *applicationOffer {
	initial := &applicationOffers{
		Version: version,
		Offers:  []*applicationOffer{offer_},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	offers, err := importApplicationOffers(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(offers, gc.HasLen, 1)
	return offers[0]
}

func minimalApplicationOfferMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"offer-uuid": "offer-uuid",
		"offer-name": "my-offer",
		"endpoints":  []interface{}{"endpoint-1", "endpoint-2"},
		"acl": map[interface{}]interface{}{
			"admin": "admin",
			"foo":   "read",
			"bar":   "consume",
		},
		"application-name":        "foo",
		"application-description": "foo description",
	}
}
