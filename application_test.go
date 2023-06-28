// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"time"

	"github.com/juju/names/v4"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type ApplicationSerializationSuite struct {
	SliceSerializationSuite
	StatusHistoryMixinSuite
}

var _ = gc.Suite(&ApplicationSerializationSuite{})

func (s *ApplicationSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "applications"
	s.sliceName = "applications"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importApplications(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["applications"] = []interface{}{}
	}
	s.StatusHistoryMixinSuite.creator = func() HasStatusHistory {
		return minimalApplication()
	}
	s.StatusHistoryMixinSuite.serializer = func(c *gc.C, initial interface{}) HasStatusHistory {
		return s.exportImportLatest(c, initial.(*application))
	}
}

func minimalApplicationMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"name":              "ubuntu",
		"type":              IAAS,
		"charm-url":         "cs:trusty/ubuntu",
		"cs-channel":        "stable",
		"charm-mod-version": 1,
		"status":            minimalStatusMap(),
		"status-history":    emptyStatusHistoryMap(),
		"settings": map[interface{}]interface{}{
			"key": "value",
		},
		"leader": "ubuntu/0",
		"leadership-settings": map[interface{}]interface{}{
			"leader": true,
		},
		"metrics-creds": "c2Vrcml0", // base64 encoded
		"resources": map[interface{}]interface{}{
			"version": 1,
			"resources": []interface{}{
				minimalResourceMap(),
			},
		},
		"units": map[interface{}]interface{}{
			"version": 3,
			"units": []interface{}{
				minimalUnitMap(),
			},
		},
		"charm-origin": minimalCharmOriginMap(),
	}
}

func minimalApplicationWithOfferMap() map[interface{}]interface{} {
	result := minimalApplicationMap()
	result["offers"] = map[interface{}]interface{}{
		"version": 2,
		"offers": []interface{}{
			minimalApplicationOfferV2Map(),
		},
	}
	return result
}

func minimalApplicationMapCAAS() map[interface{}]interface{} {
	result := minimalApplicationMap()
	result["type"] = CAAS
	result["password-hash"] = "some-hash"
	result["pod-spec"] = "some-spec"
	result["placement"] = "foo=bar"
	result["has-resources"] = true
	result["desired-scale"] = 2
	result["cloud-service"] = map[interface{}]interface{}{
		"version":     1,
		"provider-id": "some-provider",
		"addresses": []interface{}{
			map[interface{}]interface{}{"version": 2, "value": "10.0.0.1", "type": "special"},
			map[interface{}]interface{}{"version": 2, "value": "10.0.0.2", "type": "other"},
		},
	}
	result["units"] = map[interface{}]interface{}{
		"version": 3,
		"units": []interface{}{
			minimalUnitMapCAAS(),
		},
	}
	result["tools"] = minimalAgentToolsMap()
	result["operator-status"] = minimalStatusMap()
	result["charm-origin"] = minimalCharmOriginMap()
	return result
}

func minimalApplication(args ...ApplicationArgs) *application {
	if len(args) == 0 {
		args = []ApplicationArgs{minimalApplicationArgs(IAAS)}
	}
	a := newApplication(args[0])
	a.SetStatus(minimalStatusArgs())
	u := a.AddUnit(minimalUnitArgs(a.Type_))
	u.SetAgentStatus(minimalStatusArgs())
	u.SetWorkloadStatus(minimalStatusArgs())
	a.setResources([]*resource{minimalResource()})
	if a.Type_ == CAAS {
		a.SetTools(minimalAgentToolsArgs())
		a.SetOperatorStatus(minimalStatusArgs())
	} else {
		u.SetTools(minimalAgentToolsArgs())
	}
	a.SetCharmOrigin(minimalCharmOriginArgs())
	return a
}

func minimalApplicationWithOffer(args ...ApplicationArgs) *application {
	a := minimalApplication(args...)
	if a.Type_ != CAAS {
		a.setOffers([]*applicationOffer{
			{
				OfferUUID_: "offer-uuid",
				OfferName_: "my-offer",
				Endpoints_: map[string]string{
					"endpoint-1": "endpoint-1",
					"endpoint-2": "endpoint-2",
				},
				ACL_: map[string]string{
					"admin": "admin",
					"foo":   "read",
					"bar":   "consume",
				},
				ApplicationName_:        "foo",
				ApplicationDescription_: "foo description",
			},
		})
	}
	return a
}

func addMinimalApplication(model Model) {
	a := model.AddApplication(minimalApplicationArgs(IAAS))
	a.SetStatus(minimalStatusArgs())
	u := a.AddUnit(minimalUnitArgs(a.Type()))
	u.SetAgentStatus(minimalStatusArgs())
	u.SetWorkloadStatus(minimalStatusArgs())
	u.SetTools(minimalAgentToolsArgs())
}

func minimalApplicationArgs(modelType string) ApplicationArgs {
	result := ApplicationArgs{
		Tag:                  names.NewApplicationTag("ubuntu"),
		Type:                 modelType,
		CharmURL:             "cs:trusty/ubuntu",
		Channel:              "stable",
		CharmModifiedVersion: 1,
		CharmConfig: map[string]interface{}{
			"key": "value",
		},
		Leader: "ubuntu/0",
		LeadershipSettings: map[string]interface{}{
			"leader": true,
		},
		MetricsCredentials: []byte("sekrit"),
	}
	if modelType == CAAS {
		result.PasswordHash = "some-hash"
		result.PodSpec = "some-spec"
		result.Placement = "foo=bar"
		result.HasResources = true
		result.DesiredScale = 2
		result.CloudService = &CloudServiceArgs{
			ProviderId: "some-provider",
			Addresses: []AddressArgs{
				{Value: "10.0.0.1", Type: "special"},
				{Value: "10.0.0.2", Type: "other"},
			},
		}
	}
	return result
}

func (s *ApplicationSerializationSuite) TestNewApplication(c *gc.C) {
	args := ApplicationArgs{
		Tag:                  names.NewApplicationTag("magic"),
		Subordinate:          true,
		CharmURL:             "cs:jammy/magic",
		Channel:              "stable",
		CharmModifiedVersion: 1,
		ForceCharm:           true,
		Exposed:              true,
		ExposedEndpoints: map[string]ExposedEndpointArgs{
			"endpoint0": {
				ExposeToSpaceIDs: []string{"0", "42"},
			},
			"endpoint1": {
				ExposeToCIDRs: []string{"192.168.42.0/24"},
			},
		},
		MinUnits: 42, // no judgement is made by the migration code
		EndpointBindings: map[string]string{
			"rel-name": "some-space",
		},
		ApplicationConfig: map[string]interface{}{
			"config key": "config value",
		},
		CharmConfig: map[string]interface{}{
			"key": "value",
		},
		Leader: "magic/1",
		LeadershipSettings: map[string]interface{}{
			"leader": true,
		},
		MetricsCredentials: []byte("sekrit"),
		PasswordHash:       "passwordhash",
		PodSpec:            "podspec",
		Placement:          "foo=bar",
		HasResources:       true,
		DesiredScale:       2,
	}
	application := newApplication(args)

	c.Assert(application.Name(), gc.Equals, "magic")
	c.Assert(application.Tag(), gc.Equals, names.NewApplicationTag("magic"))
	c.Assert(application.Subordinate(), jc.IsTrue)
	c.Assert(application.CharmURL(), gc.Equals, "cs:jammy/magic")
	c.Assert(application.Channel(), gc.Equals, "stable")
	c.Assert(application.CharmModifiedVersion(), gc.Equals, 1)
	c.Assert(application.ForceCharm(), jc.IsTrue)
	c.Assert(application.Exposed(), jc.IsTrue)

	expEps := application.ExposedEndpoints()
	c.Assert(expEps, gc.HasLen, 2)
	ep0 := expEps["endpoint0"]
	c.Assert(ep0, gc.Not(gc.IsNil))
	c.Assert(ep0.ExposeToSpaceIDs(), gc.DeepEquals, []string{"0", "42"})
	c.Assert(ep0.ExposeToCIDRs(), gc.IsNil)
	ep1 := expEps["endpoint1"]
	c.Assert(ep1, gc.Not(gc.IsNil))
	c.Assert(ep1.ExposeToSpaceIDs(), gc.IsNil)
	c.Assert(ep1.ExposeToCIDRs(), gc.DeepEquals, []string{"192.168.42.0/24"})

	c.Assert(application.PasswordHash(), gc.Equals, "passwordhash")
	c.Assert(application.PodSpec(), gc.Equals, "podspec")
	c.Assert(application.Placement(), gc.Equals, "foo=bar")
	c.Assert(application.HasResources(), jc.IsTrue)
	c.Assert(application.DesiredScale(), gc.Equals, 2)
	c.Assert(application.CloudService(), gc.IsNil)
	c.Assert(application.StorageConstraints(), gc.HasLen, 0)
	c.Assert(application.MinUnits(), gc.Equals, 42)
	c.Assert(application.EndpointBindings(), jc.DeepEquals, args.EndpointBindings)
	c.Assert(application.ApplicationConfig(), jc.DeepEquals, args.ApplicationConfig)
	c.Assert(application.CharmConfig(), jc.DeepEquals, args.CharmConfig)
	c.Assert(application.Leader(), gc.Equals, "magic/1")
	c.Assert(application.LeadershipSettings(), jc.DeepEquals, args.LeadershipSettings)
	c.Assert(application.MetricsCredentials(), jc.DeepEquals, []byte("sekrit"))
}

func (s *ApplicationSerializationSuite) TestMinimalApplicationValid(c *gc.C) {
	application := minimalApplication()
	c.Assert(application.Validate(), jc.ErrorIsNil)
}

func (s *ApplicationSerializationSuite) TestMinimalCAASApplicationValid(c *gc.C) {
	application := minimalApplication(minimalApplicationArgs(CAAS))
	c.Assert(application.Validate(), jc.ErrorIsNil)
}

func (s *ApplicationSerializationSuite) TestMinimalMatchesCAAS(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	bytes, err := yaml.Marshal(minimalApplication(args))
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalApplicationMapCAAS())
}

func (s *ApplicationSerializationSuite) TestMinimalMatchesIAAS(c *gc.C) {
	bytes, err := yaml.Marshal(minimalApplication())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalApplicationMap())
}

func (s *ApplicationSerializationSuite) TestMinimalWithOfferMatchesIAAS(c *gc.C) {
	bytes, err := yaml.Marshal(minimalApplicationWithOffer())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalApplicationWithOfferMap())
}

func (s *ApplicationSerializationSuite) TestParsingSerializedDataWithOfferBlock(c *gc.C) {
	app := minimalApplicationWithOffer()
	application := s.exportImportLatest(c, app)
	c.Assert(application, jc.DeepEquals, app)
}

func (s *ApplicationSerializationSuite) TestAddOpenedPortRange(c *gc.C) {
	app := minimalApplication()
	app.AddOpenedPortRange(
		OpenedPortRangeArgs{
			UnitName:     "magic/0",
			EndpointName: "",
			FromPort:     666,
			ToPort:       666,
			Protocol:     "tcp",
		},
	)
	app.AddOpenedPortRange(
		OpenedPortRangeArgs{
			UnitName:     "magic/1",
			EndpointName: "",
			FromPort:     888,
			ToPort:       888,
			Protocol:     "tcp",
		},
	)
	c.Assert(app.OpenedPortRanges().ByUnit(), gc.HasLen, 2)

	application := s.exportImportLatest(c, app)
	c.Assert(application.OpenedPortRanges().ByUnit(), gc.HasLen, 2)
}

func (s *ApplicationSerializationSuite) exportImportVersion(c *gc.C, application_ *application, version int) *application {
	initial := applications{
		Version:       version,
		Applications_: []*application{application_},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	applications, err := importApplications(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(applications, gc.HasLen, 1)
	return applications[0]
}

func (s *ApplicationSerializationSuite) exportImportLatest(c *gc.C, application_ *application) *application {
	return s.exportImportVersion(c, application_, 12)
}

func (s *ApplicationSerializationSuite) TestV1ParsingReturnsLatest(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	args.Type = ""
	args.Series = "focal"
	appV1 := minimalApplication(args)

	// Make an app with fields not in v1 removed.
	appLatest := minimalApplication()
	appLatest.PasswordHash_ = ""
	appLatest.PodSpec_ = ""
	appLatest.Placement_ = ""
	appLatest.HasResources_ = false
	appLatest.DesiredScale_ = 0
	appLatest.CloudService_ = nil
	appLatest.Tools_ = nil
	appLatest.OperatorStatus_ = nil
	appLatest.Offers_ = nil
	appLatest.CharmOrigin_ = nil

	appResult := s.exportImportVersion(c, appV1, 1)
	appLatest.Series_ = ""
	c.Assert(appResult, jc.DeepEquals, appLatest)
}

func (s *ApplicationSerializationSuite) TestV2ParsingReturnsLatest(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	args.Series = "focal"
	appV1 := minimalApplication(args)

	// Make an app with fields not in v2 removed.
	appLatest := appV1
	appLatest.PasswordHash_ = ""
	appLatest.PodSpec_ = ""
	appLatest.Placement_ = ""
	appLatest.HasResources_ = false
	appLatest.DesiredScale_ = 0
	appLatest.CloudService_ = nil
	appLatest.Tools_ = nil
	appLatest.OperatorStatus_ = nil
	appLatest.Offers_ = nil
	appLatest.CharmOrigin_ = nil

	appResult := s.exportImportVersion(c, appV1, 2)
	appLatest.Series_ = ""
	c.Assert(appResult, jc.DeepEquals, appLatest)
}

func (s *ApplicationSerializationSuite) TestV3ParsingReturnsLatest(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	args.Series = "focal"
	appV2 := minimalApplication(args)

	// Make an app with fields not in v3 removed.
	appLatest := appV2
	appLatest.Placement_ = ""
	appLatest.HasResources_ = false
	appLatest.DesiredScale_ = 0
	appLatest.OperatorStatus_ = nil
	appLatest.Offers_ = nil
	appLatest.CharmOrigin_ = nil

	appResult := s.exportImportVersion(c, appV2, 3)
	appLatest.Series_ = ""
	c.Assert(appResult, jc.DeepEquals, appLatest)
}

func (s *ApplicationSerializationSuite) TestV5ParsingReturnsLatest(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	args.Series = "focal"
	appV5 := minimalApplication(args)

	// Make an app with fields not in v5 removed.
	appLatest := appV5
	appLatest.HasResources_ = false
	appLatest.CharmOrigin_ = nil

	appResult := s.exportImportVersion(c, appV5, 5)
	appLatest.Series_ = ""
	c.Assert(appResult, jc.DeepEquals, appLatest)
}

func (s *ApplicationSerializationSuite) TestV6ParsingReturnsLatest(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	args.Series = "focal"
	appV6 := minimalApplication(args)

	// Make an app with fields not in v6 removed.
	appLatest := appV6
	appLatest.CharmOrigin_ = nil

	appResult := s.exportImportVersion(c, appV6, 6)
	appLatest.Series_ = ""
	c.Assert(appResult, jc.DeepEquals, appLatest)
}

func (s *ApplicationSerializationSuite) TestParsingSerializedData(c *gc.C) {
	app := minimalApplication()
	application := s.exportImportLatest(c, app)
	c.Assert(application, jc.DeepEquals, app)
}

func (s *ApplicationSerializationSuite) TestEndpointBindings(c *gc.C) {
	args := minimalApplicationArgs(IAAS)
	args.EndpointBindings = map[string]string{
		"rel-name": "some-space",
		"other":    "other-space",
	}
	initial := minimalApplication(args)
	application := s.exportImportLatest(c, initial)
	c.Assert(application.EndpointBindings(), jc.DeepEquals, args.EndpointBindings)
}

func (s *ApplicationSerializationSuite) TestAnnotations(c *gc.C) {
	initial := minimalApplication()
	annotations := map[string]string{
		"string":  "value",
		"another": "one",
	}
	initial.SetAnnotations(annotations)

	application := s.exportImportLatest(c, initial)
	c.Assert(application.Annotations(), jc.DeepEquals, annotations)
}

func (s *ApplicationSerializationSuite) TestConstraints(c *gc.C) {
	initial := minimalApplication()
	args := ConstraintsArgs{
		Architecture: "amd64",
		Memory:       8 * gig,
		RootDisk:     40 * gig,
	}
	initial.SetConstraints(args)

	application := s.exportImportLatest(c, initial)
	c.Assert(application.Constraints(), jc.DeepEquals, newConstraints(args))
}

func (s *ApplicationSerializationSuite) TestStorageConstraints(c *gc.C) {
	args := minimalApplicationArgs(IAAS)
	args.StorageConstraints = map[string]StorageConstraintArgs{
		"first":  {Pool: "first", Size: 1234, Count: 1},
		"second": {Pool: "second", Size: 4321, Count: 7},
	}
	initial := minimalApplication(args)

	application := s.exportImportLatest(c, initial)

	constraints := application.StorageConstraints()
	c.Assert(constraints, gc.HasLen, 2)
	first, found := constraints["first"]
	c.Assert(found, jc.IsTrue)
	c.Check(first.Pool(), gc.Equals, "first")
	c.Check(first.Size(), gc.Equals, uint64(1234))
	c.Check(first.Count(), gc.Equals, uint64(1))

	second, found := constraints["second"]
	c.Assert(found, jc.IsTrue)
	c.Check(second.Pool(), gc.Equals, "second")
	c.Check(second.Size(), gc.Equals, uint64(4321))
	c.Check(second.Count(), gc.Equals, uint64(7))
}

func (s *ApplicationSerializationSuite) TestApplicationConfig(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	args.ApplicationConfig = map[string]interface{}{
		"first":  "value 1",
		"second": 42,
	}
	initial := minimalApplication(args)

	application := s.exportImportLatest(c, initial)
	c.Assert(application.ApplicationConfig(), jc.DeepEquals, map[string]interface{}{
		"first":  "value 1",
		"second": 42,
	})
}

func (s *ApplicationSerializationSuite) TestPasswordHash(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	args.PasswordHash = "passwordhash"
	initial := minimalApplication(args)

	application := s.exportImportLatest(c, initial)
	c.Assert(application.PasswordHash(), gc.Equals, "passwordhash")
}

func (s *ApplicationSerializationSuite) TestPodSpec(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	args.PodSpec = "podspec"
	initial := minimalApplication(args)

	application := s.exportImportLatest(c, initial)
	c.Assert(application.PodSpec(), gc.Equals, "podspec")
}

func (s *ApplicationSerializationSuite) TestPlacement(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	args.Placement = "foo=baz"
	initial := minimalApplication(args)

	application := s.exportImportLatest(c, initial)
	c.Assert(application.Placement(), gc.Equals, "foo=baz")
}

func (s *ApplicationSerializationSuite) TestHasResources(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	args.HasResources = true
	initial := minimalApplication(args)

	application := s.exportImportLatest(c, initial)
	c.Assert(application.HasResources(), jc.IsTrue)
}

func (s *ApplicationSerializationSuite) TestDesiredScale(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	args.DesiredScale = 3
	initial := minimalApplication(args)

	application := s.exportImportLatest(c, initial)
	c.Assert(application.DesiredScale(), gc.Equals, 3)
}

func (s *ApplicationSerializationSuite) TestProvisioningState(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	args.ProvisioningState = &ProvisioningStateArgs{
		Scaling:     true,
		ScaleTarget: 10,
	}
	initial := minimalApplication(args)

	application := s.exportImportLatest(c, initial)
	c.Assert(application.ProvisioningState().Scaling(), jc.IsTrue)
	c.Assert(application.ProvisioningState().ScaleTarget(), gc.Equals, 10)
}

func (s *ApplicationSerializationSuite) TestLease(c *gc.C) {
	now := time.Now().UTC().Round(time.Second)
	args := minimalApplicationArgs(CAAS)
	args.Lease = &LeaseArgs{
		Name:   "name",
		Holder: "holder",
		Start:  now,
		Expiry: now.Add(10 * time.Minute),
		Pinned: true,
	}
	initial := minimalApplication(args)

	application := s.exportImportLatest(c, initial)
	c.Assert(application.Lease().Name(), gc.Equals, "name")
	c.Assert(application.Lease().Holder(), gc.Equals, "holder")
	c.Assert(application.Lease().Start(), gc.DeepEquals, now)
	c.Assert(application.Lease().Expiry(), gc.DeepEquals, now.Add(10*time.Minute))
	c.Assert(application.Lease().Pinned(), jc.IsTrue)
}

func (s *ApplicationSerializationSuite) TestCloudService(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	initial := minimalApplication(args)
	serviceArgs := CloudServiceArgs{
		ProviderId: "some-provider",
		Addresses: []AddressArgs{
			{Value: "10.0.0.1", Type: "special"},
			{Value: "10.0.0.2", Type: "other"},
		},
	}
	initial.SetCloudService(serviceArgs)

	app := s.exportImportLatest(c, initial)
	c.Assert(app.CloudService(), jc.DeepEquals, newCloudService(&serviceArgs))
}

func (s *ApplicationSerializationSuite) TestLeaderValid(c *gc.C) {
	args := minimalApplicationArgs(IAAS)
	args.Leader = "ubuntu/1"
	application := newApplication(args)
	application.SetStatus(minimalStatusArgs())

	err := application.Validate()
	c.Assert(err, gc.ErrorMatches, `missing unit for leader "ubuntu/1" not valid`)
}

func (s *ApplicationSerializationSuite) TestResourcesAreValidated(c *gc.C) {
	application := minimalApplication()
	application.AddResource(ResourceArgs{Name: "foo"})
	err := application.Validate()
	c.Assert(err, gc.ErrorMatches, `resource foo: no application revision set`)
}

func (s *ApplicationSerializationSuite) TestIAASUnitMissingTools(c *gc.C) {
	app := minimalApplication()
	app.Units_.Units_[0].Tools_ = nil
	initial := applications{
		Version:       9,
		Applications_: []*application{app},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	_, err = importApplications(source)
	c.Assert(err, gc.ErrorMatches, `application 0: unit "ubuntu/0" missing tools not valid`)
}

func (s *ApplicationSerializationSuite) TestExposeMetadata(c *gc.C) {
	args := minimalApplicationArgs(IAAS)
	args.Exposed = true
	args.ExposedEndpoints = map[string]ExposedEndpointArgs{
		"endpoint0": {
			ExposeToSpaceIDs: []string{"0", "42"},
		},
		"endpoint1": {
			ExposeToCIDRs: []string{"192.168.42.0/24"},
		},
	}

	initial := minimalApplication(args)
	application := s.exportImportLatest(c, initial)

	expEps := application.ExposedEndpoints()
	c.Assert(expEps, gc.HasLen, 2)

	ep0 := expEps["endpoint0"]
	c.Assert(ep0, gc.Not(gc.IsNil))
	c.Assert(ep0.ExposeToSpaceIDs(), gc.DeepEquals, []string{"0", "42"})
	c.Assert(ep0.ExposeToCIDRs(), gc.IsNil)

	ep1 := expEps["endpoint1"]
	c.Assert(ep1, gc.Not(gc.IsNil))
	c.Assert(ep1.ExposeToSpaceIDs(), gc.IsNil)
	c.Assert(ep1.ExposeToCIDRs(), gc.DeepEquals, []string{"192.168.42.0/24"})
}

func (s *ApplicationSerializationSuite) TestApplicationSeriesToPlatform(c *gc.C) {
	appInput := map[string]interface{}{
		"version": 8,
		"applications": []map[string]interface{}{{
			"version":           8,
			"name":              "ubuntu",
			"type":              IAAS,
			"charm-url":         "cs:trusty/ubuntu",
			"cs-channel":        "stable",
			"charm-mod-version": 1,
			"status":            minimalStatusMap(),
			"status-history":    emptyStatusHistoryMap(),
			"settings": map[interface{}]interface{}{
				"key": "value",
			},
			"leader":              "ubuntu/0",
			"leadership-settings": map[interface{}]interface{}{},
			"metrics-creds":       "c2Vrcml0", // base64 encoded
			"resources": map[interface{}]interface{}{
				"version": 1,
				"resources": []interface{}{
					minimalResourceMap(),
				},
			},
			"units": map[interface{}]interface{}{
				"version": 3,
				"units": []interface{}{
					minimalUnitMap(),
				},
			},
			"series": "focal",
			"charm-origin": map[interface{}]interface{}{
				"version":  2,
				"source":   "local",
				"id":       "",
				"hash":     "",
				"revision": 0,
				"channel":  "",
				"platform": "",
			},
		}},
	}

	bytes, err := yaml.Marshal(appInput)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	app, err := importApplications(source)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(app[0].CharmOrigin().Platform(), gc.Equals, "unknown/ubuntu/20.04")
}
