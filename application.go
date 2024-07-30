// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"encoding/base64"

	"github.com/juju/collections/set"
	"github.com/juju/errors"
	"github.com/juju/names/v5"
	"github.com/juju/schema"
)

// Application represents a deployed charm in a model.
type Application interface {
	HasAnnotations
	HasConstraints
	HasOperatorStatus
	HasStatus
	HasStatusHistory

	Tag() names.ApplicationTag
	Name() string
	Type() string
	Subordinate() bool
	CharmURL() string
	Channel() string
	CharmModifiedVersion() int
	ForceCharm() bool
	MinUnits() int

	Exposed() bool
	ExposedEndpoints() map[string]ExposedEndpoint

	PasswordHash() string
	PodSpec() string
	DesiredScale() int
	Placement() string
	HasResources() bool
	CloudService() CloudService
	SetCloudService(CloudServiceArgs)

	EndpointBindings() map[string]string

	CharmConfig() map[string]interface{}
	ApplicationConfig() map[string]interface{}

	Leader() string
	LeadershipSettings() map[string]interface{}

	MetricsCredentials() []byte
	StorageDirectives() map[string]StorageDirective

	Resources() []Resource
	AddResource(ResourceArgs) Resource

	Units() []Unit
	AddUnit(UnitArgs) Unit

	CharmOrigin() CharmOrigin
	SetCharmOrigin(CharmOriginArgs)

	CharmMetadata() CharmMetadata
	SetCharmMetadata(CharmMetadataArgs)

	CharmManifest() CharmManifest
	SetCharmManifest(CharmManifestArgs)

	Tools() AgentTools
	SetTools(AgentToolsArgs)

	Offers() []ApplicationOffer
	AddOffer(ApplicationOfferArgs) ApplicationOffer

	Validate() error

	OpenedPortRanges() PortRanges
	AddOpenedPortRange(OpenedPortRangeArgs)

	ProvisioningState() ProvisioningState
}

// ExposedEndpoint encapsulates the details about the CIDRs and/or spaces that
// should be able to access ports opened by the application for a particular
// endpoint once the application is exposed.
type ExposedEndpoint interface {
	ExposeToSpaceIDs() []string
	ExposeToCIDRs() []string
}

type applications struct {
	Version       int            `yaml:"version"`
	Applications_ []*application `yaml:"applications"`
}

type application struct {
	Name_ string `yaml:"name"`
	Type_ string `yaml:"type"`
	// Series obsolete from v9. Retained for tests.
	Series_               string `yaml:"series,omitempty"`
	Subordinate_          bool   `yaml:"subordinate,omitempty"`
	CharmURL_             string `yaml:"charm-url"`
	Channel_              string `yaml:"cs-channel"`
	CharmModifiedVersion_ int    `yaml:"charm-mod-version"`

	// ForceCharm is true if an upgrade charm is forced.
	// It means upgrade even if the charm is in an error state.
	ForceCharm_ bool `yaml:"force-charm,omitempty"`
	MinUnits_   int  `yaml:"min-units,omitempty"`

	Exposed_          bool                        `yaml:"exposed,omitempty"`
	ExposedEndpoints_ map[string]*exposedEndpoint `yaml:"exposed-endpoints,omitempty"`

	Status_        *status `yaml:"status"`
	StatusHistory_ `yaml:"status-history"`

	EndpointBindings_ map[string]string `yaml:"endpoint-bindings,omitempty"`

	// CharmConfig_ and ApplicationConfig_ are the actual configuration values
	// for the charm and application, respectively. These are the values that
	// have been set by the user or the charm itself.
	CharmConfig_       map[string]interface{} `yaml:"settings"`
	ApplicationConfig_ map[string]interface{} `yaml:"application-config,omitempty"`

	Leader_             string                 `yaml:"leader,omitempty"`
	LeadershipSettings_ map[string]interface{} `yaml:"leadership-settings"`

	MetricsCredentials_ string `yaml:"metrics-creds,omitempty"`

	// unit count will be assumed by the number of units associated.
	Units_ units `yaml:"units"`

	Resources_ resources `yaml:"resources"`

	Annotations_ `yaml:"annotations,omitempty"`

	Constraints_       *constraints                 `yaml:"constraints,omitempty"`
	StorageDirectives_ map[string]*storageDirective `yaml:"storage-directives,omitempty"`

	// CAAS application fields.
	PasswordHash_      string             `yaml:"password-hash,omitempty"`
	PodSpec_           string             `yaml:"pod-spec,omitempty"`
	Placement_         string             `yaml:"placement,omitempty"`
	HasResources_      bool               `yaml:"has-resources,omitempty"`
	DesiredScale_      int                `yaml:"desired-scale,omitempty"`
	CloudService_      *cloudService      `yaml:"cloud-service,omitempty"`
	Tools_             *agentTools        `yaml:"tools,omitempty"`
	OperatorStatus_    *status            `yaml:"operator-status,omitempty"`
	ProvisioningState_ *provisioningState `yaml:"provisioning-state,omitempty"`

	OpenedPortRanges_ *deployedPortRanges `yaml:"opened-port-ranges,omitempty"`

	// Offer-related fields
	Offers_ *applicationOffers `yaml:"offers,omitempty"`

	// CharmOrigin fields
	CharmOrigin_ *charmOrigin `yaml:"charm-origin,omitempty"`

	// The following fields represent the actual charm data for the
	// application. These are the immutable parts of the application, either
	// provided by the charm itself.
	CharmMetadata_ *charmMetadata `yaml:"charm-metadata,omitempty"`
	CharmManifest_ *charmManifest `yaml:"charm-manifest,omitempty"`
	CharmActions_  *charmActions  `yaml:"charm-actions,omitempty"`
	CharmConfigs_  *charmConfigs  `yaml:"charm-configs,omitempty"`
}

// ApplicationArgs is an argument struct used to add an application to the Model.
type ApplicationArgs struct {
	Tag  names.ApplicationTag
	Type string
	// Series obsolete from v9. Retained for tests.
	Series               string
	Subordinate          bool
	CharmURL             string
	Channel              string
	CharmModifiedVersion int
	ForceCharm           bool
	PasswordHash         string
	PodSpec              string
	Placement            string
	HasResources         bool
	DesiredScale         int
	CloudService         *CloudServiceArgs
	MinUnits             int
	Exposed              bool
	ExposedEndpoints     map[string]ExposedEndpointArgs
	EndpointBindings     map[string]string
	ApplicationConfig    map[string]interface{}
	CharmConfig          map[string]interface{}
	Leader               string
	LeadershipSettings   map[string]interface{}
	StorageDirectives    map[string]StorageDirectiveArgs
	MetricsCredentials   []byte
	ProvisioningState    *ProvisioningStateArgs
}

func newApplication(args ApplicationArgs) *application {
	creds := base64.StdEncoding.EncodeToString(args.MetricsCredentials)
	app := &application{
		Name_:                 args.Tag.Id(),
		Type_:                 args.Type,
		Series_:               args.Series,
		Subordinate_:          args.Subordinate,
		CharmURL_:             args.CharmURL,
		Channel_:              args.Channel,
		CharmModifiedVersion_: args.CharmModifiedVersion,
		ForceCharm_:           args.ForceCharm,
		Exposed_:              args.Exposed,
		PasswordHash_:         args.PasswordHash,
		PodSpec_:              args.PodSpec,
		CloudService_:         newCloudService(args.CloudService),
		Placement_:            args.Placement,
		HasResources_:         args.HasResources,
		DesiredScale_:         args.DesiredScale,
		MinUnits_:             args.MinUnits,
		EndpointBindings_:     args.EndpointBindings,
		ApplicationConfig_:    args.ApplicationConfig,
		CharmConfig_:          args.CharmConfig,
		Leader_:               args.Leader,
		LeadershipSettings_:   args.LeadershipSettings,
		MetricsCredentials_:   creds,
		StatusHistory_:        newStatusHistory(),
		ProvisioningState_:    newProvisioningState(args.ProvisioningState),
	}
	app.setUnits(nil)
	app.setResources(nil)
	if len(args.StorageDirectives) > 0 {
		app.StorageDirectives_ = make(map[string]*storageDirective)
		for key, value := range args.StorageDirectives {
			app.StorageDirectives_[key] = newStorageDirective(value)
		}
	}
	if len(args.ExposedEndpoints) > 0 {
		app.ExposedEndpoints_ = make(map[string]*exposedEndpoint)
		for key, value := range args.ExposedEndpoints {
			app.ExposedEndpoints_[key] = newExposedEndpoint(value)
		}
	}
	return app
}

// Tag implements Application.
func (a *application) Tag() names.ApplicationTag {
	return names.NewApplicationTag(a.Name_)
}

// Name implements Application.
func (a *application) Name() string {
	return a.Name_
}

// Type implements Application
func (a *application) Type() string {
	return a.Type_
}

// Subordinate implements Application.
func (a *application) Subordinate() bool {
	return a.Subordinate_
}

// CharmURL implements Application.
func (a *application) CharmURL() string {
	return a.CharmURL_
}

// Channel implements Application.
func (a *application) Channel() string {
	return a.Channel_
}

// CharmModifiedVersion implements Application.
func (a *application) CharmModifiedVersion() int {
	return a.CharmModifiedVersion_
}

// ForceCharm implements Application.
func (a *application) ForceCharm() bool {
	return a.ForceCharm_
}

// Exposed implements Application.
func (a *application) Exposed() bool {
	return a.Exposed_
}

// ExposedEndpoints implements Application.
func (a *application) ExposedEndpoints() map[string]ExposedEndpoint {
	if len(a.ExposedEndpoints_) == 0 {
		return nil
	}

	result := make(map[string]ExposedEndpoint)
	for key, value := range a.ExposedEndpoints_ {
		result[key] = value
	}
	return result
}

// PasswordHash implements Application.
func (a *application) PasswordHash() string {
	return a.PasswordHash_
}

// PodSpec implements Application.
func (a *application) PodSpec() string {
	return a.PodSpec_
}

// Placement implements Application.
func (a *application) Placement() string {
	return a.Placement_
}

// HasResources implements Application.
func (a *application) HasResources() bool {
	return a.HasResources_
}

// DesiredScale implements Application.
func (a *application) DesiredScale() int {
	return a.DesiredScale_
}

// MinUnits implements Application.
func (a *application) MinUnits() int {
	return a.MinUnits_
}

// EndpointBindings implements Application.
func (a *application) EndpointBindings() map[string]string {
	return a.EndpointBindings_
}

// ApplicationConfig implements Application.
func (a *application) ApplicationConfig() map[string]interface{} {
	return a.ApplicationConfig_
}

// CharmConfig implements Application.
func (a *application) CharmConfig() map[string]interface{} {
	return a.CharmConfig_
}

// Leader implements Application.
func (a *application) Leader() string {
	return a.Leader_
}

// LeadershipSettings implements Application.
func (a *application) LeadershipSettings() map[string]interface{} {
	return a.LeadershipSettings_
}

// StorageDirectives implements Application.
func (a *application) StorageDirectives() map[string]StorageDirective {
	result := make(map[string]StorageDirective)
	for key, value := range a.StorageDirectives_ {
		result[key] = value
	}
	return result
}

// MetricsCredentials implements Application.
func (a *application) MetricsCredentials() []byte {
	// Here we are explicitly throwing away any decode error. We check that
	// the creds can be decoded when we parse the incoming data, or we encode
	// an incoming byte array, so in both cases, we know that the stored creds
	// can be decoded.
	creds, _ := base64.StdEncoding.DecodeString(a.MetricsCredentials_)
	return creds
}

// OpenedPortRanges implements Application.
func (a *application) OpenedPortRanges() PortRanges {
	if a.OpenedPortRanges_ == nil {
		a.OpenedPortRanges_ = newDeployedPortRanges()
	}
	return a.OpenedPortRanges_
}

// AddOpenedPortRange implements Application.
func (a *application) AddOpenedPortRange(args OpenedPortRangeArgs) {
	if a.OpenedPortRanges_ == nil {
		a.OpenedPortRanges_ = newDeployedPortRanges()
	}

	if a.OpenedPortRanges_.ByUnit_[args.UnitName] == nil {
		a.OpenedPortRanges_.ByUnit_[args.UnitName] = newUnitPortRanges()
	}

	a.OpenedPortRanges_.ByUnit_[args.UnitName].ByEndpoint_[args.EndpointName] = append(
		a.OpenedPortRanges_.ByUnit_[args.UnitName].ByEndpoint_[args.EndpointName],
		newUnitPortRange(args.FromPort, args.ToPort, args.Protocol),
	)
}

// OperatorStatus implements Application.
func (a *application) OperatorStatus() Status {
	// To avoid typed nils check nil here.
	if a.OperatorStatus_ == nil {
		return nil
	}
	return a.OperatorStatus_
}

// SetOperatorStatus implements Application.
func (a *application) SetOperatorStatus(args StatusArgs) {
	a.OperatorStatus_ = newStatus(args)
}

// Status implements Application.
func (a *application) Status() Status {
	// To avoid typed nils check nil here.
	if a.Status_ == nil {
		return nil
	}
	return a.Status_
}

// SetStatus implements Application.
func (a *application) SetStatus(args StatusArgs) {
	a.Status_ = newStatus(args)
}

// Units implements Application.
func (a *application) Units() []Unit {
	result := make([]Unit, len(a.Units_.Units_))
	for i, u := range a.Units_.Units_ {
		result[i] = u
	}
	return result
}

func (a *application) unitNames() set.Strings {
	result := set.NewStrings()
	for _, u := range a.Units_.Units_ {
		result.Add(u.Name())
	}
	return result
}

// AddUnit implements Application.
func (a *application) AddUnit(args UnitArgs) Unit {
	u := newUnit(args)
	a.Units_.Units_ = append(a.Units_.Units_, u)
	return u
}

func (a *application) setUnits(unitList []*unit) {
	a.Units_ = units{
		Version: 3,
		Units_:  unitList,
	}
}

// Constraints implements HasConstraints.
func (a *application) Constraints() Constraints {
	if a.Constraints_ == nil {
		return nil
	}
	return a.Constraints_
}

// SetConstraints implements HasConstraints.
func (a *application) SetConstraints(args ConstraintsArgs) {
	a.Constraints_ = newConstraints(args)
}

// CloudService implements Application.
func (a *application) CloudService() CloudService {
	if a.CloudService_ == nil {
		return nil
	}
	return a.CloudService_
}

// SetCloudService implements Application.
func (a *application) SetCloudService(args CloudServiceArgs) {
	a.CloudService_ = newCloudService(&args)
}

// Resources implements Application.
func (a *application) Resources() []Resource {
	rs := a.Resources_.Resources_
	result := make([]Resource, len(rs))
	for i, r := range rs {
		result[i] = r
	}
	return result
}

// AddResource implements Application.
func (a *application) AddResource(args ResourceArgs) Resource {
	r := newResource(args)
	a.Resources_.Resources_ = append(a.Resources_.Resources_, r)
	return r
}

func (a *application) setResources(resourceList []*resource) {
	a.Resources_ = resources{
		Version:    1,
		Resources_: resourceList,
	}
}

// Tools implements Application.
func (a *application) Tools() AgentTools {
	// To avoid a typed nil, check before returning.
	if a.Tools_ == nil {
		return nil
	}
	return a.Tools_
}

// SetTools implements Application.
func (a *application) SetTools(args AgentToolsArgs) {
	a.Tools_ = newAgentTools(args)
}

// CharmOrigin implements Application.
func (a *application) CharmOrigin() CharmOrigin {
	// To avoid a typed nil, check before returning.
	if a.CharmOrigin_ == nil {
		return nil
	}
	return a.CharmOrigin_
}

// SetCharmOrigin implements Application.
func (a *application) SetCharmOrigin(args CharmOriginArgs) {
	a.CharmOrigin_ = newCharmOrigin(args)
}

// CharmMetadata implements Application.
func (a *application) CharmMetadata() CharmMetadata {
	// To avoid a typed nil, check before returning.
	if a.CharmMetadata_ == nil {
		return nil
	}
	return a.CharmMetadata_
}

// SetCharmMetadata implements Application.
func (a *application) SetCharmMetadata(args CharmMetadataArgs) {
	a.CharmMetadata_ = newCharmMetadata(args)
}

// CharmManifest implements Application.
func (a *application) CharmManifest() CharmManifest {
	// To avoid a typed nil, check before returning.
	if a.CharmManifest_ == nil {
		return nil
	}
	return a.CharmManifest_
}

// SetCharmManifest implements Application.
func (a *application) SetCharmManifest(args CharmManifestArgs) {
	a.CharmManifest_ = newCharmManifest(args)
}

// CharmActions implements Application.
func (a *application) CharmActions() CharmActions {
	// To avoid a typed nil, check before returning.
	if a.CharmActions_ == nil {
		return nil
	}
	return a.CharmActions_
}

// SetCharmActions implements Application.
func (a *application) SetCharmActions(args CharmActionsArgs) {
	a.CharmActions_ = newCharmActions(args)
}

// CharmConfigs implements Application.
func (a *application) CharmConfigs() CharmConfigs {
	// To avoid a typed nil, check before returning.
	if a.CharmConfigs_ == nil {
		return nil
	}
	return a.CharmConfigs_
}

// SetCharmConfigs implements Application.
func (a *application) SetCharmConfigs(args CharmConfigsArgs) {
	a.CharmConfigs_ = newCharmConfigs(args)
}

// Offers implements Application.
func (a *application) Offers() []ApplicationOffer {
	if a.Offers_ == nil || len(a.Offers_.Offers) == 0 {
		return nil
	}

	res := make([]ApplicationOffer, len(a.Offers_.Offers))
	for i, offer := range a.Offers_.Offers {
		res[i] = offer
	}
	return res
}

// AddOffer implements Application.
func (a *application) AddOffer(args ApplicationOfferArgs) ApplicationOffer {
	if a.Offers_ == nil {
		a.Offers_ = &applicationOffers{
			Version: 2,
		}
	}

	offer := newApplicationOffer(args)
	a.Offers_.Offers = append(a.Offers_.Offers, offer)
	return offer
}

func (a *application) setOffers(offers []*applicationOffer) {
	a.Offers_ = &applicationOffers{
		Version: 2,
		Offers:  offers,
	}
}

// Validate implements Application.
func (a *application) Validate() error {
	if a.Name_ == "" {
		return errors.NotValidf("application missing name")
	}
	if a.Status_ == nil {
		return errors.NotValidf("application %q missing status", a.Name_)
	}

	for _, resource := range a.Resources_.Resources_ {
		if err := resource.Validate(); err != nil {
			return errors.Annotatef(err, "resource %s", resource.Name_)
		}
	}

	// If leader is set, it must match one of the units.
	var leaderFound bool
	// All of the applications units should also be valid.
	for _, u := range a.Units() {
		if err := u.Validate(); err != nil {
			return errors.Trace(err)
		}
		// We know that the unit has a name, because it validated correctly.
		if u.Name() == a.Leader_ {
			leaderFound = true
		}
	}
	if a.Leader_ != "" && !leaderFound {
		return errors.NotValidf("missing unit for leader %q", a.Leader_)
	}
	return nil
}

// ProvisioningState implements Application.
func (a *application) ProvisioningState() ProvisioningState {
	if a.ProvisioningState_ == nil {
		return nil
	}
	return a.ProvisioningState_
}

func importApplications(source map[string]interface{}) ([]*application, error) {
	checker := versionedChecker("applications")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "applications version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := applicationDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["applications"].([]interface{})
	return importApplicationList(sourceList, importFunc)
}

func importApplicationList(sourceList []interface{}, importFunc applicationDeserializationFunc) ([]*application, error) {
	result := make([]*application, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for application %d, %T", i, value)
		}
		application, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "application %d", i)
		}
		result = append(result, application)
	}
	return result, nil
}

type applicationDeserializationFunc func(map[string]interface{}) (*application, error)

var applicationDeserializationFuncs = map[int]applicationDeserializationFunc{
	1:  importApplicationV1,
	2:  importApplicationV2,
	3:  importApplicationV3,
	4:  importApplicationV4,
	5:  importApplicationV5,
	6:  importApplicationV6,
	7:  importApplicationV7,
	8:  importApplicationV8,
	9:  importApplicationV9,
	10: importApplicationV10,
	11: importApplicationV11,
	12: importApplicationV12,
	13: importApplicationV13,
}

func applicationV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"name":                schema.String(),
		"series":              schema.String(),
		"subordinate":         schema.Bool(),
		"charm-url":           schema.String(),
		"cs-channel":          schema.String(),
		"charm-mod-version":   schema.Int(),
		"force-charm":         schema.Bool(),
		"exposed":             schema.Bool(),
		"min-units":           schema.Int(),
		"status":              schema.StringMap(schema.Any()),
		"endpoint-bindings":   schema.StringMap(schema.String()),
		"settings":            schema.StringMap(schema.Any()),
		"leader":              schema.String(),
		"leadership-settings": schema.StringMap(schema.Any()),
		"storage-constraints": schema.StringMap(schema.StringMap(schema.Any())),
		"metrics-creds":       schema.String(),
		"resources":           schema.StringMap(schema.Any()),
		"units":               schema.StringMap(schema.Any()),
	}

	defaults := schema.Defaults{
		"subordinate":         false,
		"force-charm":         false,
		"exposed":             false,
		"min-units":           int64(0),
		"leader":              "",
		"metrics-creds":       "",
		"storage-constraints": schema.Omit,
		"endpoint-bindings":   schema.Omit,
		"application-config":  schema.Omit,
	}
	addAnnotationSchema(fields, defaults)
	addConstraintsSchema(fields, defaults)
	addStatusHistorySchema(fields)
	return fields, defaults
}

func applicationV2Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := applicationV1Fields()
	fields["type"] = schema.String()
	return fields, defaults
}

func applicationV3Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := applicationV2Fields()
	fields["application-config"] = schema.StringMap(schema.Any())
	fields["password-hash"] = schema.String()
	fields["pod-spec"] = schema.String()
	fields["cloud-service"] = schema.StringMap(schema.Any())
	fields["tools"] = schema.StringMap(schema.Any())
	defaults["password-hash"] = ""
	defaults["pod-spec"] = ""
	defaults["cloud-service"] = schema.Omit
	defaults["tools"] = schema.Omit
	return fields, defaults
}

func applicationV4Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := applicationV3Fields()
	fields["placement"] = schema.String()
	fields["desired-scale"] = schema.Int()
	fields["operator-status"] = schema.StringMap(schema.Any())
	defaults["placement"] = ""
	defaults["desired-scale"] = int64(0)
	defaults["operator-status"] = schema.Omit
	return fields, defaults
}

func applicationV5Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := applicationV4Fields()
	fields["offers"] = schema.StringMap(schema.Any())
	defaults["offers"] = schema.Omit
	return fields, defaults
}

func applicationV6Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := applicationV5Fields()
	fields["has-resources"] = schema.Bool()
	defaults["has-resources"] = false
	return fields, defaults
}

func applicationV7Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := applicationV6Fields()
	fields["charm-origin"] = schema.StringMap(schema.Any())
	defaults["charm-origin"] = schema.Omit
	return fields, defaults
}

func applicationV8Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := applicationV7Fields()
	fields["exposed-endpoints"] = schema.StringMap(schema.StringMap(schema.Any()))
	defaults["exposed-endpoints"] = schema.Omit
	return fields, defaults
}

func applicationV9Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := applicationV8Fields()
	delete(fields, "series")
	defaults["series"] = schema.Omit
	return fields, defaults
}

func applicationV10Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := applicationV9Fields()
	fields["opened-port-ranges"] = schema.StringMap(schema.Any())
	defaults["opened-port-ranges"] = schema.Omit
	return fields, defaults
}

func applicationV11Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := applicationV10Fields()
	fields["provisioning-state"] = schema.StringMap(schema.Any())
	defaults["provisioning-state"] = schema.Omit
	return fields, defaults
}

func applicationV12Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := applicationV11Fields()
	delete(fields, "storage-constraints")
	defaults["storage-constraints"] = schema.Omit
	fields["storage-directives"] = schema.StringMap(schema.StringMap(schema.Any()))
	defaults["storage-directives"] = schema.Omit
	return fields, defaults
}

func applicationV13Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := applicationV12Fields()
	fields["charm-metadata"] = schema.StringMap(schema.Any())
	fields["charm-manifest"] = schema.StringMap(schema.Any())
	fields["charm-actions"] = schema.StringMap(schema.Any())
	fields["charm-configs"] = schema.StringMap(schema.Any())
	defaults["charm-metadata"] = schema.Omit
	defaults["charm-manifest"] = schema.Omit
	defaults["charm-actions"] = schema.Omit
	defaults["charm-configs"] = schema.Omit
	return fields, defaults
}

func importApplicationV1(source map[string]interface{}) (*application, error) {
	fields, defaults := applicationV1Fields()
	return importApplication(fields, defaults, 1, source)
}

func importApplicationV2(source map[string]interface{}) (*application, error) {
	fields, defaults := applicationV2Fields()
	return importApplication(fields, defaults, 2, source)
}

func importApplicationV3(source map[string]interface{}) (*application, error) {
	fields, defaults := applicationV3Fields()
	return importApplication(fields, defaults, 3, source)
}

func importApplicationV4(source map[string]interface{}) (*application, error) {
	fields, defaults := applicationV4Fields()
	return importApplication(fields, defaults, 4, source)
}

func importApplicationV5(source map[string]interface{}) (*application, error) {
	fields, defaults := applicationV5Fields()
	return importApplication(fields, defaults, 5, source)
}

func importApplicationV6(source map[string]interface{}) (*application, error) {
	fields, defaults := applicationV6Fields()
	return importApplication(fields, defaults, 6, source)
}

func importApplicationV7(source map[string]interface{}) (*application, error) {
	fields, defaults := applicationV7Fields()
	return importApplication(fields, defaults, 7, source)
}

func importApplicationV8(source map[string]interface{}) (*application, error) {
	fields, defaults := applicationV8Fields()
	return importApplication(fields, defaults, 8, source)
}

func importApplicationV9(source map[string]interface{}) (*application, error) {
	fields, defaults := applicationV9Fields()
	return importApplication(fields, defaults, 9, source)
}

func importApplicationV10(source map[string]interface{}) (*application, error) {
	fields, defaults := applicationV10Fields()
	return importApplication(fields, defaults, 10, source)
}

func importApplicationV11(source map[string]interface{}) (*application, error) {
	fields, defaults := applicationV11Fields()
	return importApplication(fields, defaults, 11, source)
}

func importApplicationV12(source map[string]interface{}) (*application, error) {
	fields, defaults := applicationV12Fields()
	return importApplication(fields, defaults, 12, source)
}

func importApplicationV13(source map[string]interface{}) (*application, error) {
	fields, defaults := applicationV13Fields()
	return importApplication(fields, defaults, 13, source)
}

func importApplication(fields schema.Fields, defaults schema.Defaults, importVersion int, source map[string]interface{}) (*application, error) {
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "application schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	result := &application{
		Name_:                 valid["name"].(string),
		Type_:                 IAAS,
		Subordinate_:          valid["subordinate"].(bool),
		CharmURL_:             valid["charm-url"].(string),
		Channel_:              valid["cs-channel"].(string),
		CharmModifiedVersion_: int(valid["charm-mod-version"].(int64)),
		ForceCharm_:           valid["force-charm"].(bool),
		Exposed_:              valid["exposed"].(bool),
		MinUnits_:             int(valid["min-units"].(int64)),
		EndpointBindings_:     convertToStringMap(valid["endpoint-bindings"]),
		CharmConfig_:          valid["settings"].(map[string]interface{}),
		Leader_:               valid["leader"].(string),
		LeadershipSettings_:   valid["leadership-settings"].(map[string]interface{}),
		StatusHistory_:        newStatusHistory(),
	}

	if importVersion >= 2 {
		result.Type_ = valid["type"].(string)
	}
	if importVersion >= 3 {
		result.PasswordHash_ = valid["password-hash"].(string)
		result.PodSpec_ = valid["pod-spec"].(string)
	}
	if importVersion >= 4 {
		result.Placement_ = valid["placement"].(string)
		result.DesiredScale_ = int(valid["desired-scale"].(int64))

		if operatorStatus, ok := valid["operator-status"].(map[string]interface{}); ok {
			status, err := importStatus(operatorStatus)
			if err != nil {
				return nil, errors.Trace(err)
			}
			result.OperatorStatus_ = status
		}
	}
	if importVersion >= 5 {
		if offerMap, ok := valid["offers"]; ok {
			offers, err := importApplicationOffers(offerMap.(map[string]interface{}))
			if err != nil {
				return nil, errors.Trace(err)
			}
			result.setOffers(offers)
		}
	}
	if importVersion >= 6 {
		result.HasResources_ = valid["has-resources"].(bool)
	}

	if importVersion >= 7 {
		if charmOriginMap, ok := valid["charm-origin"]; ok {
			charmOrigin, err := importCharmOrigin(charmOriginMap.(map[string]interface{}))
			if err != nil {
				return nil, errors.Trace(err)
			}
			result.CharmOrigin_ = charmOrigin
		}
	}

	if importVersion >= 8 {
		if exposedEndpoints, ok := valid["exposed-endpoints"].(map[string]interface{}); ok {
			if result.ExposedEndpoints_, err = importExposedEndpointsMap(exposedEndpoints); err != nil {
				return nil, errors.Trace(err)
			}
		}
	}

	if importVersion >= 11 {
		if provisioningState, ok := valid["provisioning-state"].(map[string]interface{}); ok {
			if result.ProvisioningState_, err = importProvisioningState(provisioningState); err != nil {
				return nil, errors.Trace(err)
			}
		}
	}

	series, hasSeries := valid["series"].(string)
	if importVersion <= 9 && importVersion >= 7 && hasSeries {
		// If we have a series but no platform defined lets make a platform from the series
		if result.CharmOrigin_ != nil && result.CharmOrigin_.Platform_ == "" {
			platform, err := platformFromSeries(series)
			if err != nil {
				return nil, errors.Trace(err)
			}
			result.CharmOrigin_.Platform_ = platform
		}
	}

	if importVersion >= 10 {
		applicationPortRangesSource, ok := valid["opened-port-ranges"].(map[string]interface{})
		if ok {
			machPortRanges, err := importMachinePortRanges(applicationPortRangesSource)
			if err != nil {
				return nil, errors.Trace(err)
			}
			result.OpenedPortRanges_ = machPortRanges
		}

	}

	if importVersion >= 13 {
		// These fields are used to populate the charm data for the application.
		// This ensures that correct RI is maintained for the charm data
		// when migrating between models.

		if charmMetadataMap, ok := valid["charm-metadata"]; ok {
			charmMetadata, err := importCharmMetadata(charmMetadataMap.(map[string]interface{}))
			if err != nil {
				return nil, errors.Trace(err)
			}
			result.CharmMetadata_ = charmMetadata
		}

		if charmManifestMap, ok := valid["charm-manifest"]; ok {
			charmManifest, err := importCharmManifest(charmManifestMap.(map[string]interface{}))
			if err != nil {
				return nil, errors.Trace(err)
			}
			result.CharmManifest_ = charmManifest
		}

		if charmActionsMap, ok := valid["charm-actions"]; ok {
			charmActions, err := importCharmActions(charmActionsMap.(map[string]interface{}))
			if err != nil {
				return nil, errors.Trace(err)
			}
			result.CharmActions_ = charmActions
		}

		if charmConfigMap, ok := valid["charm-configs"]; ok {
			charmConfig, err := importCharmConfigs(charmConfigMap.(map[string]interface{}))
			if err != nil {
				return nil, errors.Trace(err)
			}
			result.CharmConfigs_ = charmConfig
		}
	}

	result.importAnnotations(valid)

	if err := result.importStatusHistory(valid); err != nil {
		return nil, errors.Trace(err)
	}

	if configValues, ok := valid["application-config"]; ok {
		configMap, ok := configValues.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for application-config, %T", configValues)
		}
		result.ApplicationConfig_ = configMap
	}

	if constraintsMap, ok := valid["constraints"]; ok {
		constraints, err := importConstraints(constraintsMap.(map[string]interface{}))
		if err != nil {
			return nil, errors.Trace(err)
		}
		result.Constraints_ = constraints
	}

	storageKey := "storage-directives"
	if importVersion < 12 {
		storageKey = "storage-constraints"
	}
	if constraintsMap, ok := valid[storageKey]; ok {
		directives, err := importStorageDirectives(constraintsMap.(map[string]interface{}))
		if err != nil {
			return nil, errors.Trace(err)
		}
		result.StorageDirectives_ = directives
	}

	if cloudServiceMap, ok := valid["cloud-service"]; ok {
		cloudService, err := importCloudService(cloudServiceMap.(map[string]interface{}))
		if err != nil {
			return nil, errors.Trace(err)
		}
		result.CloudService_ = cloudService
	}

	toolsMap, ok := valid["tools"].(map[string]interface{})
	if ok {
		tools, err := importAgentTools(toolsMap)
		if err != nil {
			return nil, errors.Trace(err)
		}
		result.Tools_ = tools
	}

	encodedCreds := valid["metrics-creds"].(string)
	// The model stores the creds encoded, but we want to make sure that
	// we are storing something that can be decoded.
	if _, err := base64.StdEncoding.DecodeString(encodedCreds); err != nil {
		return nil, errors.Annotate(err, "metrics credentials not valid")
	}
	result.MetricsCredentials_ = encodedCreds

	status, err := importStatus(valid["status"].(map[string]interface{}))
	if err != nil {
		return nil, errors.Trace(err)
	}
	result.Status_ = status

	resources, err := importResources(valid["resources"].(map[string]interface{}))
	if err != nil {
		return nil, errors.Trace(err)
	}
	result.setResources(resources)

	units, err := importUnits(valid["units"].(map[string]interface{}))
	if err != nil {
		return nil, errors.Trace(err)
	}
	// Units inherit model type from their application.
	for _, u := range units {
		u.Type_ = result.Type_

		// Validate to ensure expected type specific
		// attributes like tools are set.
		if err := u.Validate(); err != nil {
			return nil, errors.Trace(err)
		}
	}
	result.setUnits(units)

	return result, nil
}
