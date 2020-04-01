// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"fmt"
	"time"

	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/version"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/names.v3"
	"gopkg.in/yaml.v2"
)

type ModelSerializationSuite struct {
	testing.IsolationSuite
	StatusHistoryMixinSuite
}

var _ = gc.Suite(&ModelSerializationSuite{})

func (s *ModelSerializationSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	s.StatusHistoryMixinSuite.creator = func() HasStatusHistory {
		return s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	}
	s.StatusHistoryMixinSuite.serializer = func(c *gc.C, initial interface{}) HasStatusHistory {
		return s.exportImport(c, initial.(Model))
	}
}

func (*ModelSerializationSuite) TestNil(c *gc.C) {
	_, err := importModel(nil)
	c.Check(err, gc.ErrorMatches, "version: expected int, got nothing")
}

func (*ModelSerializationSuite) TestMissingVersion(c *gc.C) {
	_, err := importModel(map[string]interface{}{})
	c.Check(err, gc.ErrorMatches, "version: expected int, got nothing")
}

func (*ModelSerializationSuite) TestNonIntVersion(c *gc.C) {
	_, err := importModel(map[string]interface{}{
		"version": "hello",
	})
	c.Check(err.Error(), gc.Equals, `version: expected int, got string("hello")`)
}

func (*ModelSerializationSuite) TestUnknownVersion(c *gc.C) {
	_, err := importModel(map[string]interface{}{
		"version": 42,
	})
	c.Check(err.Error(), gc.Equals, `version 42 not valid`)
}

func (*ModelSerializationSuite) TestUpdateConfig(c *gc.C) {
	model := NewModel(ModelArgs{
		Config: map[string]interface{}{
			"name": "awesome",
			"uuid": "some-uuid",
		},
		CloudRegion: "some-region",
	})
	model.UpdateConfig(map[string]interface{}{
		"name": "something else",
		"key":  "value",
	})
	c.Assert(model.Config(), jc.DeepEquals, map[string]interface{}{
		"name": "something else",
		"uuid": "some-uuid",
		"key":  "value",
	})
}

func (*ModelSerializationSuite) TestType(c *gc.C) {
	model := NewModel(ModelArgs{Type: "faas"})
	c.Check(model.Type(), gc.Equals, "faas")
}

func (*ModelSerializationSuite) TestCloudCredentials(c *gc.C) {
	owner := names.NewUserTag("me")
	model := NewModel(ModelArgs{
		Owner: owner,
	})
	args := CloudCredentialArgs{
		Owner:    owner,
		Cloud:    names.NewCloudTag("altostratus"),
		Name:     "creds",
		AuthType: "fuzzy",
		Attributes: map[string]string{
			"key": "value",
		},
	}
	model.SetCloudCredential(args)
	creds := model.CloudCredential()

	c.Check(creds.Owner(), gc.Equals, args.Owner.Id())
	c.Check(creds.Cloud(), gc.Equals, args.Cloud.Id())
	c.Check(creds.Name(), gc.Equals, args.Name)
	c.Check(creds.AuthType(), gc.Equals, args.AuthType)
	c.Check(creds.Attributes(), jc.DeepEquals, args.Attributes)
}

func (s *ModelSerializationSuite) exportImport(c *gc.C, initial Model) Model {
	bytes, err := Serialize(initial)
	c.Assert(err, jc.ErrorIsNil)
	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	return model
}

func (s *ModelSerializationSuite) TestParsingModelV1(c *gc.C) {
	model, err := Deserialize([]byte(modelV1example))
	c.Assert(err, jc.ErrorIsNil)
	c.Check(model.Validate(), jc.ErrorIsNil)
	// Version 1 also incorrectly serialized machine cloud instance
	// status. So when parsing a v1 model, the cloud instance status
	// is set to unknown.
	machines := model.Machines()
	instance := machines[0].Instance()
	c.Check(instance.Status().Value(), gc.Equals, "unknown")
}

func (s *ModelSerializationSuite) TestVersions(c *gc.C) {
	args := ModelArgs{
		Type:  IAAS,
		Owner: names.NewUserTag("magic"),
		Config: map[string]interface{}{
			"name": "awesome",
			"uuid": "some-uuid",
		},
		LatestToolsVersion: version.MustParse("2.0.1"),
		EnvironVersion:     123,
		Blocks: map[string]string{
			"all-changes": "locked down",
		},
		Cloud:       "vapour",
		CloudRegion: "east-west",
	}
	initial := NewModel(args).(*model)
	c.Assert(initial.Applications_.Version, gc.Equals, len(applicationDeserializationFuncs))
	c.Assert(initial.Actions_.Version, gc.Equals, 3)
	c.Assert(initial.Operations_.Version, gc.Equals, 1)
	c.Assert(initial.Filesystems_.Version, gc.Equals, len(filesystemDeserializationFuncs))
	c.Assert(initial.Relations_.Version, gc.Equals, len(relationFieldsFuncs))
	c.Assert(initial.RemoteEntities_.Version, gc.Equals, len(remoteEntityFieldsFuncs))
	c.Assert(initial.RemoteApplications_.Version, gc.Equals, len(remoteApplicationFieldsFuncs))
	c.Assert(initial.Spaces_.Version, gc.Equals, len(spaceDeserializationFuncs))
	c.Assert(initial.Volumes_.Version, gc.Equals, len(volumeDeserializationFuncs))
	c.Assert(initial.FirewallRules_.Version, gc.Equals, len(firewallRuleFieldsFuncs))
	c.Assert(initial.OfferConnections_.Version, gc.Equals, len(offerConnectionDeserializationFuncs))
	c.Assert(initial.ExternalControllers_.Version, gc.Equals, len(externalControllerDeserializationFuncs))
}

func (s *ModelSerializationSuite) TestParsingYAML(c *gc.C) {
	s.testParsingYAMLWithMachine(c, func(initial Model) {
		addMinimalMachine(initial, "0")
	})
}

func (s *ModelSerializationSuite) TestParsingYAMLWithMissingModificationStatus(c *gc.C) {
	s.testParsingYAMLWithMachine(c, func(initial Model) {
		addMinimalMachineWithMissingModificationStatus(initial, "0")
	})
}

func (s *ModelSerializationSuite) testParsingYAMLWithMachine(c *gc.C, machineFn func(Model)) {
	args := ModelArgs{
		Type:  IAAS,
		Owner: names.NewUserTag("magic"),
		Config: map[string]interface{}{
			"name": "awesome",
			"uuid": "some-uuid",
		},
		LatestToolsVersion: version.MustParse("2.0.1"),
		EnvironVersion:     123,
		Blocks: map[string]string{
			"all-changes": "locked down",
		},
		Cloud:       "vapour",
		CloudRegion: "east-west",
	}
	initial := NewModel(args)
	initial.SetCloudCredential(CloudCredentialArgs{
		Name:  "creds",
		Cloud: names.NewCloudTag("vapour"),
		Owner: names.NewUserTag("admin"),
	})
	adminUser := names.NewUserTag("admin")
	initial.AddUser(UserArgs{
		Name:        adminUser,
		CreatedBy:   adminUser,
		DateCreated: time.Date(2015, 10, 9, 12, 34, 56, 0, time.UTC),
	})
	initial.SetStatus(StatusArgs{Value: "available"})
	machineFn(initial)
	addMinimalApplication(initial)
	model := s.exportImport(c, initial)

	c.Assert(model.Type(), gc.Equals, IAAS)
	c.Assert(model.Owner(), gc.Equals, args.Owner)
	c.Assert(model.Tag().Id(), gc.Equals, "some-uuid")
	c.Assert(model.Config(), jc.DeepEquals, args.Config)
	c.Assert(model.Cloud(), gc.Equals, "vapour")
	c.Assert(model.CloudRegion(), gc.Equals, "east-west")
	c.Assert(model.CloudCredential(), jc.DeepEquals, initial.CloudCredential())
	c.Assert(model.LatestToolsVersion(), gc.Equals, args.LatestToolsVersion)
	c.Assert(model.EnvironVersion(), gc.Equals, args.EnvironVersion)
	c.Assert(model.Blocks(), jc.DeepEquals, args.Blocks)
	users := model.Users()
	c.Assert(users, gc.HasLen, 1)
	c.Assert(users[0].Name(), gc.Equals, adminUser)
	machines := model.Machines()
	c.Assert(machines, gc.HasLen, 1)
	c.Assert(machines[0].Id(), gc.Equals, "0")
	applications := model.Applications()
	c.Assert(applications, gc.HasLen, 1)
	c.Assert(applications[0].Name(), gc.Equals, "ubuntu")
}

func (s *ModelSerializationSuite) newModel(args ModelArgs) Model {
	initial := NewModel(args)
	initial.SetStatus(StatusArgs{Value: "available"})
	return initial
}

func (s *ModelSerializationSuite) TestParsingOptionals(c *gc.C) {
	args := ModelArgs{
		Owner: names.NewUserTag("magic"),
		Config: map[string]interface{}{
			"name": "awesome",
			"uuid": "some-uuid",
		},
	}
	initial := s.newModel(args)
	model := s.exportImport(c, initial)
	c.Assert(model.LatestToolsVersion(), gc.Equals, version.Zero)
}

func (s *ModelSerializationSuite) TestAnnotations(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	annotations := map[string]string{
		"string":  "value",
		"another": "one",
	}
	initial.SetAnnotations(annotations)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Annotations(), jc.DeepEquals, annotations)
}

func (s *ModelSerializationSuite) TestSequences(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	initial.SetSequence("machine", 4)
	initial.SetSequence("application-foo", 3)
	initial.SetSequence("application-bar", 1)
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(model.Sequences(), jc.DeepEquals, map[string]int{
		"machine":         4,
		"application-foo": 3,
		"application-bar": 1,
	})
}

func (s *ModelSerializationSuite) TestConstraints(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := ConstraintsArgs{
		Architecture: "amd64",
		Memory:       8 * gig,
	}
	initial.SetConstraints(args)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Constraints(), jc.DeepEquals, newConstraints(args))
}

func (*ModelSerializationSuite) TestModelValidation(c *gc.C) {
	model := NewModel(ModelArgs{})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, "missing model owner not valid")
	c.Assert(err, jc.Satisfies, errors.IsNotValid)
}

func (*ModelSerializationSuite) TestModelValidationMissingStatus(c *gc.C) {
	model := NewModel(ModelArgs{Owner: names.NewUserTag("owner")})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, "missing status not valid")
	c.Assert(err, jc.Satisfies, errors.IsNotValid)
}

func (s *ModelSerializationSuite) TestModelValidationChecksMachines(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner"), CloudRegion: "some-region"})
	model.AddMachine(MachineArgs{})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, "machine missing id not valid")
	c.Assert(err, jc.Satisfies, errors.IsNotValid)
}

func (s *ModelSerializationSuite) addMachineToModel(model Model, id string) Machine {
	machine := model.AddMachine(MachineArgs{Id: names.NewMachineTag(id)})
	machine.SetInstance(CloudInstanceArgs{InstanceId: "magic"})
	machine.SetTools(minimalAgentToolsArgs())
	machine.SetStatus(minimalStatusArgs())
	machine.Instance().SetStatus(minimalStatusArgs())
	machine.Instance().SetModificationStatus(minimalStatusArgs())
	return machine
}

func (s *ModelSerializationSuite) TestModelValidationChecksMachinesGood(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner"), CloudRegion: "some-region"})
	s.addMachineToModel(model, "0")
	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestModelValidationChecksOpenPortsUnits(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner"), CloudRegion: "some-region"})
	machine := s.addMachineToModel(model, "0")
	machine.AddOpenedPorts(OpenedPortsArgs{
		OpenedPorts: []PortRangeArgs{
			{
				UnitName: "missing/0",
				FromPort: 8080,
				ToPort:   8080,
				Protocol: "tcp",
			},
		},
	})
	err := model.Validate()
	c.Assert(err.Error(), gc.Equals, "unknown unit names in open ports: [missing/0]")
}

func (s *ModelSerializationSuite) TestModelValidationChecksApplications(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner"), CloudRegion: "some-region"})
	model.AddApplication(ApplicationArgs{})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, "application missing name not valid")
	c.Assert(err, jc.Satisfies, errors.IsNotValid)
}

func (s *ModelSerializationSuite) addApplicationToModel(model Model, name string, numUnits int) Application {
	application := model.AddApplication(ApplicationArgs{
		Tag:                names.NewApplicationTag(name),
		CharmConfig:        map[string]interface{}{},
		LeadershipSettings: map[string]interface{}{},
	})
	application.SetStatus(minimalStatusArgs())
	for i := 0; i < numUnits; i++ {
		// The index i is used as both the machine id and the unit id.
		// A happy coincidence.
		machine := s.addMachineToModel(model, fmt.Sprint(i))
		unit := application.AddUnit(UnitArgs{
			Tag:     names.NewUnitTag(fmt.Sprintf("%s/%d", name, i)),
			Machine: machine.Tag(),
		})
		unit.SetTools(minimalAgentToolsArgs())
		unit.SetAgentStatus(minimalStatusArgs())
		unit.SetWorkloadStatus(minimalStatusArgs())
	}

	return application
}

func (s *ModelSerializationSuite) wordpressModel() (Model, Endpoint, Endpoint) {
	model := s.newModel(ModelArgs{
		Owner: names.NewUserTag("owner"),
		Config: map[string]interface{}{
			"uuid": "some-uuid",
		},
		CloudRegion: "some-region",
	})
	s.addApplicationToModel(model, "wordpress", 2)
	s.addApplicationToModel(model, "mysql", 1)

	// Add a relation between wordpress and mysql.
	rel := model.AddRelation(RelationArgs{
		Id:  42,
		Key: "special key",
	})
	rel.SetStatus(minimalStatusArgs())

	wordpressEndpoint := rel.AddEndpoint(EndpointArgs{
		ApplicationName: "wordpress",
		Name:            "db",
		// Ignoring other aspects of endpoints.
	})
	mysqlEndpoint := rel.AddEndpoint(EndpointArgs{
		ApplicationName: "mysql",
		Name:            "mysql",
		// Ignoring other aspects of endpoints.
	})
	return model, wordpressEndpoint, mysqlEndpoint
}

func (s *ModelSerializationSuite) wordpressModelWithSettings() Model {
	model, wordpressEndpoint, mysqlEndpoint := s.wordpressModel()

	s.setEndpointSettings(wordpressEndpoint, "wordpress/0", "wordpress/1")
	s.setEndpointSettings(mysqlEndpoint, "mysql/0")

	return model
}

func (s *ModelSerializationSuite) TestModelValidationChecksRelationsMissingSettings(c *gc.C) {
	model, _, _ := s.wordpressModel()
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, "missing relation settings for units \\[wordpress/0 wordpress/1\\] in relation 42")
}

func (s *ModelSerializationSuite) TestModelValidationChecksRelationsMissingSettings2(c *gc.C) {
	model, wordpressEndpoint, _ := s.wordpressModel()

	s.setEndpointSettings(wordpressEndpoint, "wordpress/0", "wordpress/1")

	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, "missing relation settings for units \\[mysql/0\\] in relation 42")
}

func (s *ModelSerializationSuite) TestModelValidationChecksRelations(c *gc.C) {
	model := s.wordpressModelWithSettings()
	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) addSubordinateEndpoints(c *gc.C, rel Relation, app string) (Endpoint, Endpoint) {
	appEndpoint := rel.AddEndpoint(EndpointArgs{
		ApplicationName: app,
		Name:            "logging",
		Scope:           "container",
		// Ignoring other aspects of endpoints.
	})
	loggingEndpoint := rel.AddEndpoint(EndpointArgs{
		ApplicationName: "logging",
		Name:            "logging",
		Scope:           "container",
		// Ignoring other aspects of endpoints.
	})
	return appEndpoint, loggingEndpoint
}

func (s *ModelSerializationSuite) setEndpointSettings(ep Endpoint, units ...string) {
	for _, unit := range units {
		ep.SetUnitSettings(unit, map[string]interface{}{
			"key": "value",
		})
	}
}

func (s *ModelSerializationSuite) TestModelValidationChecksRelationsWithSubordinates(c *gc.C) {
	model := s.wordpressModelWithSettings()

	s.addApplicationToModel(model, "logging", 3)

	// Add a subordinate relations between logging and both wordpress and mysql.
	rel := model.AddRelation(RelationArgs{
		Id:  43,
		Key: "some key",
	})
	wordpressEndpoint, loggingEndpoint := s.addSubordinateEndpoints(c, rel, "wordpress")
	s.setEndpointSettings(wordpressEndpoint, "wordpress/0", "wordpress/1")
	s.setEndpointSettings(loggingEndpoint, "logging/0", "logging/1")

	rel = model.AddRelation(RelationArgs{
		Id:  44,
		Key: "other key",
	})
	mysqlEndpoint, loggingEndpoint := s.addSubordinateEndpoints(c, rel, "mysql")
	s.setEndpointSettings(mysqlEndpoint, "mysql/0")
	s.setEndpointSettings(loggingEndpoint, "logging/2")

	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestModelSerializationWithRelations(c *gc.C) {
	initial := s.wordpressModelWithSettings()
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)
	result, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, jc.DeepEquals, initial)
}

func (s *ModelSerializationSuite) TestModelSerializationWithRemoteEntities(c *gc.C) {
	model := s.newModel(ModelArgs{
		Owner: names.NewUserTag("owner"),
		Config: map[string]interface{}{
			"uuid": "some-uuid",
		},
		CloudRegion: "some-region",
	})
	model.AddRemoteEntity(RemoteEntityArgs{
		Token: "xxx-aaa-bbb",
	})
	model.AddRemoteEntity(RemoteEntityArgs{
		Token:    "zzz-ccc-yyy",
		Macaroon: "some-macaroon-that-should-be-discharged",
	})
	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
	bytes, err := yaml.Marshal(model)
	c.Assert(err, jc.ErrorIsNil)
	result, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, jc.DeepEquals, model)
}

func (s *ModelSerializationSuite) TestModelSerializationWithRelationNetworks(c *gc.C) {
	model := s.newModel(ModelArgs{
		Owner: names.NewUserTag("owner"),
		Config: map[string]interface{}{
			"uuid": "some-uuid",
		},
		CloudRegion: "some-region",
	})
	model.AddRelationNetwork(RelationNetworkArgs{
		ID:          names.NewControllerTag("ctrl-uuid-3").String(),
		RelationKey: "relation-key",
		CIDRS:       []string{"10.0.1.0/16"},
	})
	model.AddRelationNetwork(RelationNetworkArgs{
		ID:          names.NewControllerTag("ctrl-uuid-4").String(),
		RelationKey: "relation-key",
		CIDRS:       []string{"12.0.1.1/24"},
	})
	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
	bytes, err := yaml.Marshal(model)
	c.Assert(err, jc.ErrorIsNil)
	result, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, jc.DeepEquals, model)
}

func (s *ModelSerializationSuite) TestModelValidationChecksSubnets(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	model.AddSubnet(SubnetArgs{CIDR: "10.0.0.0/24", SpaceID: "3"})
	model.AddSubnet(SubnetArgs{CIDR: "10.0.1.0/24"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `subnet "10.0.0.0/24" references non-existent space "3"`)
	model.AddSpace(SpaceArgs{Id: "3"})
	err = model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressMachineID(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	model.AddIPAddress(IPAddressArgs{Value: "192.168.1.0", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ip address "192.168.1.0" references non-existent machine "42"`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressDeviceName(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := IPAddressArgs{Value: "192.168.1.0", MachineID: "42", DeviceName: "foo"}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ip address "192.168.1.0" references non-existent device "foo"`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressValueEmpty(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := IPAddressArgs{MachineID: "42", DeviceName: "foo"}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ip address has invalid value ""`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressValueInvalid(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := IPAddressArgs{MachineID: "42", DeviceName: "foo", Value: "foobar"}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ip address has invalid value "foobar"`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressSubnetEmpty(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := IPAddressArgs{MachineID: "42", DeviceName: "foo", Value: "192.168.1.1"}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ip address "192.168.1.1" has empty subnet CIDR`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressSubnetInvalid(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := IPAddressArgs{
		MachineID:  "42",
		DeviceName: "foo",
		Value:      "192.168.1.1",
		SubnetCIDR: "foo",
	}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ip address "192.168.1.1" has invalid subnet CIDR "foo"`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressSucceeds(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := IPAddressArgs{
		MachineID:  "42",
		DeviceName: "foo",
		Value:      "192.168.1.1",
		SubnetCIDR: "192.168.1.0/24",
	}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressGatewayAddressInvalid(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := IPAddressArgs{
		MachineID:      "42",
		DeviceName:     "foo",
		Value:          "192.168.1.1",
		SubnetCIDR:     "192.168.1.0/24",
		GatewayAddress: "foo",
	}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ip address "192.168.1.1" has invalid gateway address "foo"`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressGatewayAddressValid(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := IPAddressArgs{
		MachineID:        "42",
		DeviceName:       "foo",
		Value:            "192.168.1.2",
		SubnetCIDR:       "192.168.1.0/24",
		GatewayAddress:   "192.168.1.1",
		IsDefaultGateway: true,
	}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestModelValidationChecksLinkLayerDeviceMachineId(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `device "foo" references non-existent machine "42"`)
	s.addMachineToModel(model, "42")
	err = model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestModelValidationChecksLinkLayerName(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{MachineID: "42"})
	s.addMachineToModel(model, "42")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, "device has empty name.*")
}

func (s *ModelSerializationSuite) TestModelValidationChecksLinkLayerMACAddress(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := LinkLayerDeviceArgs{MachineID: "42", Name: "foo", MACAddress: "DEADBEEF"}
	model.AddLinkLayerDevice(args)
	s.addMachineToModel(model, "42")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `device "foo" has invalid MACAddress "DEADBEEF"`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksParentExists(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := LinkLayerDeviceArgs{MachineID: "42", Name: "foo", ParentName: "bar", MACAddress: "01:23:45:67:89:ab"}
	model.AddLinkLayerDevice(args)
	s.addMachineToModel(model, "42")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `device "foo" has non-existent parent "bar"`)
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "bar", MachineID: "42"})
	err = model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestModelValidationChecksParentIsNotItself(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := LinkLayerDeviceArgs{MachineID: "42", Name: "foo", ParentName: "foo"}
	model.AddLinkLayerDevice(args)
	s.addMachineToModel(model, "42")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `device "foo" is its own parent`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksParentIsABridge(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := LinkLayerDeviceArgs{MachineID: "42", Name: "foo", ParentName: "m#43#d#bar"}
	model.AddLinkLayerDevice(args)
	args2 := LinkLayerDeviceArgs{MachineID: "43", Name: "bar"}
	model.AddLinkLayerDevice(args2)
	s.addMachineToModel(model, "42")
	s.addMachineToModel(model, "43")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `device "foo" on a container but not a bridge`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksChildDeviceContained(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := LinkLayerDeviceArgs{MachineID: "42", Name: "foo", ParentName: "m#43#d#bar"}
	model.AddLinkLayerDevice(args)
	args2 := LinkLayerDeviceArgs{MachineID: "43", Name: "bar", Type: "bridge"}
	model.AddLinkLayerDevice(args2)
	s.addMachineToModel(model, "42")
	s.addMachineToModel(model, "43")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ParentName "m#43#d#bar" for non-container machine "42"`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksParentOnHost(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := LinkLayerDeviceArgs{MachineID: "41/lxd/0", Name: "foo", ParentName: "m#43#d#bar"}
	model.AddLinkLayerDevice(args)
	args2 := LinkLayerDeviceArgs{MachineID: "43", Name: "bar", Type: "bridge"}
	model.AddLinkLayerDevice(args2)
	machine := s.addMachineToModel(model, "41")
	container := machine.AddContainer(MachineArgs{Id: names.NewMachineTag("41/lxd/0")})
	container.SetInstance(CloudInstanceArgs{InstanceId: "magic"})
	container.Instance().SetStatus(minimalStatusArgs())
	container.SetTools(minimalAgentToolsArgs())
	container.SetStatus(minimalStatusArgs())
	s.addMachineToModel(model, "43")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `parent machine of device "foo" not host machine "41"`)
}

func (s *ModelSerializationSuite) TestModelValidationLinkLayerDeviceContainer(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	args := LinkLayerDeviceArgs{MachineID: "43/lxd/0", Name: "foo", ParentName: "m#43#d#bar"}
	model.AddLinkLayerDevice(args)
	args2 := LinkLayerDeviceArgs{MachineID: "43", Name: "bar", Type: "bridge"}
	model.AddLinkLayerDevice(args2)
	machine := s.addMachineToModel(model, "43")
	container := machine.AddContainer(MachineArgs{Id: names.NewMachineTag("43/lxd/0")})
	container.SetInstance(CloudInstanceArgs{InstanceId: "magic"})
	container.Instance().SetStatus(minimalStatusArgs())
	container.SetTools(minimalAgentToolsArgs())
	container.SetStatus(minimalStatusArgs())
	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestNewModelSetsRemoteApplications(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("mogwai")})
	c.Assert(model.RemoteApplications(), gc.IsNil)
}

func (s *ModelSerializationSuite) TestModelValidationHandlesRemoteApplications(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("ink-spots")})
	remoteApp := model.AddRemoteApplication(RemoteApplicationArgs{
		Tag:             names.NewApplicationTag("mysql"),
		OfferUUID:       "offer-uuid",
		URL:             "other.mysql",
		SourceModel:     names.NewModelTag("some-model"),
		IsConsumerProxy: true,
	})
	remoteApp.AddEndpoint(RemoteEndpointArgs{
		Name:      "db",
		Role:      "provider",
		Interface: "mysql",
	})

	s.addApplicationToModel(model, "wordpress", 1)
	rel := model.AddRelation(RelationArgs{
		Id:  101,
		Key: "wordpress:db mysql:db",
	})
	rel.AddEndpoint(EndpointArgs{
		ApplicationName: "wordpress",
		Name:            "db",
	})
	rel.AddEndpoint(EndpointArgs{
		ApplicationName: "mysql",
		Name:            "db",
	})

	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func asStringMap(c *gc.C, model Model) map[string]interface{} {
	bytes, err := Serialize(model)
	c.Assert(err, jc.ErrorIsNil)

	var data map[string]interface{}
	err = yaml.Unmarshal(bytes, &data)
	c.Assert(err, jc.ErrorIsNil)
	return data
}

func (s *ModelSerializationSuite) TestSerializesRemoteApplications(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("veils")})
	rapp := model.AddRemoteApplication(RemoteApplicationArgs{
		Tag:             names.NewApplicationTag("bloom"),
		OfferUUID:       "offer-uuid",
		URL:             "other.mysql",
		SourceModel:     names.NewModelTag("some-model"),
		IsConsumerProxy: true,
	})
	rapp.AddEndpoint(RemoteEndpointArgs{
		Name:      "db",
		Role:      "provider",
		Interface: "mysql",
	})
	rapp.SetStatus(StatusArgs{
		Value:   "running",
		Updated: time.Date(2017, 5, 9, 12, 1, 0, 0, time.UTC),
	})
	data := asStringMap(c, model)
	remoteSection, ok := data["remote-applications"]
	c.Assert(ok, jc.IsTrue)

	// Re-serialize just that bit so we can check it.
	bytes, err := yaml.Marshal(remoteSection)
	c.Assert(err, jc.ErrorIsNil)

	expected := `
remote-applications:
- endpoints:
    endpoints:
    - interface: mysql
      name: db
      role: provider
    version: 1
  is-consumer-proxy: true
  name: bloom
  offer-uuid: offer-uuid
  source-model-uuid: some-model
  spaces:
    spaces: []
    version: 1
  status:
    status:
      neverset: false
      updated: "2017-05-09T12:01:00Z"
      value: running
    version: 2
  url: other.mysql
version: 2
`[1:]
	c.Assert(string(bytes), gc.Equals, expected)
}

func (s *ModelSerializationSuite) TestImportingWithRemoteApplications(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("veils")})
	rapp := initial.AddRemoteApplication(RemoteApplicationArgs{
		Tag:             names.NewApplicationTag("bloom"),
		OfferUUID:       "offer-uuid",
		URL:             "other.mysql",
		SourceModel:     names.NewModelTag("some-model"),
		IsConsumerProxy: true,
	})
	rapp.AddEndpoint(RemoteEndpointArgs{
		Name:      "db",
		Role:      "provider",
		Interface: "mysql",
	})
	rapp.SetStatus(StatusArgs{
		Value:   "hey",
		Updated: time.Now(),
	})
	remoteApplications := initial.RemoteApplications()

	bytes, err := Serialize(initial)
	c.Assert(err, jc.ErrorIsNil)

	result, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	ra := result.RemoteApplications()
	c.Assert(ra, gc.HasLen, 1)
	c.Assert(ra[0], gc.DeepEquals, remoteApplications[0])
}

func (s *ModelSerializationSuite) TestRemoteApplicationsGetter(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("veils")})
	model.AddRemoteApplication(RemoteApplicationArgs{
		Tag:             names.NewApplicationTag("bloom"),
		OfferUUID:       "offer-uuid",
		URL:             "other.mysql",
		SourceModel:     names.NewModelTag("some-model"),
		IsConsumerProxy: true,
	})
	result := model.RemoteApplications()
	c.Assert(result, gc.HasLen, 1)
}

func (s *ModelSerializationSuite) TestSerializesOfferConnections(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("veils")})
	model.AddOfferConnection(OfferConnectionArgs{
		OfferUUID:       "offer-uuid",
		RelationID:      1,
		RelationKey:     "relation-key",
		SourceModelUUID: "some-model-uuid",
		UserName:        "fred",
	})
	data := asStringMap(c, model)
	offerSection, ok := data["offer-connections"]
	c.Assert(ok, jc.IsTrue)

	// Re-serialize just that bit so we can check it.
	bytes, err := yaml.Marshal(offerSection)
	c.Assert(err, jc.ErrorIsNil)

	expected := `
offer-connections:
- offer-uuid: offer-uuid
  relation-id: 1
  relation-key: relation-key
  source-model-uuid: some-model-uuid
  user-name: fred
version: 1
`[1:]
	c.Assert(string(bytes), gc.Equals, expected)
}

func (s *ModelSerializationSuite) TestImportingWithOfferConnections(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("veils")})
	initial.AddOfferConnection(OfferConnectionArgs{
		OfferUUID:       "offer-uuid",
		RelationID:      1,
		RelationKey:     "relation-key",
		SourceModelUUID: "some-model-uuid",
		UserName:        "fred",
	})
	offerConnections := initial.OfferConnections()

	bytes, err := Serialize(initial)
	c.Assert(err, jc.ErrorIsNil)

	result, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	offers := result.OfferConnections()
	c.Assert(offers, gc.HasLen, 1)
	c.Assert(offers[0], gc.DeepEquals, offerConnections[0])
}

func (s *ModelSerializationSuite) TestOfferConnectionsGetter(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("veils")})
	model.AddOfferConnection(OfferConnectionArgs{
		OfferUUID:       "offer-uuid",
		RelationID:      1,
		RelationKey:     "relation-key",
		SourceModelUUID: "some-model-uuid",
		UserName:        "fred",
	})
	result := model.OfferConnections()
	c.Assert(result, gc.HasLen, 1)
}

func (s *ModelSerializationSuite) TestSerializesExternalControllers(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("veils")})
	model.AddExternalController(ExternalControllerArgs{
		Tag:    names.NewControllerTag("controller-name"),
		Alias:  "moon-ball",
		Addrs:  []string{"1.2.3.4", "10.12.11.243"},
		CACert: "magic-cert",
		Models: []string{"aaaa-bbbb"},
	})
	data := asStringMap(c, model)
	ctrlSection, ok := data["external-controllers"]
	c.Assert(ok, jc.IsTrue)

	// Re-serialize just that bit so we can check it.
	bytes, err := yaml.Marshal(ctrlSection)
	c.Assert(err, jc.ErrorIsNil)

	expected := `
external-controllers:
- addrs:
  - 1.2.3.4
  - 10.12.11.243
  alias: moon-ball
  ca-cert: magic-cert
  id: controller-name
  models:
  - aaaa-bbbb
version: 1
`[1:]
	c.Assert(string(bytes), gc.Equals, expected)
}

func (s *ModelSerializationSuite) TestImportingWithExternalControllers(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("veils")})
	initial.AddExternalController(ExternalControllerArgs{
		Tag:    names.NewControllerTag("controller-name"),
		Alias:  "moon-ball",
		Addrs:  []string{"1.2.3.4", "10.12.11.243"},
		CACert: "magic-cert",
		Models: []string{"aaaa-bbbb"},
	})
	offerConnections := initial.ExternalControllers()

	bytes, err := Serialize(initial)
	c.Assert(err, jc.ErrorIsNil)

	result, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	offers := result.ExternalControllers()
	c.Assert(offers, gc.HasLen, 1)
	c.Assert(offers[0], gc.DeepEquals, offerConnections[0])
}

func (s *ModelSerializationSuite) TestExternalControllersGetter(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("veils")})
	model.AddExternalController(ExternalControllerArgs{
		Tag:    names.NewControllerTag("controller-name"),
		Alias:  "moon-ball",
		Addrs:  []string{"1.2.3.4", "10.12.11.243"},
		CACert: "magic-cert",
	})
	result := model.ExternalControllers()
	c.Assert(result, gc.HasLen, 1)
}

func (s *ModelSerializationSuite) TestSetAndGetSLA(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	sla := model.SetSLA("essential", "bob", "creds")
	c.Assert(sla.Level(), gc.Equals, "essential")
	c.Assert(sla.Owner(), gc.Equals, "bob")
	c.Assert(sla.Credentials(), gc.Equals, "creds")

	getSla := model.SLA()
	c.Assert(getSla.Level(), gc.Equals, sla.Level())
	c.Assert(getSla.Credentials(), jc.DeepEquals, sla.Credentials())
}

func (s *ModelSerializationSuite) TestSLA(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	sla := initial.SetSLA("essential", "bob", "creds")
	c.Assert(sla.Level(), gc.Equals, "essential")
	c.Assert(sla.Owner(), gc.Equals, "bob")
	c.Assert(sla.Credentials(), gc.Equals, "creds")

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.SLA().Level(), gc.Equals, "essential")
	c.Assert(model.SLA().Owner(), gc.Equals, "bob")
	c.Assert(model.SLA().Credentials(), gc.Equals, "creds")
}

func (s *ModelSerializationSuite) TestGetAndSetMeterStatus(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("veils")})
	ms := model.SetMeterStatus("RED", "info message")
	c.Assert(ms.Code(), gc.Equals, "RED")
	c.Assert(ms.Info(), gc.Equals, "info message")

	getms := model.MeterStatus()
	c.Assert(getms.Code(), gc.Equals, ms.Code())
	c.Assert(getms.Info(), gc.Equals, ms.Info())
}

func (s *ModelSerializationSuite) TestMeterStatus(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	ms := initial.SetMeterStatus("RED", "info message")
	c.Assert(ms.Code(), gc.Equals, "RED")
	c.Assert(ms.Info(), gc.Equals, "info message")

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.MeterStatus().Code(), gc.Equals, "RED")
	c.Assert(model.MeterStatus().Info(), gc.Equals, "info message")
}

func (s *ModelSerializationSuite) TestSerializesToLatestVersion(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("ben-harper")})
	data := asStringMap(c, initial)
	versionValue, ok := data["version"]
	c.Assert(ok, jc.IsTrue)
	version, ok := versionValue.(int)
	c.Assert(ok, jc.IsTrue)
	c.Assert(version, gc.Equals, 7)
}

func (s *ModelSerializationSuite) TestVersion1Works(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("ben-harper")})
	data := asStringMap(c, initial)
	data["version"] = 1

	bytes, err := yaml.Marshal(data)
	c.Assert(err, jc.ErrorIsNil)
	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(model.Owner(), gc.Equals, names.NewUserTag("ben-harper"))
	c.Assert(model.Type(), gc.Equals, IAAS)
}

func (s *ModelSerializationSuite) TestVersion1IgnoresRemoteApplications(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("ben-harper")})
	initial.AddRemoteApplication(RemoteApplicationArgs{
		Tag:             names.NewApplicationTag("bloom"),
		OfferUUID:       "offer-uuid",
		URL:             "other.mysql",
		SourceModel:     names.NewModelTag("some-model"),
		IsConsumerProxy: true,
	})
	data := asStringMap(c, initial)
	data["version"] = 1

	bytes, err := yaml.Marshal(data)
	c.Assert(err, jc.ErrorIsNil)
	result, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)

	// Doesn't import the remote applications - version 1 models
	// didn't know about them.
	c.Assert(result.RemoteApplications(), gc.HasLen, 0)
}

func (s *ModelSerializationSuite) TestSpaces(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	space := initial.AddSpace(SpaceArgs{Id: "1", Name: "special"})
	c.Assert(space.Name(), gc.Equals, "special")
	c.Assert(space.Id(), gc.Equals, "1")

	spaces := initial.Spaces()
	c.Assert(spaces, gc.HasLen, 1)
	c.Assert(spaces[0], gc.Equals, space)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Spaces(), jc.DeepEquals, spaces)

}

func (s *ModelSerializationSuite) TestLinkLayerDevice(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	device := initial.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo"})
	c.Assert(device.Name(), gc.Equals, "foo")
	devices := initial.LinkLayerDevices()
	c.Assert(devices, gc.HasLen, 1)
	c.Assert(devices[0], jc.DeepEquals, device)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	c.Logf(string(bytes))

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.LinkLayerDevices(), jc.DeepEquals, devices)
}

func (s *ModelSerializationSuite) TestSubnets(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	initial.AddSubnet(SubnetArgs{CIDR: "10.0.20.0/24", SpaceID: "0"})
	subnet := initial.AddSubnet(SubnetArgs{CIDR: "10.0.0.0/24"})
	c.Assert(subnet.CIDR(), gc.Equals, "10.0.0.0/24")
	subnets := initial.Subnets()
	c.Assert(subnets, gc.HasLen, 2)
	c.Assert(subnets[1], jc.DeepEquals, subnet)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Subnets(), jc.DeepEquals, subnets)
}

func (s *ModelSerializationSuite) TestIPAddress(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	addr := initial.AddIPAddress(IPAddressArgs{Value: "10.0.0.4"})
	c.Assert(addr.Value(), gc.Equals, "10.0.0.4")
	addresses := initial.IPAddresses()
	c.Assert(addresses, gc.HasLen, 1)
	c.Assert(addresses[0], jc.DeepEquals, addr)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.IPAddresses(), jc.DeepEquals, addresses)
}

func (s *ModelSerializationSuite) TestSSHHostKey(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	key := initial.AddSSHHostKey(SSHHostKeyArgs{MachineID: "foo"})
	c.Assert(key.MachineID(), gc.Equals, "foo")
	keys := initial.SSHHostKeys()
	c.Assert(keys, gc.HasLen, 1)
	c.Assert(keys[0], jc.DeepEquals, key)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.SSHHostKeys(), jc.DeepEquals, keys)
}

func (s *ModelSerializationSuite) TestCloudImageMetadata(c *gc.C) {
	storageSize := uint64(3)
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	image := initial.AddCloudImageMetadata(CloudImageMetadataArgs{
		Stream:          "stream",
		Region:          "region-test",
		Version:         "14.04",
		Series:          "trusty",
		Arch:            "arch",
		VirtType:        "virtType-test",
		RootStorageType: "rootStorageType-test",
		RootStorageSize: &storageSize,
		Source:          "test",
		Priority:        2,
		ImageId:         "1",
		DateCreated:     2,
	})
	c.Assert(image.Stream(), gc.Equals, "stream")
	c.Assert(image.Region(), gc.Equals, "region-test")
	c.Assert(image.Version(), gc.Equals, "14.04")
	c.Assert(image.Arch(), gc.Equals, "arch")
	c.Assert(image.VirtType(), gc.Equals, "virtType-test")
	c.Assert(image.RootStorageType(), gc.Equals, "rootStorageType-test")
	value, ok := image.RootStorageSize()
	c.Assert(ok, jc.IsTrue)
	c.Assert(value, gc.Equals, uint64(3))
	c.Assert(image.Source(), gc.Equals, "test")
	c.Assert(image.Priority(), gc.Equals, 2)
	c.Assert(image.ImageId(), gc.Equals, "1")
	c.Assert(image.DateCreated(), gc.Equals, int64(2))

	metadata := initial.CloudImageMetadata()
	c.Assert(metadata, gc.HasLen, 1)
	c.Assert(metadata[0], jc.DeepEquals, image)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.CloudImageMetadata(), jc.DeepEquals, metadata)
}

func (s *ModelSerializationSuite) TestAction(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	enqueued := time.Now().UTC()
	action := initial.AddAction(ActionArgs{
		Name:       "foo",
		Enqueued:   enqueued,
		Parameters: map[string]interface{}{},
		Results:    map[string]interface{}{},
	})
	c.Assert(action.Name(), gc.Equals, "foo")
	c.Assert(action.Enqueued(), gc.Equals, enqueued)
	actions := initial.Actions()
	c.Assert(actions, gc.HasLen, 1)
	c.Assert(actions[0], jc.DeepEquals, action)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Actions(), jc.DeepEquals, actions)
}

func (s *ModelSerializationSuite) TestOperation(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	enqueued := time.Now().UTC()
	op := initial.AddOperation(OperationArgs{
		Summary:           "foo",
		Enqueued:          enqueued,
		CompleteTaskCount: 666,
	})
	c.Assert(op.Summary(), gc.Equals, "foo")
	c.Assert(op.CompleteTaskCount(), gc.Equals, 666)
	c.Assert(op.Enqueued(), gc.Equals, enqueued)
	operations := initial.Operations()
	c.Assert(operations, gc.HasLen, 1)
	c.Assert(operations[0], jc.DeepEquals, op)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Operations(), jc.DeepEquals, operations)
}

func (s *ModelSerializationSuite) TestVolumeValidation(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	model.AddVolume(testVolumeArgs())
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `volume\[0\]: volume "1234" missing status not valid`)
}

func (s *ModelSerializationSuite) TestVolumes(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	volume := initial.AddVolume(testVolumeArgs())
	volume.SetStatus(minimalStatusArgs())
	volumes := initial.Volumes()
	c.Assert(volumes, gc.HasLen, 1)
	c.Assert(volumes[0], gc.Equals, volume)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Volumes(), jc.DeepEquals, volumes)
}

func (s *ModelSerializationSuite) TestFilesystemValidation(c *gc.C) {
	model := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	model.AddFilesystem(testFilesystemArgs())
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `filesystem\[0\]: filesystem "1234" missing status not valid`)
}

func (s *ModelSerializationSuite) TestFilesystems(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	filesystem := initial.AddFilesystem(testFilesystemArgs())
	filesystem.SetStatus(minimalStatusArgs())
	filesystem.AddAttachment(testFilesystemAttachmentArgs())
	filesystems := initial.Filesystems()
	c.Assert(filesystems, gc.HasLen, 1)
	c.Assert(filesystems[0], gc.Equals, filesystem)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Filesystems(), jc.DeepEquals, filesystems)
}

func (s *ModelSerializationSuite) TestFirewallRule(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	firewallRule := initial.AddFirewallRule(MinimalFireWallArgs())
	firewallRules := initial.FirewallRules()
	c.Assert(firewallRules, gc.HasLen, 1)
	c.Assert(firewallRules[0], jc.DeepEquals, firewallRule)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.FirewallRules(), jc.DeepEquals, firewallRules)
}

func (s *ModelSerializationSuite) TestStorage(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	storage := initial.AddStorage(testStorageArgs())
	storages := initial.Storages()
	c.Assert(storages, gc.HasLen, 1)
	c.Assert(storages[0], jc.DeepEquals, storage)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Storages(), jc.DeepEquals, storages)
}

func (s *ModelSerializationSuite) TestStorageValidate(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	storageArgs := testStorageArgs()
	storageArgs.Owner = nil
	storage := initial.AddStorage(storageArgs)
	storages := initial.Storages()
	c.Assert(storages, gc.HasLen, 1)
	c.Assert(storages[0], jc.DeepEquals, storage)

	err := initial.Validate()
	c.Assert(err, gc.ErrorMatches, `storage\[0\] attachment referencing unknown unit "unit-postgresql-0" not valid`)
}

func (s *ModelSerializationSuite) TestStoragePools(c *gc.C) {
	initial := s.newModel(ModelArgs{Owner: names.NewUserTag("owner")})
	poolOne := map[string]interface{}{
		"foo":   42,
		"value": true,
	}
	poolTwo := map[string]interface{}{
		"value": "spanner",
	}
	initial.AddStoragePool(StoragePoolArgs{
		Name: "one", Provider: "sparkles", Attributes: poolOne})
	initial.AddStoragePool(StoragePoolArgs{
		Name: "two", Provider: "spanner", Attributes: poolTwo})

	pools := initial.StoragePools()
	c.Assert(pools, gc.HasLen, 2)
	one, two := pools[0], pools[1]
	c.Check(one.Name(), gc.Equals, "one")
	c.Check(one.Provider(), gc.Equals, "sparkles")
	c.Check(one.Attributes(), jc.DeepEquals, poolOne)
	c.Check(two.Name(), gc.Equals, "two")
	c.Check(two.Provider(), gc.Equals, "spanner")
	c.Check(two.Attributes(), jc.DeepEquals, poolTwo)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)

	pools = model.StoragePools()
	c.Assert(pools, gc.HasLen, 2)
	one, two = pools[0], pools[1]
	c.Check(one.Name(), gc.Equals, "one")
	c.Check(one.Provider(), gc.Equals, "sparkles")
	c.Check(one.Attributes(), jc.DeepEquals, poolOne)
	c.Check(two.Name(), gc.Equals, "two")
	c.Check(two.Provider(), gc.Equals, "spanner")
	c.Check(two.Attributes(), jc.DeepEquals, poolTwo)
}

func (s *ModelSerializationSuite) TestStatus(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: names.NewUserTag("owner")})
	initial.SetStatus(minimalStatusArgs())
	expected := minimalStatus()
	c.Check(initial.Status(), jc.DeepEquals, expected)

	model := s.exportImport(c, initial)
	c.Check(model.Status(), jc.DeepEquals, expected)
}

// modelV1example was taken from a Juju 2.1 model dump, which is version
// 1, and among other things is missing model status, which version 2 makes
// manditory.
const modelV1example = `
actions:
  actions: []
  version: 1
applications:
  applications:
  - charm-mod-version: 0
    charm-url: cs:ubuntu-10
    cs-channel: stable
    leader: ubuntu/1
    leadership-settings: {}
    name: ubuntu
    resources:
      resources: []
      version: 1
    series: xenial
    settings: {}
    status:
      status:
        message: waiting for machine
        updated: 2017-02-21T19:47:23.691434191Z
        value: waiting
      version: 1
    status-history:
      history: []
      version: 1
    units:
      units:
      - agent-status:
          status:
            updated: 2017-03-28T02:53:17.758361087Z
            value: idle
          version: 1
        agent-status-history:
          history:
          - updated: 2017-03-28T02:53:17.758361087Z
            value: idle
          - message: running update-status hook
            updated: 2017-03-28T02:53:17.560360624Z
            value: executing
          - updated: 2017-03-28T02:48:17.827186321Z
            value: idle
          - message: running update-status hook
            updated: 2017-03-28T02:48:17.559780509Z
            value: executing
          - updated: 2017-03-28T02:43:17.78767961Z
            value: idle
          - message: running update-status hook
            updated: 2017-03-28T02:43:17.558255742Z
            value: executing
          - updated: 2017-03-28T02:38:17.758809684Z
            value: idle
          - message: running update-status hook
            updated: 2017-03-28T02:38:17.559809345Z
            value: executing
          - updated: 2017-03-28T02:33:17.806717092Z
            value: idle
          version: 1
        machine: "1"
        meter-status-code: NOT SET
        name: ubuntu/1
        password-hash: Il/M2+WlhkUA5zASHj+QvE66
        payloads:
          payloads: []
          version: 1
        resources:
          resources: []
          version: 1
        tools:
          sha256: ""
          size: 0
          tools-version: 2.1-rc1.1-xenial-amd64
          url: ""
          version: 1
        workload-status:
          status:
            message: ready
            updated: 2017-02-21T20:00:50.219219299Z
            value: active
          version: 1
        workload-status-history:
          history: []
          version: 1
        workload-version: "16.04"
        workload-version-history:
          history: []
          version: 1
      version: 1
  version: 1
cloud: dev
cloud-credential:
  attributes:
    client-cert: |+
      -----BEGIN CERTIFICATE-----
      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
      -----END CERTIFICATE-----

    client-key: |+
      -----BEGIN RSA PRIVATE KEY-----
      MIIJKwIBAAKCAgEAv5jHYlGsQl7SK+YXEGA1kSt2ZgA73P7g6WEPYxnBsdECd8T6
      HJJ7UB81bI4ZhlvEGE30EpvyRmk7n2ArdJveVewPHvZK2HZ1bgAbm6abUIKLMIVS
      CMJwvzymCYwhvwwtVTcUtzyD/AKPWlcd8R5TPRtsHvY8zk2vk8kq334U2ruIxYuE
      ZMVQdDD/u7kmd9xFGa1PJzySvdfpkATGGn9N442ARSLKKYeq7vLXQwhY4ibBbueC
      KNAfX91tG90DjP4wLFkFuc+iivQfDRys57HQzkTsu6n8GY7/A7hTchqP/Rc/rz6G
      ZJL5tlNNnChnlCtoDdP5m20gBWRh5gSAlUyAtwjtOagAH5fL+o5G8kQUF/W7nvNG
      OY/SidXbzxk3IHQlFgk9grtjfcBbfgfbzOvP6G6apLI91JM7NuheqsX3LNktvcJn
      7xLZc6+jZy+XZbc87hi2mIR7DVP5D9n17P0NBeMldaH9H7wZTC4iM0+VFcK/KvY0
      rTr8petVYQS3OuVym1qnmVMs0mei1bB6aFhDTFh+ruklMP47R9vkkq/501hR6dLx
      voOmAuKgcvFC2idGlW7lumroUUQ4+R8iYLgflllz9qUgc24Ue85/YjSnA5pE9Qma
      +fgDW9SAaO32KzGP8MPSBpVHVWNZewg7dq6ntp95DZ5QgN5khJt2bH2UyesCAwEA
      AQKCAgEAnB+HYROCVbbkdgcRBjQPklKjMOzB2wwKA8Imgq9rSVUyOJxW3k9yklBL
      /UAxhm1idheXak6O9lcr0WvRHT0hyEwJ6kXxqT+l4tPNq2DwsIMfnpMUzLR8oShv
      d9oraX0nt4ehpsp2FjWT0J6qdF8snt+ok/Y8iDN/feJvwBwCLFaiVa6hXGf4biW7
      TaHKueLQn+K8XBGc1XuKA+QP9WmE84bLXgrCR2MYo4tYY3P60ZWZC6E0S8ODvV82
      WH0ZNpuub4S/CLEsFyRj5RBXyJj5uAssCKO0HLOME+Dwhkskx6xZJZjCdFPbjXmG
      BVhbRr60FIiFneQnMp2gtIk4qe/j9ViJDYAp9trLeSbqLuGfDg3xNHWf6u9lgQ1e
      nCYUb/7ACmAs9lDONqyfcFHZEtFvRLjPJRWwye2VdG2UOMFR5FN+qnHouy8SxxpJ
      ss6iSBcgvd6yWEiJl/F6AI7wKzjAQOaQ7yG6ROb7oIZ8VHFA++IQJc1HG1F044fd
      aaW6RGHo4IsMoWrusRqcuwGmrdHmmo6xHMilExJlAP3vp6kNwxHBuSBPYgAZ2LwJ
      70oMDCk3q2xhONBwXo6yr3vEhg6X+tbe8/yDPC3h1RVRV6Zy3HLJXO2YQ49B2Q/R
      JKWFPRYEyM41dRLcni7OMOuRBqsk1boQ8OlyMFZLAmtIIw4gqvkCggEBANCdmEeU
      6+nkCr6b28RsVD3hq/CH3hp+/JgrYDKplug4XApDJwrJR+jXdADCD53zNloRfHSc
      oAqtaEirCbo+2DQNZOmcxuaI9DPPg5Bm/7+pTy59VAJ1KS391FNwLpjrPeXylAfG
      bUAeOySPFN0hiXBA1jVSqBXxw2yoBjUxOK87LOw4v6JBwe3OQN2FECMYKREgrLUQ
      w9KMJElc67+kypQqbhwkYzGitccV/h1MJDPsoPf2qef+CV6+n0pChnWJnEHCN46d
      w8FOAsslFjVlgu2orUYM1bjtsllm8iVkLyaEuhlPW0zCIF9VsPMvcD1GlBBarzZp
      qo1H+Om0jtB2VD0CggEBAOsdlp+M0ML9+U0KyUg02vo28iOaVGLTQnHjy64aktEX
      6G6R3KW2mLandxV62QsHAlsUwFjKqJLalWVh+wRQvpR6TUWicrPJIa307pLtHpiQ
      yGrB1iWjgJFVJgQprxqy/qrUKI+ZMO1yK1mFQrV665VLSH+lNWtzsywXRlOJ5Bsz
      250pWnGCy6t/xbkgXbj18b0mpLjPOu4VkJ2nLrVcxBV3c/UoHDuBzJVNqlcvlmMp
      ThVSFnsxCo4uy8HhsH3f52SVA3HEkDGg/tRfPoDz6p4YjzIRtHpY/4Z4H6xcEj3g
      kbzOjrEgQDFRkbCZZ9/zCul6NkyyXul/fCqCRn6u8UcCggEBAJXZDcAlLYw03K7r
      v2GJOr20c0/0GErJ+mDHj3L0tEwb56kLcWjjCf8re8zrmFIpmFn8A3yz4JWq16ST
      Dwq0B5LkKB9SGOERcPAiV+uKwJwIXrMTHRwi0jCSCkjg5Oe82hppM4GeO216CZuK
      Fz97zoBOYk+tDsWsgmZzDvp4X7Im/G61mJlRSJ6rr5Yu9VdwDFecM9Jft3luZsY3
      s7NWCJmDHNKkJIhLyuy1VHHw9nRBvaI/kO3uYQaoQD0UKgcVkKL/ge60Th91DLak
      7h3uy6wwpD4UDBSo0Jo9QyQuoVu2rQJvKTKqopw4LkGQSrwJDWPt77tTDUosb5RX
      uNnulTUCggEBANeaYH+bH+1P/QdbNs1SOuRs8osXgP7HAA10eVkE4VGA/RI4DACi
      e1Q0KY23A8WK/ewMEX7bCM7yR0GbIhcI/Fsn9ChBGbIoZQwiqYxuiToaus67Redq
      EgIz9RKoLvzq24JH35IfRrDXm00SWOQW/mX/jVIQa/ZHOSzbgxAkSNtxKJjsTRX+
      fUqddvGW7psoXi+4eiFHV6DwgZcwsjJ6CQ4uZlWQHKOtGbBociZVazEvtXzzs83w
      YN+Vph/7GF+1rXmc9HWlbR01p7mURbr28lVb7CRb/AaeCmSDT3g9TjUT9FERkeR3
      0KXpSRKK+qhxNbZ47cZTY5n34CMTKBYP0w0CggEBALN6qZGMPF9nEDFm77+Q82q5
      DuHbmbHJyG2lDrpFhlpAKnNNW9yqg8eQSW4GVP89Um/pkQTeSzo2aDec09P1Ld+O
      NxYltPHLZmlyxNRaEuG8XwUpy+stYUdex8jMhYMX4CkqNZLHVKC4e+cYHI24jj9W
      yeWPi6RYSvdpjJkLMPZpyBww3hiUbHdORa+WeyqeH9+XnhsFRdwQ0KcI656d4Exb
      ptaPkp3EJ+cxbqMSPrkG70abUR5Y4tu9RhM4Kkfcs6p745mp0C3elpBC0UKBPVjT
      rhD8s7nR/ywuwlh4TnEjvx33JmTOrctYRgm6C96yEAUPboVnrHDvN4TBsKe9wfc=
      -----END RSA PRIVATE KEY-----

    server-cert: |+
      -----BEGIN CERTIFICATE-----
      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
      -----END CERTIFICATE-----

  auth-type: certificate
  cloud: dev
  name: default
  owner: admin
  version: 1
cloud-image-metadata:
  cloudimagemetadata:
  - arch: amd64
    date-created: 1485305700000000000
    image-id: 698a8146-d6d9-4352-99fe-6557ebce5661
    priority: 10
    region: us-east-3
    root-storage-type: ""
    series: xenial
    source: default ubuntu cloud images
    stream: released
    version: "16.04"
    virt-type: kvm
  - arch: amd64
    date-created: 1485305700000000000
    image-id: 698a8146-d6d9-4352-99fe-6557ebce5661
    priority: 10
    region: us-east-2
    root-storage-type: ""
    series: xenial
    source: default ubuntu cloud images
    stream: released
    version: "16.04"
    virt-type: kvm
  - arch: amd64
    date-created: 1485305700000000000
    image-id: 698a8146-d6d9-4352-99fe-6557ebce5661
    priority: 10
    region: us-east-1
    root-storage-type: ""
    series: xenial
    source: default ubuntu cloud images
    stream: released
    version: "16.04"
    virt-type: kvm
  - arch: amd64
    date-created: 1485305700000000000
    image-id: 698a8146-d6d9-4352-99fe-6557ebce5661
    priority: 10
    region: us-west-1
    root-storage-type: ""
    series: xenial
    source: default ubuntu cloud images
    stream: released
    version: "16.04"
    virt-type: kvm
  - arch: amd64
    date-created: 1485305700000000000
    image-id: 698a8146-d6d9-4352-99fe-6557ebce5661
    priority: 10
    region: eu-ams-1
    root-storage-type: ""
    series: xenial
    source: default ubuntu cloud images
    stream: released
    version: "16.04"
    virt-type: kvm
  - arch: amd64
    date-created: 1485305700000000000
    image-id: 698a8146-d6d9-4352-99fe-6557ebce5661
    priority: 10
    region: us-sw-1
    root-storage-type: ""
    series: xenial
    source: default ubuntu cloud images
    stream: released
    version: "16.04"
    virt-type: kvm
  version: 1
config:
  agent-metadata-url: ""
  agent-stream: released
  agent-version: 2.1-rc1.1
  apt-ftp-proxy: ""
  apt-http-proxy: http://10.250.171.1:8000
  apt-https-proxy: ""
  apt-mirror: ""
  authorized-keys: |
    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDP9HEBWHyZiPWu4JF0YJ7+H2XH/BaXi7u4kj92z6fJl/LbWvDvWAYEFQdkX+IAHoax+CZQKLYKu9Nx9S328bA1/cBRNu0h7roOsoeUs1fTcvSTa6+KXRw1xMEZ3RMBYPhpI9QPRTMdIF0Mt4OzFjvZt1x8EQZNlnpjhY25H8d+24pkQINLS7ixRcsvqAKERr5e55P+GHf68p2+eXAhQFGNNUErXLsaeUhGlLLiGUUoKlGNmTRn9dC9TKlN4v+woyqk9GdfUhKN5qE9E0VkdprEV88YO+SQOLQjHVDCRIsyVjLRTy45WWGIU8EFg8BDCLLfN8+w8pm6LnsR+p5Z7SpH juju-client-key
    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD055U3fdCpzMpa+B2IhTX0Y8fBsRpo/Z4M6Vv31wDzn0klov3NVVa9uKu3GGS0+Y7DFc0JF5yUVRJ63r8yeSos2ejsV2TAMDnUOPE7fMjqb9JG22izoXurlNeaU61Smwhb/XQB8SB32HiLv7cjG5qey5FFc+VOOsadikwivfnzyKu+uTHOIAFVZMPn1nLcXzdEn4ktj1Gxa01uYchxf00K496pRGDRRE7LrKEj4V/xqhLoWF0XPFQtmSzwIgoJB6YGpF0kj6ZIZTRKFi4qqxIbu1jtF7xAJ8D1ZccdipDdJLOpZy5YJ5ELb4scgOTSBl0LYb9vOW139JkRiFV2ZE77 tim@elwood
  automatically-retry-hooks: true
  default-series: xenial
  development: false
  disable-network-management: false
  enable-os-refresh-update: true
  enable-os-upgrade: false
  firewall-mode: instance
  ftp-proxy: ""
  http-proxy: ""
  https-proxy: ""
  ignore-machine-addresses: false
  image-metadata-url: ""
  image-stream: released
  logforward-enabled: false
  logging-config: <root>=DEBUG;unit=DEBUG
  name: foo
  net-bond-reconfigure-delay: 17
  no-proxy: ""
  provisioner-harvest-mode: destroyed
  proxy-ssh: false
  resource-tags: ""
  ssl-hostname-verification: true
  test-mode: false
  transmit-vendor-metrics: true
  type: lxd
  uuid: bd3fae18-5ea1-4bc5-8837-45400cf1f8f6
filesystems:
  filesystems: []
  version: 1
ip-addresses:
  ip-addresses:
  - config-method: loopback
    device-name: lo
    dns-search-domains: []
    dns-servers: []
    gateway-address: ""
    machine-id: "1"
    subnet-cidr: 127.0.0.0/8
    value: 127.0.0.1
  - config-method: loopback
    device-name: lo
    dns-search-domains: []
    dns-servers: []
    gateway-address: ""
    machine-id: "1"
    subnet-cidr: ::1/128
    value: ::1
  - config-method: static
    device-name: eth0
    dns-search-domains: []
    dns-servers: []
    gateway-address: ""
    machine-id: "1"
    subnet-cidr: 10.250.171.0/24
    value: 10.250.171.49
  version: 1
link-layer-devices:
  link-layer-devices:
  - is-autostart: true
    is-up: true
    mac-address: ""
    machine-id: "1"
    mtu: 65536
    name: lo
    parent-name: ""
    type: loopback
  - is-autostart: true
    is-up: true
    mac-address: 52:c1:28:cb:49:5d
    machine-id: "1"
    mtu: 1500
    name: lxdbr0
    parent-name: ""
    type: bridge
  - is-autostart: true
    is-up: true
    mac-address: 00:16:3e:57:78:a9
    machine-id: "1"
    mtu: 1500
    name: eth0
    parent-name: ""
    type: ethernet
  version: 1
machines:
  machines:
  - block-devices:
      block-devices: []
      version: 1
    containers: []
    id: "1"
    instance:
      architecture: amd64
      instance-id: juju-f1f8f6-1
      status: ""
      version: 1
    jobs:
    - host-units
    machine-addresses:
    - origin: machine
      scope: local-cloud
      type: ipv4
      value: 10.250.171.49
      version: 1
    - origin: machine
      scope: local-machine
      type: ipv4
      value: 127.0.0.1
      version: 1
    - origin: machine
      scope: local-machine
      type: ipv6
      value: ::1
      version: 1
    nonce: machine-0:92103e48-c933-45c3-8ced-c03962b0cf48
    password-hash: bynwWY8+0lUvTpNwjeKI0JUl
    preferred-private-address:
      origin: provider
      scope: local-cloud
      type: ipv4
      value: 10.250.171.49
      version: 1
    preferred-public-address:
      origin: provider
      scope: local-cloud
      type: ipv4
      value: 10.250.171.49
      version: 1
    provider-addresses:
    - origin: provider
      scope: local-cloud
      type: ipv4
      value: 10.250.171.49
      version: 1
    series: xenial
    status:
      status:
        updated: 2017-03-06T01:07:47.466407673Z
        value: started
      version: 1
    status-history:
      history: []
      version: 1
    supported-containers:
    - lxd
    tools:
      sha256: ""
      size: 0
      tools-version: 2.1-rc1.1-xenial-amd64
      url: ""
      version: 1
  version: 1
owner: admin
relations:
  relations: []
  version: 1
sequences:
  application-ubuntu: 2
  machine: 2
spaces:
  spaces: []
  version: 1
ssh-host-keys:
  ssh-host-keys:
  - keys:
    - |
      ssh-dss AAAAB3NzaC1kc3MAAACBAOqfiPafVrcicla2xFFi4Ar72NMRRgOrfwWJ2/WS815bmLk2kUCLEkCnAXt26PqaPNl9yb4ZQvr9TX3HibcufQMtDOrg6OYvTc7VD4mmwi6+ftXgtuf1lIPmoQEMa3mjbMaczNxURSvM12naqJB71SnsqAb2n/kP2YlOOgnlfSVtAAAAFQCNY0qAxGWO8KyLXa7C33dibdIZHwAAAIBeAOSHwvM2PVzGlvzqcAzkZyaCn32zzzy+QcwByuTGsON287NGnWX/0zp+j6rb2dsmA5LBUvTZlT6swjjSGxOwX1QiqkfxAvpMJ0DbHjr5uBEl6KrhMPxUhiaik4UPtF5CLUdaHq0ULJ9ke2LmhfDlTkSAesDbjb9XBFGvZ6sc5gAAAIBrDwBgY2RN842wg304goL3NecR4EUGpjx7aXabT3UvBsvFuMGkAd8AlMXdMbmSwnfly8PQk0mzjnbttvfGV4MKsnRmCov8Rlr1nMKIysFh+X2IYyanFFYGd654P+MWiqG60iC+ZLLf7g6wy2BZd3qd+a0r8Zv+lCDwsUQPGq/jow== root@juju-f1f8f6-1
    - |
      ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBHrX4vn/zOY8DB9dpjxev4ea0EVxpuHeEEtf5yAO1ZMLmgYFpmT5nKUU9lqOFfonZC25jwDAQgmPiHj1C6BVTQ0= root@juju-f1f8f6-1
    - |
      ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEvfmGgBYIX4qNBTYMxSD8XjzwWEGhMgaBSjtQBow7ww root@juju-f1f8f6-1
    - |
      ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDOKwdPMuAmdRXjiCe/RWOx+UGYio/5o7VbPpCz3Ar3kc5BMEUh45clvEO0iw7gqFEN7ZhtaYyG2VzZdF4N7IO+O/FYuooFn7Ng6Bik/iRnbDiyorNgb7mLETMfuNc6eBqwQUph7K4SPX+O9hyiqLK1HgXUwcI7vxTG3m6GUPkM36pbRbP2xfKP4NPBA+KKS/6AimJFC5mOjq6X/JL0nUnbOjQos9nBJEj7RNAtCzVJLXUIlXiLqNV9jBvz1QyKBEhsAcimXwN/XfrMSBWfP5mEgwtIrZILqAvvsjkTF3KfrXqu7SehtgGQS/7NeFDIfP3YQV/mWqNjJTtTvXH+ub0j root@juju-f1f8f6-1
    machine-id: "1"
  version: 1
storage-pools:
  pools: []
  version: 1
storages:
  storages: []
  version: 1
subnets:
  subnets: []
  version: 1
users:
  users:
  - access: admin
    created-by: admin
    date-created: 2017-02-07T02:33:07Z
    display-name: admin
    last-connection: 2017-02-22T00:43:58Z
    name: admin
  version: 1
version: 1
volumes:
  version: 1
  volumes: []
`
