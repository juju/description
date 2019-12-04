// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"sort"

	"github.com/juju/errors"
	"github.com/juju/schema"
)

// ApplicationOffer represents an offer for a an application's endpoints.
type ApplicationOffer interface {
	OfferUUID() string
	OfferName() string
	Endpoints() []string
	EndpointsMap() map[string]string
	ACL() map[string]string
	ApplicationName() string
	ApplicationDescription() string
}

var _ ApplicationOffer = (*applicationOffer)(nil)

type applicationOffers struct {
	Version int                 `yaml:"version"`
	Offers  []*applicationOffer `yaml:"offers,omitempty"`
}

type applicationOffer struct {
	OfferUUID_              string            `yaml:"offer-uuid,omitempty"`
	OfferName_              string            `yaml:"offer-name"`
	Endpoints_              []string          `yaml:"endpoints,omitempty"`
	EndpointsMap_           map[string]string `yaml:"endpoints-map,omitempty"`
	ACL_                    map[string]string `yaml:"acl,omitempty"`
	ApplicationName_        string            `yaml:"application-name,omitempty"`
	ApplicationDescription_ string            `yaml:"application-description,omitempty"`
}

// OfferUUID returns the underlying offer UUID.
// The offer UUID is required when migrating a CMR model between controllers.
func (o *applicationOffer) OfferUUID() string {
	return o.OfferUUID_
}

// OfferName implements ApplicationOffer.
func (o *applicationOffer) OfferName() string {
	return o.OfferName_
}

// Endpoints implements ApplicationOffer.
func (o *applicationOffer) Endpoints() []string {
	return o.Endpoints_
}

// EndpointsMap returns the representation of both the internal and external
// endpoints. This is useful for CMR migration, where we need to match internal
// offers when importing.
func (o *applicationOffer) EndpointsMap() map[string]string {
	return o.EndpointsMap_
}

// ACL implements ApplicationOffer. It returns a map were keys are users and
// values are access permissions.
func (o *applicationOffer) ACL() map[string]string {
	return o.ACL_
}

// ApplicationName returns the ApplicationName for CMR model migration to happen.
func (o *applicationOffer) ApplicationName() string {
	return o.ApplicationName_
}

// ApplicationDescription returns the ApplicationDescription for CMR model migration to happen.
func (o *applicationOffer) ApplicationDescription() string {
	return o.ApplicationDescription_
}

// ApplicationOfferArgs is an argument struct used to instanciate a new
// applicationOffer instance that implements ApplicationOffer.
type ApplicationOfferArgs struct {
	OfferUUID              string
	OfferName              string
	EndpointsMap           map[string]string
	ACL                    map[string]string
	ApplicationName        string
	ApplicationDescription string
}

func newApplicationOffer(args ApplicationOfferArgs) *applicationOffer {
	// To ensure that there isn't a split brain scenario for Endpoints and
	// EndpointsMap, we populate the Endpoints from the Endpoints, to be
	// backwards compatible.
	endpoints := mapValuesToSlice(args.EndpointsMap)
	return &applicationOffer{
		OfferUUID_:              args.OfferUUID,
		OfferName_:              args.OfferName,
		Endpoints_:              endpoints,
		EndpointsMap_:           args.EndpointsMap,
		ACL_:                    args.ACL,
		ApplicationName_:        args.ApplicationName,
		ApplicationDescription_: args.ApplicationDescription,
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
	2: importApplicationOfferV2,
}

func applicationOfferV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"offer-name": schema.String(),
		"endpoints":  schema.List(schema.String()),
		"acl":        schema.Map(schema.String(), schema.String()),
	}
	return fields, schema.Defaults{}
}

func applicationOfferV2Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := applicationOfferV1Fields()
	fields["offer-uuid"] = schema.String()
	fields["application-name"] = schema.String()
	fields["application-description"] = schema.String()
	fields["endpoints-map"] = schema.Map(schema.String(), schema.String())

	defaults["application-description"] = schema.Omit
	defaults["endpoints"] = schema.Omit

	return fields, defaults
}

func importApplicationOffer(fields schema.Fields, defaults schema.Defaults, importVersion int, source interface{}) (*applicationOffer, error) {
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "application offer v%d schema check failed", importVersion)
	}
	valid := coerced.(map[string]interface{})

	validACL := valid["acl"].(map[interface{}]interface{})
	aclMap := make(map[string]string, len(validACL))
	for user, access := range validACL {
		aclMap[user.(string)] = access.(string)
	}

	offer := &applicationOffer{
		OfferName_: valid["offer-name"].(string),
		ACL_:       aclMap,
	}

	// Manage how we handle endpoints.
	if importVersion == 1 {
		// When importing version 1 of the description, we should just treat
		// endpoints as a slice string.
		validEndpoints := valid["endpoints"].([]interface{})
		endpoints := make([]string, len(validEndpoints))
		for i, ep := range validEndpoints {
			endpoints[i] = ep.(string)
		}
		offer.Endpoints_ = endpoints
	}

	if importVersion >= 2 {
		offer.OfferUUID_ = valid["offer-uuid"].(string)
		offer.ApplicationName_ = valid["application-name"].(string)
		offer.ApplicationDescription_ = valid["application-description"].(string)

		// When importing version 2 or greater of the description, we should
		// ensure that we use EndpointsMap is used, then also populate endpoints
		// as well, so that we remain consistent.
		validEndpointsMap := valid["endpoints-map"].(map[interface{}]interface{})
		endpointsMap := make(map[string]string, len(validEndpointsMap))
		for k, ep := range validEndpointsMap {
			endpointsMap[k.(string)] = ep.(string)
		}
		offer.EndpointsMap_ = endpointsMap
		// Backfill the endpoints from the EndpointsMap
		offer.Endpoints_ = mapValuesToSlice(endpointsMap)
	}

	return offer, nil
}

func importApplicationOfferV1(source interface{}) (*applicationOffer, error) {
	fields, defaults := applicationOfferV1Fields()
	return importApplicationOffer(fields, defaults, 1, source)
}

func importApplicationOfferV2(source interface{}) (*applicationOffer, error) {
	fields, defaults := applicationOfferV2Fields()
	return importApplicationOffer(fields, defaults, 2, source)
}

func mapValuesToSlice(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for _, ep := range m {
		result = append(result, ep)
	}
	sort.Strings(result)
	return result
}
