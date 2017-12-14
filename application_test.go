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
		"type":              "iaas",
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
			"version": 1,
			"units": []interface{}{
				minimalUnitMap(),
			},
		},
	}
}

func minimalApplication(args ...ApplicationArgs) *application {
	if len(args) == 0 {
		args = []ApplicationArgs{minimalApplicationArgs()}
	}
	s := newApplication(args[0])
	s.SetStatus(minimalStatusArgs())
	u := s.AddUnit(minimalUnitArgs())
	u.SetAgentStatus(minimalStatusArgs())
	u.SetWorkloadStatus(minimalStatusArgs())
	u.SetTools(minimalAgentToolsArgs())
	s.setResources([]*resource{minimalResource()})
	return s
}

func addMinimalApplication(model Model) {
	s := model.AddApplication(minimalApplicationArgs())
	s.SetStatus(minimalStatusArgs())
	u := s.AddUnit(minimalUnitArgs())
	u.SetAgentStatus(minimalStatusArgs())
	u.SetWorkloadStatus(minimalStatusArgs())
	u.SetTools(minimalAgentToolsArgs())
}

func minimalApplicationArgs() ApplicationArgs {
	return ApplicationArgs{
		Tag:                  names.NewApplicationTag("ubuntu"),
		Series:               "trusty",
		Type:                 "iaas",
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

func (s *ApplicationSerializationSuite) TestMinimalMatches(c *gc.C) {
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
	return s.exportImportVersion(c, application_, 3)
}

func (s *ApplicationSerializationSuite) TestV1ParsingReturnsLatest(c *gc.C) {
	args := minimalApplicationArgs()
	args.Type = ""
	appV1 := minimalApplication(args)
	appLatest := minimalApplication()
	appResult := s.exportImportVersion(c, appV1, 1)
	c.Assert(appResult, jc.DeepEquals, appLatest)
}

func (s *ApplicationSerializationSuite) TestV2ParsingReturnsLatest(c *gc.C) {
	args := minimalApplicationArgs()
	appV1 := minimalApplication(args)
	appLatest := minimalApplication()
	appResult := s.exportImportVersion(c, appV1, 2)
	c.Assert(appResult, jc.DeepEquals, appLatest)
}

func (s *ApplicationSerializationSuite) TestParsingSerializedData(c *gc.C) {
	app := minimalApplication()
	application := s.exportImportLatest(c, app)
	c.Assert(application, jc.DeepEquals, app)
}

func (s *ApplicationSerializationSuite) TestEndpointBindings(c *gc.C) {
	args := minimalApplicationArgs()
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
	args := minimalApplicationArgs()
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
	args := minimalApplicationArgs()
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

func (s *ApplicationSerializationSuite) TestLeaderValid(c *gc.C) {
	args := minimalApplicationArgs()
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
