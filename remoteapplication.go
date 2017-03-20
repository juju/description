// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
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
	IsConsumerProxy() bool

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
	IsConsumerProxy_ bool            `yaml:"is-consumer-proxy,omitempty"`
}

// RemoteApplicationArgs is an argument struct used to add a remote
// application to the Model.
type RemoteApplicationArgs struct {
	Tag             names.ApplicationTag
	OfferName       string
	URL             string
	SourceModel     names.ModelTag
	IsConsumerProxy bool
}

func newRemoteApplication(args RemoteApplicationArgs) *remoteApplication {
	a := &remoteApplication{
		Name_:            args.Tag.Id(),
		OfferName_:       args.OfferName,
		URL_:             args.URL,
		SourceModelUUID_: args.SourceModel.Id(),
		IsConsumerProxy_: args.IsConsumerProxy,
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

// IsConsumerProxy implements RemoteApplication.
func (a *remoteApplication) IsConsumerProxy() bool {
	return a.IsConsumerProxy_
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

func importRemoteApplications(source interface{}) ([]*remoteApplication, error) {
	checker := versionedChecker("remote-applications")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "remote applications version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := remoteApplicationDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["remote-applications"].([]interface{})
	return importRemoteApplicationList(sourceList, importFunc)
}

func importRemoteApplicationList(sourceList []interface{}, importFunc remoteApplicationDeserializationFunc) ([]*remoteApplication, error) {
	result := make([]*remoteApplication, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for remote application %d, %T", i, value)
		}
		device, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "remote application %d", i)
		}
		result = append(result, device)
	}
	return result, nil
}

type remoteApplicationDeserializationFunc func(map[string]interface{}) (*remoteApplication, error)

var remoteApplicationDeserializationFuncs = map[int]remoteApplicationDeserializationFunc{
	1: importRemoteApplicationV1,
}

func importRemoteApplicationV1(source map[string]interface{}) (*remoteApplication, error) {
	fields := schema.Fields{
		"name":              schema.String(),
		"offer-name":        schema.String(),
		"url":               schema.String(),
		"source-model-uuid": schema.String(),
		"endpoints":         schema.StringMap(schema.Any()),
		"is-consumer-proxy": schema.Bool(),
	}

	defaults := schema.Defaults{
		"endpoints":         schema.Omit,
		"is-consumer-proxy": false,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "remote application v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	result := &remoteApplication{
		Name_:            valid["name"].(string),
		OfferName_:       valid["offer-name"].(string),
		URL_:             valid["url"].(string),
		SourceModelUUID_: valid["source-model-uuid"].(string),
		IsConsumerProxy_: valid["is-consumer-proxy"].(bool),
	}

	if rawEndpoints, ok := valid["endpoints"]; ok {
		endpointsMap := rawEndpoints.(map[string]interface{})
		endpoints, err := importRemoteEndpoints(endpointsMap)
		if err != nil {
			return nil, errors.Trace(err)
		}
		result.setEndpoints(endpoints)
	}

	return result, nil
}
