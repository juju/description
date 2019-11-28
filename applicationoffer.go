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
	ACL() map[string]string
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
	Endpoints              []string
	ACL                    map[string]string
	ApplicationName        string
	ApplicationDescription string
}

func newApplicationOffer(args ApplicationOfferArgs) *applicationOffer {
	return &applicationOffer{
		OfferUUID_:              args.OfferUUID,
		OfferName_:              args.OfferName,
		Endpoints_:              args.Endpoints,
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
	return fields, defaults
}

func importApplicationOffer(fields schema.Fields, defaults schema.Defaults, importVersion int, source interface{}) (*applicationOffer, error) {
	checker := schema.FieldMap(fields, defaults)

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

	validACL := valid["acl"].(map[interface{}]interface{})
	aclMap := make(map[string]string, len(validACL))
	for user, access := range validACL {
		aclMap[user.(string)] = access.(string)
	}

	offer := &applicationOffer{
		OfferName_: valid["offer-name"].(string),
		Endpoints_: endpoints,
		ACL_:       aclMap,
	}

	if importVersion >= 2 {
		offer.OfferUUID_ = valid["offer-uuid"].(string)
		offer.ApplicationName_ = valid["application-name"].(string)
		offer.ApplicationDescription_ = valid["application-description"].(string)
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
