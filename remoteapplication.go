// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"gopkg.in/juju/names.v2"
)

// RemoteApplication represents an application in another model that
// can participate in a relation in this model.
type RemoteApplication interface {
	Tag() names.ApplicationTag
	Name() string
	OfferName() string
	URL() string
	SourceModelTag() names.ModelTag
	Registered() bool

	Endpoints() []RemoteEndpoint
	AddEndpoint(RemoteEndpointArgs) RemoteEndpoint
}

type remoteApplications struct {
	Version            int                  `yaml:"version"`
	RemoteApplications []*remoteApplication `yaml:"remote-applications"`
}

type remoteApplication struct {
	Name_            string          `yaml:"name"`
	OfferName_       string          `yaml:"offer-name"`
	URL_             string          `yaml:"url"`
	SourceModelUUID_ string          `yaml:"source-model-uuid"`
	Endpoints_       remoteEndpoints `yaml:"endpoints,omitempty"`
	Registered_      bool            `yaml:"registered,omitempty"`
}

// RemoteApplicationArgs is an argument struct used to add a remote
// application to the Model.
type RemoteApplicationArgs struct {
	Tag         names.ApplicationTag
	OfferName   string
	URL         string
	SourceModel names.ModelTag
	Registered  bool
}

func newRemoteApplication(args RemoteApplicationArgs) *remoteApplication {
	a := &remoteApplication{
		Name_:            args.Tag.Id(),
		OfferName_:       args.OfferName,
		URL_:             args.URL,
		SourceModelUUID_: args.SourceModel.Id(),
		Registered_:      args.Registered,
	}
	a.setEndpoints(nil)
	return a
}

// Tag implements RemoteApplication.
func (a *remoteApplication) Tag() names.ApplicationTag {
	return names.NewApplicationTag(a.Name_)
}

// Name implements RemoteApplication.
func (a *remoteApplication) Name() string {
	return a.Name_
}

// OfferName implements RemoteApplication.
func (a *remoteApplication) OfferName() string {
	return a.OfferName_
}

// URL implements RemoteApplication.
func (a *remoteApplication) URL() string {
	return a.URL_
}

// SourceModelTag implements RemoteApplication.
func (a *remoteApplication) SourceModelTag() names.ModelTag {
	return names.NewModelTag(a.SourceModelUUID_)
}

// Registered implements RemoteApplication.
func (a *remoteApplication) Registered() bool {
	return a.Registered_
}

// Endpoints implements RemoteApplication.
func (a *remoteApplication) Endpoints() []RemoteEndpoint {
	result := make([]RemoteEndpoint, len(a.Endpoints_.Endpoints))
	for i, endpoint := range a.Endpoints_.Endpoints {
		result[i] = endpoint
	}
	return result
}

// AddEndpoint implements RemoteApplication.
func (a *remoteApplication) AddEndpoint(args RemoteEndpointArgs) RemoteEndpoint {
	ep := newRemoteEndpoint(args)
	a.Endpoints_.Endpoints = append(a.Endpoints_.Endpoints, ep)
	return ep
}

func (a *remoteApplication) setEndpoints(endpointList []*remoteEndpoint) {
	a.Endpoints_ = remoteEndpoints{
		Version:   1,
		Endpoints: endpointList,
	}
}

func importRemoteApplications(sourceMap map[string]interface{}) ([]*remoteApplication, error) {
	// TODO(babbageclunk): implement importing remote applications
	if val, ok := sourceMap["remote-applications"]; ok {
		if items, ok := val.([]interface{}); ok && len(items) != 0 {
			return nil, errors.New("importing remote applications is not supported yet")
		}
	}
	return nil, nil
}
