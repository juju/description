// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

// ApplicationOffer represents an offer for a an application's endpoints.
type ApplicationOffer interface {
	OfferName() string
	Endpoints() []string
}

var _ ApplicationOffer = (*applicationOffer)(nil)

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
