// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/names.v2"
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
		"series":            "trusty",
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
			"version": 2,
			"units": []interface{}{
				minimalUnitMap(),
			},
		},
	}
}

func minimalApplicationMapCAAS() map[interface{}]interface{} {
	result := minimalApplicationMap()
	result["type"] = CAAS
	result["password-hash"] = "some-hash"
	result["pod-spec"] = "some-spec"
	result["placement"] = "foo=bar"
	result["desired-scale"] = 2
	result["cloud-service"] = map[interface{}]interface{}{
		"version":     1,
		"provider-id": "some-provider",
		"addresses": []interface{}{
			map[interface{}]interface{}{"version": 1, "value": "10.0.0.1", "type": "special"},
			map[interface{}]interface{}{"version": 1, "value": "10.0.0.2", "type": "other"},
		},
	}
	result["units"] = map[interface{}]interface{}{
		"version": 2,
		"units": []interface{}{
			minimalUnitMapCAAS(),
		},
	}
	result["tools"] = minimalAgentToolsMap()
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
	} else {
		u.SetTools(minimalAgentToolsArgs())
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
		Series:               "trusty",
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
		Series:               "zesty",
		Subordinate:          true,
		CharmURL:             "cs:zesty/magic",
		Channel:              "stable",
		CharmModifiedVersion: 1,
		ForceCharm:           true,
		Exposed:              true,
		MinUnits:             42, // no judgement is made by the migration code
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
		DesiredScale:       2,
	}
	application := newApplication(args)

	c.Assert(application.Name(), gc.Equals, "magic")
	c.Assert(application.Tag(), gc.Equals, names.NewApplicationTag("magic"))
	c.Assert(application.Series(), gc.Equals, "zesty")
	c.Assert(application.Subordinate(), jc.IsTrue)
	c.Assert(application.CharmURL(), gc.Equals, "cs:zesty/magic")
	c.Assert(application.Channel(), gc.Equals, "stable")
	c.Assert(application.CharmModifiedVersion(), gc.Equals, 1)
	c.Assert(application.ForceCharm(), jc.IsTrue)
	c.Assert(application.Exposed(), jc.IsTrue)
	c.Assert(application.PasswordHash(), gc.Equals, "passwordhash")
	c.Assert(application.PodSpec(), gc.Equals, "podspec")
	c.Assert(application.Placement(), gc.Equals, "foo=bar")
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
	return s.exportImportVersion(c, application_, 4)
}

func (s *ApplicationSerializationSuite) TestV1ParsingReturnsLatest(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	args.Type = ""
	appV1 := minimalApplication(args)

	// Make an app with fields not in v1 removed.
	appLatest := minimalApplication()
	appLatest.PasswordHash_ = ""
	appLatest.PodSpec_ = ""
	appLatest.Placement_ = ""
	appLatest.DesiredScale_ = 0
	appLatest.CloudService_ = nil
	appLatest.Tools_ = nil

	appResult := s.exportImportVersion(c, appV1, 1)
	c.Assert(appResult, jc.DeepEquals, appLatest)
}

func (s *ApplicationSerializationSuite) TestV2ParsingReturnsLatest(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	appV1 := minimalApplication(args)

	// Make an app with fields not in v2 removed.
	appLatest := appV1
	appLatest.PasswordHash_ = ""
	appLatest.PodSpec_ = ""
	appLatest.Placement_ = ""
	appLatest.DesiredScale_ = 0
	appLatest.CloudService_ = nil
	appLatest.Tools_ = nil

	appResult := s.exportImportVersion(c, appV1, 2)
	c.Assert(appResult, jc.DeepEquals, appLatest)
}

func (s *ApplicationSerializationSuite) TestV3ParsingReturnsLatest(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	appV2 := minimalApplication(args)

	// Make an app with fields not in v3 removed.
	appLatest := appV2
	appLatest.Placement_ = ""
	appLatest.DesiredScale_ = 0

	appResult := s.exportImportVersion(c, appV2, 3)
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

func (s *ApplicationSerializationSuite) TestDesiredScale(c *gc.C) {
	args := minimalApplicationArgs(CAAS)
	args.DesiredScale = 3
	initial := minimalApplication(args)

	application := s.exportImportLatest(c, initial)
	c.Assert(application.DesiredScale(), gc.Equals, 3)
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

func (s *ApplicationSerializationSuite) TestCAASMissingToolsValidated(c *gc.C) {
	app := minimalApplication(minimalApplicationArgs(CAAS))
	app.Tools_ = nil
	err := app.Validate()
	c.Assert(err, gc.ErrorMatches, `application "ubuntu" missing tools not valid`)
}

func (s *ApplicationSerializationSuite) TestCAASApplicationMissingTools(c *gc.C) {
	app := minimalApplication(minimalApplicationArgs(CAAS))
	app.Tools_ = nil
	initial := applications{
		Version:       3,
		Applications_: []*application{app},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	_, err = importApplications(source)
	c.Assert(err, gc.ErrorMatches, "application 0: tools metadata in CAAS model not found")
}

func (s *ApplicationSerializationSuite) TestIAASUnitMissingTools(c *gc.C) {
	app := minimalApplication()
	app.Units_.Units_[0].Tools_ = nil
	initial := applications{
		Version:       3,
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
