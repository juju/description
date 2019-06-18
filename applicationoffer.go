// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// ApplicationOffer represents an offer for a an application's endpoints.
type ApplicationOffer interface {
	OfferName() string
	Endpoints() []string
}

var _ ApplicationOffer = (*applicationOffer)(nil)

type applicationOffers struct {
	Version int                 `yaml:"version"`
	Offers  []*applicationOffer `yaml:"offers"`
}

type applicationOffer struct {
	OfferName_ string   `yaml:"offer-name"`
	Endpoints_ []string `yaml:"endpoints"`
}

func (o *applicationOffer) OfferName() string {
	return o.OfferName_
}

func (o *applicationOffer) Endpoints() []string {
	return o.Endpoints_
}

// ApplicationOfferArgs is an argument struct used to instanciate a new
// applicationOffer instance that implements ApplicationOffer.
type ApplicationOfferArgs struct {
	OfferName string
	Endpoints []string
}

func newApplicationOffer(args ApplicationOfferArgs) *applicationOffer {
	return &applicationOffer{
		OfferName_: args.OfferName,
		Endpoints_: args.Endpoints,
	}
}

func importApplicationOffers(source map[string]interface{}) ([]*applicationOffer, error) {
	checker := versionedChecker("offers")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "offers version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := applicationOfferDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	sourceList := valid["offers"].([]interface{})
	return importApplicationOfferList(sourceList, importFunc)
}

func importApplicationOfferList(sourceList []interface{}, importFunc applicationOfferDeserializationFunc) ([]*applicationOffer, error) {
	result := make([]*applicationOffer, 0, len(sourceList))

	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for application offer %d, %T", i, value)
		}

		offer, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "application offer %d", i)
		}
		result = append(result, offer)
	}
	return result, nil
}

type applicationOfferDeserializationFunc func(interface{}) (*applicationOffer, error)

var applicationOfferDeserializationFuncs = map[int]applicationOfferDeserializationFunc{
	1: importApplicationOfferV1,
}

func importApplicationOfferV1(source interface{}) (*applicationOffer, error) {
	fields := schema.Fields{
		"offer-name": schema.String(),
		"endpoints":  schema.List(schema.String()),
	}
	checker := schema.FieldMap(fields, nil)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "application offer v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	validEndpoints := valid["endpoints"].([]interface{})
	endpoints := make([]string, len(validEndpoints))
	for i, ep := range validEndpoints {
		endpoints[i] = ep.(string)
	}

	return &applicationOffer{
		OfferName_: valid["offer-name"].(string),
		Endpoints_: endpoints,
	}, nil
}
