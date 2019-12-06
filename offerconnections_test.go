// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type OfferConnectionSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&OfferConnectionSerializationSuite{})

func (s *OfferConnectionSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "offer connections"
	s.sliceName = "offer-connections"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importOfferConnections(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["offer-connections"] = []interface{}{}
	}
}

func (s *OfferConnectionSerializationSuite) TestNewOfferConnection(c *gc.C) {
	offer := newOfferConnection(OfferConnectionArgs{
		OfferUUID:       "offer-uuid",
		RelationID:      1,
		RelationKey:     "relation-key",
		SourceModelUUID: "source-model-uuid",
		UserName:        "fred",
	})

	c.Check(offer.OfferUUID(), gc.Equals, "offer-uuid")
	c.Check(offer.RelationID(), gc.Equals, 1)
	c.Check(offer.RelationKey(), gc.Equals, "relation-key")
	c.Check(offer.SourceModelUUID(), gc.Equals, "source-model-uuid")
	c.Check(offer.UserName(), gc.Equals, "fred")
}

func (s *OfferConnectionSerializationSuite) TestParsingSerializedData(c *gc.C) {
	initial := newOfferConnection(OfferConnectionArgs{
		OfferUUID:       "offer-uuid",
		RelationID:      1,
		RelationKey:     "relation-key",
		SourceModelUUID: "source-model-uuid",
		UserName:        "fred",
	})

	offer := s.exportImportLatest(c, initial)
	c.Assert(offer, jc.DeepEquals, initial)
}

func (s *OfferConnectionSerializationSuite) exportImportLatest(c *gc.C, offer *offerConnection) *offerConnection {
	return s.exportImportVersion(c, offer, 1)
}

func (s *OfferConnectionSerializationSuite) exportImportVersion(c *gc.C, offer_ *offerConnection, version int) *offerConnection {
	initial := &offerConnections{
		Version:          version,
		OfferConnections: []*offerConnection{offer_},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	offers, err := importOfferConnections(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(offers, gc.HasLen, 1)
	return offers[0]
}
